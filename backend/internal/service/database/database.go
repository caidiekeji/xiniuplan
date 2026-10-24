package database

import (
	"database/sql"
	"fmt"

	"cloudpivot/internal/dto"

	"gorm.io/gorm"
)

// DatabaseService 库表管理服务
type DatabaseService struct {
	mysqlConnService
}

// NewDatabaseService 创建库表管理服务
func NewDatabaseService(db *gorm.DB) *DatabaseService {
	return &DatabaseService{
		mysqlConnService: mysqlConnService{db: db},
	}
}

// ListDatabases 列出数据库（过滤系统库）
func (s *DatabaseService) ListDatabases(instanceID uint) ([]dto.DatabaseInfo, error) {
	conn, _, err := s.connectInstance(instanceID, "")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query("SHOW DATABASES")
	if err != nil {
		return nil, fmt.Errorf("查询数据库列表失败: %w", err)
	}
	defer rows.Close()

	dbs := make([]dto.DatabaseInfo, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		if !systemDatabases[name] {
			dbs = append(dbs, dto.DatabaseInfo{Name: name})
		}
	}
	return dbs, rows.Err()
}

// CreateDatabase 创建数据库
func (s *DatabaseService) CreateDatabase(instanceID uint, name, charset, collation string) error {
	if err := validateMySQLIdentifier(name); err != nil {
		return err
	}
	if charset == "" {
		charset = "utf8mb4"
	}
	if collation == "" {
		collation = "utf8mb4_unicode_ci"
	}
	// 校验 charset 和 collation 防止注入
	if !charsetRegex.MatchString(charset) {
		return fmt.Errorf("无效的字符集: %s", charset)
	}
	if !collationRegex.MatchString(collation) {
		return fmt.Errorf("无效的排序规则: %s", collation)
	}

	conn, _, err := s.connectInstance(instanceID, "")
	if err != nil {
		return err
	}
	defer conn.Close()

	query := fmt.Sprintf("CREATE DATABASE `%s` CHARACTER SET %s COLLATE %s", name, charset, collation)
	if _, err := conn.Exec(query); err != nil {
		return fmt.Errorf("创建数据库失败: %w", err)
	}
	return nil
}

// DeleteDatabase 删除数据库
func (s *DatabaseService) DeleteDatabase(instanceID uint, name string) error {
	if err := validateMySQLIdentifier(name); err != nil {
		return err
	}
	if systemDatabases[name] {
		return fmt.Errorf("不能删除系统数据库: %s", name)
	}

	conn, _, err := s.connectInstance(instanceID, "")
	if err != nil {
		return err
	}
	defer conn.Close()

	query := fmt.Sprintf("DROP DATABASE `%s`", name)
	if _, err := conn.Exec(query); err != nil {
		return fmt.Errorf("删除数据库失败: %w", err)
	}
	return nil
}

// ListTables 列出指定库的所有表
func (s *DatabaseService) ListTables(instanceID uint, dbName string) ([]dto.TableInfo, error) {
	conn, _, err := s.connectInstance(instanceID, dbName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query("SHOW TABLES")
	if err != nil {
		return nil, fmt.Errorf("查询表列表失败: %w", err)
	}
	defer rows.Close()

	tables := make([]dto.TableInfo, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, dto.TableInfo{Name: name})
	}
	return tables, rows.Err()
}

