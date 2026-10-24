package dto

// CreateDBInstanceRequest 创建数据库实例请求
type CreateDBInstanceRequest struct {
	Type         string `json:"type" binding:"required"` // mysql | postgres | redis
	Version      string `json:"version"`                 // 5.7 / 8.0
	Port         int    `json:"port"`                    // 0=自动分配
	RootPassword string `json:"root_password"`           // 空=随机生成
	Remark       string `json:"remark"`
}

// ImportDBInstanceRequest 导入已有数据库实例请求（不创建Docker容器）
type ImportDBInstanceRequest struct {
	Type     string `json:"type" binding:"required"` // mysql | postgres | redis
	Host     string `json:"host" binding:"required"` // 数据库地址
	Port     int    `json:"port" binding:"required"` // 端口
	Username string `json:"username"`                // 管理员用户名
	Password string `json:"password"`                // 管理员密码
	Version  string `json:"version"`                 // 版本（可选，自动探测）
	Remark   string `json:"remark"`
}

// DiscoveredDB 自动发现的数据库
type DiscoveredDB struct {
	Type    string `json:"type"`    // mysql | postgres | redis
	Host    string `json:"host"`    // 地址
	Port    int    `json:"port"`    // 端口
	Version string `json:"version"` // 探测到的版本
	Running bool   `json:"running"` // 是否运行中
}

// DBInstanceResponse 数据库实例响应
type DBInstanceResponse struct {
	ID              uint   `json:"id"`
	Type            string `json:"type"`
	Version         string `json:"version"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Username        string `json:"username"`
	ContainerID     string `json:"container_id"`
	IsExternal      bool   `json:"is_external"`
	Status          string `json:"status"`
	ContainerStatus string `json:"container_status"`
	Remark          string `json:"remark"`
	CreatedAt       string `json:"created_at"`
}

// CreateDatabaseRequest 创建数据库请求
type CreateDatabaseRequest struct {
	Name      string `json:"name" binding:"required"`
	Charset   string `json:"charset"`   // 默认 utf8mb4
	Collation string `json:"collation"` // 默认 utf8mb4_unicode_ci
}

// DatabaseInfo 数据库信息
type DatabaseInfo struct {
	Name string `json:"name"`
}

// TableInfo 表信息
type TableInfo struct {
	Name string `json:"name"`
}

// TableStructure 表结构
type TableStructure struct {
	Columns []TableColumn `json:"columns"`
	Indexes []TableIndex  `json:"indexes"`
}

// TableColumn 表字段
type TableColumn struct {
	Field   string      `json:"field"`
	Type    string      `json:"type"`
	Null    string      `json:"null"`
	Key     string      `json:"key"`
	Default interface{} `json:"default"`
	Extra   string      `json:"extra"`
}

// TableIndex 表索引
type TableIndex struct {
	Table        string      `json:"table"`
	NonUnique    int         `json:"non_unique"`
	KeyName      string      `json:"key_name"`
	SeqInIndex   int         `json:"seq_in_index"`
	ColumnName   string      `json:"column_name"`
	Collation    string      `json:"collation"`
	Cardinality  int64       `json:"cardinality"`
	SubPart      interface{} `json:"sub_part"`
	Packed       interface{} `json:"packed"`
	Null         string      `json:"null"`
	IndexType    string      `json:"index_type"`
	Comment      string      `json:"comment"`
	IndexComment string      `json:"index_comment"`
}

// TableData 表数据分页
type TableData struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Total   int64           `json:"total"`
	Page    int             `json:"page"`
	// 与 common.PageData 保持一致的 camelCase（此前用 page_size，与其他分页接口不一致）
	PageSize int `json:"pageSize"`
}

// CreateDatabaseUserRequest 创建数据库用户请求
type CreateDatabaseUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Host     string `json:"host"` // 默认 %
}

// UserInfo 用户信息
type UserInfo struct {
	User string `json:"username"`
	Host string `json:"host"`
}

// GrantRequest 授权请求
type GrantRequest struct {
	// Username 来自 URL path（handler 会用 path 值覆盖），不应在 body 里重复必填，
	// 否则前端不重复传 username 就会 400，授权按钮 100% 失败（2026-10-23 闭环审计）
	Username   string   `json:"username"`
	Host       string   `json:"host"`
	DBName     string   `json:"db_name" binding:"required"`
	Privileges []string `json:"privileges" binding:"required"`
}

// ExecuteSQLRequest 执行 SQL 请求
type ExecuteSQLRequest struct {
	DBName           string `json:"db_name"`
	SQL              string `json:"sql" binding:"required"`
	AllowDestructive bool   `json:"allow_destructive"`
}

// SQLResultResponse SQL 执行结果
type SQLResultResponse struct {
	Columns         []string        `json:"columns"`
	Rows            [][]interface{} `json:"rows"`
	AffectedRows    int64           `json:"affected_rows"`
	ExecutionTimeMs int64           `json:"execution_time_ms"`
}

// ImportRequest 导入 SQL 请求
type ImportRequest struct {
	DBName  string `json:"db_name"`
	Content string `json:"content" binding:"required"`
}

// ImportResult 导入结果
type ImportResult struct {
	TotalStatements int      `json:"total_statements"`
	SuccessCount    int      `json:"success_count"`
	FailedCount     int      `json:"failed_count"`
	Errors          []string `json:"errors"`
}
