package database

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"

	"gorm.io/gorm"
)

const sqlExecutionTimeout = 30 * time.Second

// 危险操作检测正则
var (
	dropDatabaseRegex  = regexp.MustCompile(`(?i)^\s*DROP\s+DATABASE`)
	dropTableRegex     = regexp.MustCompile(`(?i)^\s*DROP\s+TABLE`)
	deleteWithoutWhere = regexp.MustCompile(`(?i)^\s*DELETE\s+FROM\s+[^\s;]+(\s*;?\s*)$`)
)

// SQLService SQL 执行器服务
type SQLService struct {
	mysqlConnService
	executor *common.CommandExecutor
}

// NewSQLService 创建 SQL 执行器服务
func NewSQLService(db *gorm.DB) *SQLService {
	exec := common.NewCommandExecutor(120 * time.Second)
	exec.Register(&common.CommandTemplate{
		Name: "docker_exec_mysqldump",
		Bin:  "docker",
		Args: []string{
			"exec", "{{name}}",
			"mysqldump", "-h", "127.0.0.1", "-uroot", "-p{{password}}",
			"--databases", "{{dbname}}",
		},
	})
	exec.Register(&common.CommandTemplate{
		Name: "docker_exec_mysqldump_table",
		Bin:  "docker",
		Args: []string{
			"exec", "{{name}}",
			"mysqldump", "-h", "127.0.0.1", "-uroot", "-p{{password}}",
			"{{dbname}}", "{{tablename}}",
		},
	})
	return &SQLService{
		mysqlConnService: mysqlConnService{db: db},
		executor:         exec,
	}
}

// ExecuteSQL 执行 SQL
func (s *SQLService) ExecuteSQL(instanceID uint, dbName, sqlStr string, allowDestructive bool) (*dto.SQLResultResponse, error) {
	sqlStr = strings.TrimSpace(sqlStr)
	if sqlStr == "" {
		return nil, fmt.Errorf("SQL 语句不能为空")
	}

	// 危险操作检测
	if !allowDestructive {
		if err := checkDangerousSQL(sqlStr); err != nil {
			return nil, err
		}
	}

	conn, _, err := s.connectInstance(instanceID, dbName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 设置执行超时
	ctx, cancel := context.WithTimeout(context.Background(), sqlExecutionTimeout)
	defer cancel()

	start := time.Now()
	result := &dto.SQLResultResponse{}

	// 判断是查询还是修改
	if isQuerySQL(sqlStr) {
		rows, err := conn.QueryContext(ctx, sqlStr)
		if err != nil {
			return nil, fmt.Errorf("查询失败: %w", err)
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return nil, err
		}
		result.Columns = columns

		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}
			if err := rows.Scan(valuePtrs...); err != nil {
				return nil, err
			}
			row := make([]interface{}, len(columns))
			for i, v := range values {
				if b, ok := v.([]byte); ok {
					row[i] = string(b)
				} else {
					row[i] = v
				}
			}
			result.Rows = append(result.Rows, row)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	} else {
		// 修改操作
		execResult, err := conn.ExecContext(ctx, sqlStr)
		if err != nil {
			return nil, fmt.Errorf("执行失败: %w", err)
		}
		affected, _ := execResult.RowsAffected()
		result.AffectedRows = affected
	}

	result.ExecutionTimeMs = time.Since(start).Milliseconds()
	return result, nil
}

// checkDangerousSQL 检测危险 SQL
func checkDangerousSQL(sqlStr string) error {
	if dropDatabaseRegex.MatchString(sqlStr) {
		return fmt.Errorf("禁止执行 DROP DATABASE（如需执行请设置 allow_destructive=true）")
	}
	if dropTableRegex.MatchString(sqlStr) {
		return fmt.Errorf("禁止执行 DROP TABLE（如需执行请设置 allow_destructive=true）")
	}
	if deleteWithoutWhere.MatchString(sqlStr) {
		return fmt.Errorf("禁止执行无 WHERE 条件的 DELETE（如需执行请设置 allow_destructive=true）")
	}
	return nil
}

// isQuerySQL 判断是否为查询语句
func isQuerySQL(sqlStr string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sqlStr))
	return strings.HasPrefix(upper, "SELECT") ||
		strings.HasPrefix(upper, "SHOW") ||
		strings.HasPrefix(upper, "DESCRIBE") ||
		strings.HasPrefix(upper, "DESC") ||
		strings.HasPrefix(upper, "EXPLAIN")
}

// ImportSQL 导入 SQL 文件内容
func (s *SQLService) ImportSQL(instanceID uint, dbName, content string) (*dto.ImportResult, error) {
	conn, _, err := s.connectInstance(instanceID, dbName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 按 ; 分割语句
	statements := splitSQLStatements(content)
	result := &dto.ImportResult{
		TotalStatements: len(statements),
	}

	ctx := context.Background()
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := conn.ExecContext(ctx, stmt); err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, fmt.Sprintf("语句失败: %s | 错误: %v", truncateString(stmt, 100), err))
			continue
		}
		result.SuccessCount++
	}
	return result, nil
}