// GetTableStructure 获取表结构（字段 + 索引）
func (s *DatabaseService) GetTableStructure(instanceID uint, dbName, tableName string) (*dto.TableStructure, error) {
	if err := validateMySQLIdentifier(tableName); err != nil {
		return nil, err
	}
	conn, _, err := s.connectInstance(instanceID, dbName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	result := &dto.TableStructure{}

	// DESCRIBE
	descRows, err := conn.Query(fmt.Sprintf("DESCRIBE `%s`", tableName))
	if err != nil {
		return nil, fmt.Errorf("获取表字段失败: %w", err)
	}
	defer descRows.Close()

	// DESCRIBE 的列数理论上也可能随版本变化，按实际列数动态 Scan
	descCols, err := descRows.Columns()
	if err != nil {
		return nil, err
	}

	for descRows.Next() {
		var col dto.TableColumn
		var defaultVal sql.NullString
		targets := []interface{}{
			&col.Field, &col.Type, &col.Null, &col.Key, &defaultVal, &col.Extra,
		}
		for len(targets) < len(descCols) {
			targets = append(targets, new(sql.NullString))
		}
		if err := descRows.Scan(targets...); err != nil {
			return nil, err
		}
		if defaultVal.Valid {
			col.Default = defaultVal.String
		}
		result.Columns = append(result.Columns, col)
	}
	if err := descRows.Err(); err != nil {
		return nil, err
	}

	// SHOW INDEX
	idxRows, err := conn.Query(fmt.Sprintf("SHOW INDEX FROM `%s`", tableName))
	if err != nil {
		return nil, fmt.Errorf("获取表索引失败: %w", err)
	}
	defer idxRows.Close()

	// SHOW INDEX 的列数随 MySQL 版本变化：5.7 是 13 列，8.0 增加
	// Visible、Expression 两列共 15 列。写死列数会让 8.0 上报
	// 「expected 15 destination arguments in Scan, not 13」，
	// 导致「表结构」页面 100% 报错。这里按实际列数动态 Scan。
	idxCols, err := idxRows.Columns()
	if err != nil {
		return nil, err
	}

	for idxRows.Next() {
		var idx dto.TableIndex
		var subPart, packed sql.NullString
		targets := buildIndexScanTargets(&idx, &subPart, &packed, len(idxCols))
		if err := idxRows.Scan(targets...); err != nil {
			return nil, err
		}
		if subPart.Valid {
			idx.SubPart = subPart.String
		}
		if packed.Valid {
			idx.Packed = packed.String
		}
		result.Indexes = append(result.Indexes, idx)
	}
	return result, idxRows.Err()
}

// buildIndexScanTargets 生成 SHOW INDEX 的 Scan 目标列表。
//
// SHOW INDEX 的列数随 MySQL 版本变化：5.7 是 13 列，8.0 增加 Visible、
// Expression 两列共 15 列。写死列数会让 8.0 上报
// 「expected 15 destination arguments in Scan, not 13」，
// 导致「表结构」页面 100% 报错。这里按实际列数动态补齐目标。
func buildIndexScanTargets(idx *dto.TableIndex, subPart, packed *sql.NullString, colCount int) []interface{} {
	// 前 13 列为各版本共有，顺序固定
	targets := []interface{}{
		&idx.Table, &idx.NonUnique, &idx.KeyName, &idx.SeqInIndex,
		&idx.ColumnName, &idx.Collation, &idx.Cardinality, subPart,
		packed, &idx.Null, &idx.IndexType, &idx.Comment, &idx.IndexComment,
	}
	// 版本新增列用通用类型接收后忽略
	for len(targets) < colCount {
		targets = append(targets, new(sql.NullString))
	}
	return targets
}

// ListTableData 分页查询表数据
func (s *DatabaseService) ListTableData(instanceID uint, dbName, tableName string, page, pageSize int) (*dto.TableData, error) {
	if err := validateMySQLIdentifier(tableName); err != nil {
		return nil, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 500 {
		pageSize = 20
	}

	conn, _, err := s.connectInstance(instanceID, dbName)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// 获取总数
	var total int64
	countRow := conn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName))
	if err := countRow.Scan(&total); err != nil {
		return nil, fmt.Errorf("查询总数失败: %w", err)
	}

	// 分页查询（参数化 LIMIT）
	offset := (page - 1) * pageSize
	rows, err := conn.Query(
		fmt.Sprintf("SELECT * FROM `%s` LIMIT ? OFFSET ?", tableName),
		pageSize, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("查询表数据失败: %w", err)
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// 读取行数据
	var dataRows [][]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		// 转换 []byte 为 string
		row := make([]interface{}, len(columns))
		for i, v := range values {
			if b, ok := v.([]byte); ok {
				row[i] = string(b)
			} else {
				row[i] = v
			}
		}
		dataRows = append(dataRows, row)
	}

	return &dto.TableData{
		Columns:  columns,
		Rows:     dataRows,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, rows.Err()
}
