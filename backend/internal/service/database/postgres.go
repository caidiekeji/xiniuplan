package database

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"cloudpivot/internal/dto"
	"cloudpivot/internal/model"

	"github.com/jackc/pgx/v5"
	"gorm.io/gorm"
)

// postgresSystemDatabases PostgreSQL 系统库过滤
var postgresSystemDatabases = map[string]bool{
	"postgres":  true,
	"template0": true,
	"template1": true,
}

// pgIdentifierRegex 合法标识符校验
var pgIdentifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_$]*$`)

// pgEngine PostgreSQL 引擎，实现 DatabaseEngine 接口
type pgEngine struct {
	db       *gorm.DB
	instance *model.DatabaseInstance
}

// newPgEngine 创建 PostgreSQL 引擎
func newPgEngine(db *gorm.DB, instance *model.DatabaseInstance) *pgEngine {
	return &pgEngine{db: db, instance: instance}
}

// connString 构建连接字符串
func (e *pgEngine) connString(dbName string) string {
	username := e.instance.Username
	if username == "" {
		username = "postgres"
	}
	host := e.instance.Host
	if host == "" {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		username, e.instance.RootPassword, host, e.instance.Port, dbName)
}

// connect 连接到指定数据库
func (e *pgEngine) connect(ctx context.Context, dbName string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, e.connString(dbName))
	if err != nil {
		return nil, fmt.Errorf("连接 PostgreSQL 失败: %w", err)
	}
	return conn, nil
}

// validateIdentifier 校验标识符合法性
func validateIdentifier(name string) error {
	if !pgIdentifierRegex.MatchString(name) {
		return fmt.Errorf("无效的标识符: %s", name)
	}
	return nil
}

// quoteIdentifier 引用标识符
func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// splitTableRef 拆分 schema.table 格式
func splitTableRef(table string) (schema, tbl string) {
	if idx := strings.Index(table, "."); idx > 0 {
		return table[:idx], table[idx+1:]
	}
	return "public", table
}

// convertPGValues 将 pgx 行值转换为 JSON 友好类型
func convertPGValues(values []interface{}) []interface{} {
	result := make([]interface{}, len(values))
	for i, v := range values {
		result[i] = convertPGValue(v)
	}
	return result
}

func convertPGValue(v interface{}) interface{} {
	switch val := v.(type) {
	case []byte:
		return string(val)
	case time.Time:
		return val.Format(time.RFC3339Nano)
	case nil:
		return nil
	default:
		return v
	}
}

// isPGQuerySQL 判断是否为查询语句
func isPGQuerySQL(sqlStr string) bool {
	upper := strings.ToUpper(strings.TrimSpace(sqlStr))
	return strings.HasPrefix(upper, "SELECT") ||
		strings.HasPrefix(upper, "WITH") ||
		strings.HasPrefix(upper, "SHOW") ||
		strings.HasPrefix(upper, "TABLE") ||
		strings.HasPrefix(upper, "VALUES") ||
		strings.HasPrefix(upper, "EXPLAIN")
}

// ========== DatabaseEngine 接口实现 ==========

func (e *pgEngine) ListDatabases() ([]string, error) {
	ctx := context.Background()
	conn, err := e.connect(ctx, "postgres")
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname")
	if err != nil {
		return nil, fmt.Errorf("查询数据库列表失败: %w", err)
	}
	defer rows.Close()

	dbs := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		if !postgresSystemDatabases[name] {
			dbs = append(dbs, name)
		}
	}
	return dbs, rows.Err()
}

func (e *pgEngine) CreateDatabase(name string) error {
	if err := validateIdentifier(name); err != nil {
		return err
	}

	ctx := context.Background()
	conn, err := e.connect(ctx, "postgres")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	sql := fmt.Sprintf("CREATE DATABASE %s ENCODING 'UTF8'", quoteIdentifier(name))
	if _, err := conn.Exec(ctx, sql); err != nil {
		return fmt.Errorf("创建数据库失败: %w", err)
	}
	return nil
}

// CreateDatabaseEx 创建数据库（支持编码、locale、owner）
func (e *pgEngine) CreateDatabaseEx(name, encoding, locale, owner string) error {
	if err := validateIdentifier(name); err != nil {
		return err
	}
	if encoding == "" {
		encoding = "UTF8"
	}

	ctx := context.Background()
	conn, err := e.connect(ctx, "postgres")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE DATABASE %s ENCODING '%s'", quoteIdentifier(name), encoding))
	if locale != "" {
		sb.WriteString(fmt.Sprintf(" LC_COLLATE '%s' LC_CTYPE '%s'", locale, locale))
	}
	if owner != "" {
		if err := validateIdentifier(owner); err != nil {
			return err
		}
		sb.WriteString(fmt.Sprintf(" OWNER %s", quoteIdentifier(owner)))
	}

	if _, err := conn.Exec(ctx, sb.String()); err != nil {
		return fmt.Errorf("创建数据库失败: %w", err)
	}
	return nil
}

func (e *pgEngine) DropDatabase(name string) error {
	if postgresSystemDatabases[name] {
		return fmt.Errorf("不能删除系统数据库: %s", name)
	}
	if err := validateIdentifier(name); err != nil {
		return err
	}

	ctx := context.Background()
	conn, err := e.connect(ctx, "postgres")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	// 终止所有连接后删除
	conn.Exec(ctx, fmt.Sprintf(
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '%s' AND pid <> pg_backend_pid()", name))

	sql := fmt.Sprintf("DROP DATABASE %s", quoteIdentifier(name))
	if _, err := conn.Exec(ctx, sql); err != nil {
		return fmt.Errorf("删除数据库失败: %w", err)
	}
	return nil
}

func (e *pgEngine) ListTables(dbName string) ([]TableInfo, error) {
	if err := validateIdentifier(dbName); err != nil {
		return nil, err
	}

	ctx := context.Background()
	conn, err := e.connect(ctx, dbName)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema')
		ORDER BY table_name
	`)
	if err != nil {
		return nil, fmt.Errorf("查询表列表失败: %w", err)
	}
	defer rows.Close()

	tables := make([]TableInfo, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, TableInfo{Name: name})
	}
	return tables, rows.Err()
}