// splitSQLStatements 按分号分割 SQL 语句（简单处理，不处理存储过程中的分号）
func splitSQLStatements(content string) []string {
	var statements []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	inBacktick := false

	for i := 0; i < len(content); i++ {
		ch := content[i]
		switch ch {
		case '\'':
			if !inDoubleQuote && !inBacktick {
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if !inSingleQuote && !inBacktick {
				inDoubleQuote = !inDoubleQuote
			}
		case '`':
			if !inSingleQuote && !inDoubleQuote {
				inBacktick = !inBacktick
			}
		case ';':
			if !inSingleQuote && !inDoubleQuote && !inBacktick {
				statements = append(statements, current.String())
				current.Reset()
				continue
			}
		}
		current.WriteByte(ch)
	}
	if remaining := strings.TrimSpace(current.String()); remaining != "" {
		statements = append(statements, remaining)
	}
	return statements
}

// ExportSQL 导出数据库或表为 SQL 文本
func (s *SQLService) ExportSQL(instanceID uint, dbName, tableName string) (string, error) {
	// 校验标识符，防止 SQL 注入
	if err := validateMySQLIdentifier(dbName); err != nil {
		return "", fmt.Errorf("无效的数据库名: %w", err)
	}
	if tableName != "" {
		if err := validateMySQLIdentifier(tableName); err != nil {
			return "", fmt.Errorf("无效的表名: %w", err)
		}
	}

	instance, err := s.getInstance(instanceID)
	if err != nil {
		return "", err
	}

	containerName := fmt.Sprintf("%s%d", containerNamePrefix(instance.Type), instance.ID)
	params := map[string]string{
		"name":     containerName,
		"password": instance.RootPassword,
		"dbname":   dbName,
	}

	templateName := "docker_exec_mysqldump"
	if tableName != "" {
		templateName = "docker_exec_mysqldump_table"
		params["tablename"] = tableName
	}

	result, err := s.executor.Execute(context.Background(), templateName, params)
	if err != nil || result.ExitCode != 0 {
		// 回退方案：通过 SELECT * 生成 INSERT 语句
		return s.exportFallback(instanceID, dbName, tableName)
	}
	return result.Stdout, nil
}

// exportFallback 回退导出方案：查询数据并生成 SQL
func (s *SQLService) exportFallback(instanceID uint, dbName, tableName string) (string, error) {
	conn, _, err := s.connectInstance(instanceID, dbName)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("-- 犀牛面板 SQL 导出 (回退方案)\n-- 数据库: %s\n\n", dbName))

	tables := []string{tableName}
	if tableName == "" {
		// 导出所有表
		rows, err := conn.Query("SHOW TABLES")
		if err != nil {
			return "", fmt.Errorf("获取表列表失败: %w", err)
		}
		defer rows.Close()
		tables = nil
		for rows.Next() {
			var t string
			rows.Scan(&t)
			tables = append(tables, t)
		}
	}

	for _, tbl := range tables {
		// 校验表名，跳过非法标识符并记录日志，不中断整个导出
		if err := validateMySQLIdentifier(tbl); err != nil {
			slog.Warn("[export] skip invalid table name", "table", tbl, "error", err)
			continue
		}

		// 获取建表语句
		var createTable string
		var tableNameOut string
		row := conn.QueryRow(fmt.Sprintf("SHOW CREATE TABLE `%s`", tbl))
		if err := row.Scan(&tableNameOut, &createTable); err == nil {
			sb.WriteString(fmt.Sprintf("-- 表结构: %s\n", tbl))
			sb.WriteString(fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n", tbl))
			sb.WriteString(createTable + ";\n\n")
		}

		// 导出数据
		dataRows, err := conn.Query(fmt.Sprintf("SELECT * FROM `%s`", tbl))
		if err != nil {
			continue
		}
		columns, _ := dataRows.Columns()
		sb.WriteString(fmt.Sprintf("-- 表数据: %s\n", tbl))
		for dataRows.Next() {
			values := make([]interface{}, len(columns))
			ptrs := make([]interface{}, len(columns))
			for i := range values {
				ptrs[i] = &values[i]
			}
			dataRows.Scan(ptrs...)
			sb.WriteString(fmt.Sprintf("INSERT INTO `%s` (`%s`) VALUES (", tbl, strings.Join(columns, "`,`")))
			for i, v := range values {
				if i > 0 {
					sb.WriteString(",")
				}
				if v == nil {
					sb.WriteString("NULL")
				} else if b, ok := v.([]byte); ok {
					sb.WriteString(fmt.Sprintf("'%s'", escapeSQLString(string(b))))
				} else {
					sb.WriteString(fmt.Sprintf("'%v'", escapeSQLString(fmt.Sprintf("%v", v))))
				}
			}
			sb.WriteString(");\n")
		}
		dataRows.Close()
		sb.WriteString("\n")
	}
	return sb.String(), nil
}

// escapeSQLString 转义 SQL 字符串
func escapeSQLString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
