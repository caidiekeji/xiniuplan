package dto

// CreatePostgresDatabaseRequest 创建 PostgreSQL 数据库请求
type CreatePostgresDatabaseRequest struct {
	Name     string `json:"name" binding:"required"`
	Encoding string `json:"encoding"` // 默认 UTF8
	Locale   string `json:"locale"`   // 默认 en_US.UTF8
	Owner    string `json:"owner"`    // 默认 postgres
}

// CreatePostgresUserRequest 创建 PostgreSQL 用户请求
type CreatePostgresUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	// 角色属性
	Superuser   bool `json:"superuser"`
	CreateDB    bool `json:"createdb"`
	CreateRole  bool `json:"createrole"`
	Login       bool `json:"login"` // 默认 true
	Replication bool `json:"replication"`
}

// PostgresGrantRequest PostgreSQL 授权请求
type PostgresGrantRequest struct {
	Privileges []string `json:"privileges" binding:"required"` // ALL | SELECT | INSERT | UPDATE | DELETE | CONNECT | TEMPORARY
}

// PostgresUserInfo PostgreSQL 用户信息
type PostgresUserInfo struct {
	Username    string `json:"username"`
	Superuser   bool   `json:"superuser"`
	CreateDB    bool   `json:"createdb"`
	CreateRole  bool   `json:"createrole"`
	Login       bool   `json:"login"`
	Replication bool   `json:"replication"`
	ValidUntil  string `json:"valid_until"`
}

// PostgresTableInfo PostgreSQL 表信息（含估计行数）
type PostgresTableInfo struct {
	Name        string `json:"name"`
	Schema      string `json:"schema"`
	Type        string `json:"type"` // BASE TABLE | VIEW | MATERIALIZED VIEW
	RowEstimate int64  `json:"row_estimate"`
}