// ListTablesEx 列出表（含 schema、类型、估计行数）
func (e *pgEngine) ListTablesEx(dbName string) ([]dto.PostgresTableInfo, error) {
	if err := validateIdentifier(dbName); err != nil {
		return nil, err
	}

	ctx := context.Background()
	conn, err := e.connect(ctx, dbName)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, `
		SELECT t.table_name, t.table_schema, t.table_type,
		       COALESCE(c.reltuples::bigint, 0) AS row_estimate
		FROM information_schema.tables t
		LEFT JOIN pg_class c ON c.relname = t.table_name
		LEFT JOIN pg_namespace n ON n.oid = c.relnamespace AND n.nspname = t.table_schema
		WHERE t.table_schema NOT IN ('pg_catalog', 'information_schema')
		ORDER BY t.table_name
	`)
	if err != nil {
		return nil, fmt.Errorf("查询表列表失败: %w", err)
	}
	defer rows.Close()

	tables := make([]dto.PostgresTableInfo, 0)
	for rows.Next() {
		var t dto.PostgresTableInfo
		if err := rows.Scan(&t.Name, &t.Schema, &t.Type, &t.RowEstimate); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (e *pgEngine) GetTableSchema(dbName, table string) (TableSchema, error) {
	if err := validateIdentifier(dbName); err != nil {
		return TableSchema{}, err
	}
	schema, tbl := splitTableRef(table)
	if err := validateIdentifier(schema); err != nil {
		return TableSchema{}, err
	}
	if err := validateIdentifier(tbl); err != nil {
		return TableSchema{}, err
	}

	ctx := context.Background()
	conn, err := e.connect(ctx, dbName)
	if err != nil {
		return TableSchema{}, err
	}
	defer conn.Close(ctx)

	result := TableSchema{}

	// 列信息
	colRows, err := conn.Query(ctx, `
		SELECT column_name, data_type, is_nullable, column_default,
		       CASE WHEN character_maximum_length IS NOT NULL
		            THEN data_type || '(' || character_maximum_length || ')'
		            WHEN numeric_precision IS NOT NULL AND numeric_scale IS NOT NULL
		            THEN data_type || '(' || numeric_precision || ',' || numeric_scale || ')'
		            ELSE data_type
		       END AS full_type,
		       CASE WHEN is_identity = 'YES' THEN 'auto_increment' ELSE '' END AS extra
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
		ORDER BY ordinal_position
	`, schema, tbl)
	if err != nil {
		return TableSchema{}, fmt.Errorf("获取表字段失败: %w", err)
	}
	for colRows.Next() {
		var col TableColumn
		var defaultVal *string
		// full_type（第 5 列）是带长度/精度的完整类型，会覆盖裸 data_type，
		// 因此必须用独立变量接收，不能和 col.Type 共用 —— 否则两次 Scan
		// 会互相覆盖，最终 col.Type 变成裸类型（如 varchar 而非 varchar(64)）
		var fullType string
		if err := colRows.Scan(&col.Field, &col.Type, &col.Null, &defaultVal, &fullType, &col.Extra); err != nil {
			colRows.Close()
			return TableSchema{}, err
		}
		if fullType != "" {
			col.Type = fullType
		}
		col.Key = ""
		if defaultVal != nil {
			col.Default = *defaultVal
		}
		result.Columns = append(result.Columns, col)
	}
	colRows.Close()

	// 主键标记
	pkRows, _ := conn.Query(ctx, `
		SELECT kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
		WHERE tc.table_schema = $1 AND tc.table_name = $2 AND tc.constraint_type = 'PRIMARY KEY'
	`, schema, tbl)
	pkCols := make(map[string]bool)
	for pkRows.Next() {
		var col string
		pkRows.Scan(&col)
		pkCols[col] = true
	}
	pkRows.Close()
	for i := range result.Columns {
		if pkCols[result.Columns[i].Field] {
			result.Columns[i].Key = "PRI"
		}
	}

	// 索引
	idxRows, err := conn.Query(ctx, `
		SELECT indexname, indexdef FROM pg_indexes
		WHERE schemaname = $1 AND tablename = $2
		ORDER BY indexname
	`, schema, tbl)
	if err == nil {
		for idxRows.Next() {
			var name, def string
			idxRows.Scan(&name, &def)
			// 解析索引类型和列
			idxType := "BTREE"
			if strings.Contains(strings.ToUpper(def), "HASH") {
				idxType = "HASH"
			} else if strings.Contains(strings.ToUpper(def), "GIN") {
				idxType = "GIN"
			} else if strings.Contains(strings.ToUpper(def), "GIST") {
				idxType = "GIST"
			}
			// 提取列名
			colName := ""
			if start := strings.Index(def, "("); start > 0 {
				if end := strings.Index(def[start:], ")"); end > 0 {
					colName = def[start+1 : start+end]
				}
			}
			nonUnique := !strings.Contains(strings.ToUpper(def), "UNIQUE INDEX") &&
				!strings.Contains(strings.ToUpper(def), "PRIMARY KEY")
			result.Indexes = append(result.Indexes, TableIndex{
				KeyName:    name,
				ColumnName: colName,
				NonUnique:  nonUnique,
				IndexType:  idxType,
				SeqInIndex: 1,
			})
		}
		idxRows.Close()
	}

	// 约束
	constraintRows, err := conn.Query(ctx, `
		SELECT tc.constraint_name, tc.constraint_type,
		       string_agg(kcu.column_name, ',' ORDER BY kcu.ordinal_position) AS columns,
		       ccu.table_name AS ref_table,
		       string_agg(ccu.column_name, ',' ORDER BY ccu.column_name) AS ref_columns
		FROM information_schema.table_constraints tc
		LEFT JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name AND tc.table_schema = kcu.table_schema
		LEFT JOIN information_schema.constraint_column_usage ccu ON tc.constraint_name = ccu.constraint_name
		WHERE tc.table_schema = $1 AND tc.table_name = $2
		GROUP BY tc.constraint_name, tc.constraint_type, ccu.table_name
		ORDER BY tc.constraint_type, tc.constraint_name
	`, schema, tbl)
	if err == nil {
		for constraintRows.Next() {
			var c TableConstraint
			var refTable, refCols *string
			if err := constraintRows.Scan(&c.Name, &c.Type, &c.Columns, &refTable, &refCols); err != nil {
				continue
			}
			if refTable != nil {
				c.RefTable = *refTable
			}
			if refCols != nil {
				c.RefCols = *refCols
			}
			result.Constraints = append(result.Constraints, c)
		}
		constraintRows.Close()
	}

	return result, nil
}

func (e *pgEngine) QueryTableData(dbName, table string, page, pageSize int) (*QueryResult, error) {
	if err := validateIdentifier(dbName); err != nil {
		return nil, err
	}
	schema, tbl := splitTableRef(table)
	if err := validateIdentifier(schema); err != nil {
		return nil, err
	}
	if err := validateIdentifier(tbl); err != nil {
		return nil, err
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 500 {
		pageSize = 20
	}

	ctx := context.Background()
	conn, err := e.connect(ctx, dbName)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	fullTable := quoteIdentifier(schema) + "." + quoteIdentifier(tbl)

	// 总数
	var total int64
	if err := conn.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", fullTable)).Scan(&total); err != nil {
		return nil, fmt.Errorf("查询总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	rows, err := conn.Query(ctx, fmt.Sprintf("SELECT * FROM %s LIMIT $1 OFFSET $2", fullTable), pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("查询表数据失败: %w", err)
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescs))
	for i, fd := range fieldDescs {
		columns[i] = fd.Name
	}

	var dataRows [][]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		dataRows = append(dataRows, convertPGValues(values))
	}

	return &QueryResult{
		Columns:  columns,
		Rows:     dataRows,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, rows.Err()
}

func (e *pgEngine) ExecuteSQL(dbName, sqlStr string, allowDestructive bool) (*SQLResult, error) {
	sqlStr = strings.TrimSpace(sqlStr)
	if sqlStr == "" {
		return nil, fmt.Errorf("SQL 语句不能为空")
	}

	// 危险操作检测（复用通用检测）
	if !allowDestructive {
		if err := checkDangerousSQL(sqlStr); err != nil {
			return nil, err
		}
	}

	if err := validateIdentifier(dbName); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), sqlExecutionTimeout)
	defer cancel()

	conn, err := e.connect(ctx, dbName)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	start := time.Now()
	result := &SQLResult{}

	// 使用 Query 执行，通过 FieldDescriptions 判断是否有结果集
	rows, err := conn.Query(ctx, sqlStr)
	if err != nil {
		return nil, fmt.Errorf("执行失败: %w", err)
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	if len(fieldDescs) > 0 {
		// 有结果集
		columns := make([]string, len(fieldDescs))
		for i, fd := range fieldDescs {
			columns[i] = fd.Name
		}
		result.Columns = columns

		for rows.Next() {
			values, err := rows.Values()
			if err != nil {
				return nil, err
			}
			result.Rows = append(result.Rows, convertPGValues(values))
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	} else {
		// 命令操作
		tag := rows.CommandTag()
		result.AffectedRows = tag.RowsAffected()
	}

	result.ExecutionTimeMs = time.Since(start).Milliseconds()
	return result, nil
}

func (e *pgEngine) ListUsers() ([]UserInfo, error) {
	ctx := context.Background()
	conn, err := e.connect(ctx, "postgres")
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, "SELECT rolname FROM pg_roles WHERE rolcanlogin = true ORDER BY rolname")
	if err != nil {
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}
	users := make([]UserInfo, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		users = append(users, UserInfo{User: name, Host: "-"})
	}
	return users, rows.Err()
}

