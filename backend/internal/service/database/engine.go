package database

import (
	"fmt"

	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// TableInfo 表信息
type TableInfo struct {
	Name string `json:"name"`
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
	KeyName    string `json:"key_name"`
	ColumnName string `json:"column_name"`
	NonUnique  bool   `json:"non_unique"`
	IndexType  string `json:"index_type"`
	SeqInIndex int    `json:"seq_in_index"`
}

// TableConstraint 表约束
type TableConstraint struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // PRIMARY KEY | FOREIGN KEY | UNIQUE | CHECK
	Columns  string `json:"columns"`
	RefTable string `json:"ref_table,omitempty"`
	RefCols  string `json:"ref_columns,omitempty"`
}

// TableSchema 表结构
type TableSchema struct {
	Columns     []TableColumn     `json:"columns"`
	Indexes     []TableIndex      `json:"indexes"`
	Constraints []TableConstraint `json:"constraints"`
}

// QueryResult 分页查询结果
type QueryResult struct {
	Columns  []string        `json:"columns"`
	Rows     [][]interface{} `json:"rows"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// SQLResult SQL 执行结果
type SQLResult struct {
	Columns         []string        `json:"columns"`
	Rows            [][]interface{} `json:"rows"`
	AffectedRows    int64           `json:"affected_rows"`
	ExecutionTimeMs int64           `json:"execution_time_ms"`
}

// UserInfo 用户信息
type UserInfo struct {
	User string `json:"user"`
	Host string `json:"host"`
}

// DatabaseEngine 通用数据库引擎接口
type DatabaseEngine interface {
	ListDatabases() ([]string, error)
	CreateDatabase(name string) error
	DropDatabase(name string) error
	ListTables(dbName string) ([]TableInfo, error)
	GetTableSchema(dbName, table string) (TableSchema, error)
	QueryTableData(dbName, table string, page, pageSize int) (*QueryResult, error)
	ExecuteSQL(dbName, sql string, allowDestructive bool) (*SQLResult, error)
	ListUsers() ([]UserInfo, error)
	CreateUser(username, password, host string) error
	DropUser(username, host string) error
	GrantPrivileges(username, host, dbName string, privileges []string) error
	RevokePrivileges(username, host, dbName string) error
}

// NewDatabaseEngine 根据实例类型创建对应数据库引擎
func NewDatabaseEngine(db *gorm.DB, instanceID uint) (DatabaseEngine, error) {
	var instance model.DatabaseInstance
	if err := db.First(&instance, instanceID).Error; err != nil {
		return nil, fmt.Errorf("实例不存在: %d", instanceID)
	}
	switch instance.Type {
	case "mysql":
		return newMySQLEngine(db, &instance), nil
	case "postgres":
		return newPgEngine(db, &instance), nil
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s（仅支持 mysql / postgres）", instance.Type)
	}
}