// ListUsersEx 列出用户（含详细属性）
func (e *pgEngine) ListUsersEx() ([]dto.PostgresUserInfo, error) {
	ctx := context.Background()
	conn, err := e.connect(ctx, "postgres")
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, `
		SELECT rolname, rolsuper, rolcreatedb, rolcreaterole, rolcanlogin, rolreplication,
		       COALESCE(rolvaliduntil::text, '') AS valid_until
		FROM pg_roles ORDER BY rolname
	`)
	if err != nil {
		return nil, fmt.Errorf("查询用户列表失败: %w", err)
	}
	users := make([]dto.PostgresUserInfo, 0)
	for rows.Next() {
		var u dto.PostgresUserInfo
		if err := rows.Scan(&u.Username, &u.Superuser, &u.CreateDB, &u.CreateRole,
			&u.Login, &u.Replication, &u.ValidUntil); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (e *pgEngine) CreateUser(username, password, host string) error {
	if err := validateIdentifier(username); err != nil {
		return err
	}

	ctx := context.Background()
	conn, err := e.connect(ctx, "postgres")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	sql := fmt.Sprintf("CREATE ROLE %s WITH LOGIN PASSWORD '%s'",
		quoteIdentifier(username), strings.ReplaceAll(password, "'", "''"))
	if _, err := conn.Exec(ctx, sql); err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}
	return nil
}

// CreateUserEx 创建用户（含角色属性）
func (e *pgEngine) CreateUserEx(req dto.CreatePostgresUserRequest) error {
	if err := validateIdentifier(req.Username); err != nil {
		return err
	}

	ctx := context.Background()
	conn, err := e.connect(ctx, "postgres")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	var attrs []string
	if req.Login {
		attrs = append(attrs, "LOGIN")
	} else {
		attrs = append(attrs, "NOLOGIN")
	}
	if req.Superuser {
		attrs = append(attrs, "SUPERUSER")
	} else {
		attrs = append(attrs, "NOSUPERUSER")
	}
	if req.CreateDB {
		attrs = append(attrs, "CREATEDB")
	}
	if req.CreateRole {
		attrs = append(attrs, "CREATEROLE")
	}
	if req.Replication {
		attrs = append(attrs, "REPLICATION")
	}

	sql := fmt.Sprintf("CREATE ROLE %s WITH %s PASSWORD '%s'",
		quoteIdentifier(req.Username),
		strings.Join(attrs, " "),
		strings.ReplaceAll(req.Password, "'", "''"))

	if _, err := conn.Exec(ctx, sql); err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}
	return nil
}

func (e *pgEngine) DropUser(username, host string) error {
	if err := validateIdentifier(username); err != nil {
		return err
	}

	ctx := context.Background()
	conn, err := e.connect(ctx, "postgres")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	sql := fmt.Sprintf("DROP ROLE %s", quoteIdentifier(username))
	if _, err := conn.Exec(ctx, sql); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}
	return nil
}

func (e *pgEngine) GrantPrivileges(username, host, dbName string, privileges []string) error {
	if err := validateIdentifier(username); err != nil {
		return err
	}
	if err := validateIdentifier(dbName); err != nil {
		return err
	}

	// 权限白名单
	allowed := map[string]bool{
		"ALL":        true,
		"CONNECT":    true,
		"CREATE":     true,
		"TEMPORARY":  true,
		"TEMP":       true,
		"SELECT":     true,
		"INSERT":     true,
		"UPDATE":     true,
		"DELETE":     true,
		"TRUNCATE":   true,
		"REFERENCES": true,
		"TRIGGER":    true,
		"USAGE":      true,
	}
	privList := make([]string, 0, len(privileges))
	for _, p := range privileges {
		upper := strings.ToUpper(strings.TrimSpace(p))
		if !allowed[upper] {
			return fmt.Errorf("不支持的权限: %s", p)
		}
		privList = append(privList, upper)
	}

	ctx := context.Background()
	conn, err := e.connect(ctx, "postgres")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	privStr := strings.Join(privList, ", ")
	sql := fmt.Sprintf("GRANT %s ON DATABASE %s TO %s",
		privStr, quoteIdentifier(dbName), quoteIdentifier(username))
	if _, err := conn.Exec(ctx, sql); err != nil {
		return fmt.Errorf("授权失败: %w", err)
	}

	// 同时授予 schema 上的 USAGE 权限和表上的默认权限
	conn.Exec(ctx, fmt.Sprintf("GRANT USAGE ON SCHEMA public TO %s", quoteIdentifier(username)))
	conn.Exec(ctx, fmt.Sprintf("GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO %s", quoteIdentifier(username)))
	conn.Exec(ctx, fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO %s", quoteIdentifier(username)))

	return nil
}

func (e *pgEngine) RevokePrivileges(username, host, dbName string) error {
	if err := validateIdentifier(username); err != nil {
		return err
	}
	if err := validateIdentifier(dbName); err != nil {
		return err
	}

	ctx := context.Background()
	conn, err := e.connect(ctx, "postgres")
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	sql := fmt.Sprintf("REVOKE ALL PRIVILEGES ON DATABASE %s FROM %s",
		quoteIdentifier(dbName), quoteIdentifier(username))
	if _, err := conn.Exec(ctx, sql); err != nil {
		return fmt.Errorf("撤销权限失败: %w", err)
	}
	return nil
}

// ========== PostgresService 实例级包装 ==========

// PostgresService PostgreSQL 管理服务（实例级）
type PostgresService struct {
	db *gorm.DB
}

// NewPostgresService 创建 PostgreSQL 管理服务
func NewPostgresService(db *gorm.DB) *PostgresService {
	return &PostgresService{db: db}
}

// GetEngine 获取指定实例的 PostgreSQL 引擎
func (s *PostgresService) GetEngine(instanceID uint) (*pgEngine, error) {
	var instance model.DatabaseInstance
	if err := s.db.First(&instance, instanceID).Error; err != nil {
		return nil, fmt.Errorf("实例不存在: %d", instanceID)
	}
	if instance.Type != "postgres" {
		return nil, NewUnsupportedTypeError(instance.Type, "该操作仅支持 PostgreSQL 实例")
	}
	return newPgEngine(s.db, &instance), nil
}
