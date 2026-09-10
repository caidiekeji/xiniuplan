package database

import (
	"cloudpivot/internal/model"

	"gorm.io/gorm"
)

// mysqlEngine MySQL 引擎适配器，包装现有 MySQL 服务实现 DatabaseEngine 接口
type mysqlEngine struct {
	instance    *model.DatabaseInstance
	databaseSvc *DatabaseService
	userSvc     *UserService
	sqlSvc      *SQLService
}

// newMySQLEngine 创建 MySQL 引擎
func newMySQLEngine(db *gorm.DB, instance *model.DatabaseInstance) *mysqlEngine {
	return &mysqlEngine{
		instance:    instance,
		databaseSvc: NewDatabaseService(db),
		userSvc:     NewUserService(db),
		sqlSvc:      NewSQLService(db),
	}
}

func (e *mysqlEngine) ListDatabases() ([]string, error) {
	dbs, err := e.databaseSvc.ListDatabases(e.instance.ID)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(dbs))
	for _, d := range dbs {
		result = append(result, d.Name)
	}
	return result, nil
}

func (e *mysqlEngine) CreateDatabase(name string) error {
	return e.databaseSvc.CreateDatabase(e.instance.ID, name, "", "")
}

func (e *mysqlEngine) DropDatabase(name string) error {
	return e.databaseSvc.DeleteDatabase(e.instance.ID, name)
}

func (e *mysqlEngine) ListTables(dbName string) ([]TableInfo, error) {
	tables, err := e.databaseSvc.ListTables(e.instance.ID, dbName)
	if err != nil {
		return nil, err
	}
	result := make([]TableInfo, 0, len(tables))
	for _, t := range tables {
		result = append(result, TableInfo{Name: t.Name})
	}
	return result, nil
}

func (e *mysqlEngine) GetTableSchema(dbName, table string) (TableSchema, error) {
	structure, err := e.databaseSvc.GetTableStructure(e.instance.ID, dbName, table)
	if err != nil {
		return TableSchema{}, err
	}

	schema := TableSchema{
		Columns: make([]TableColumn, 0, len(structure.Columns)),
		Indexes: make([]TableIndex, 0, len(structure.Indexes)),
	}

	for _, col := range structure.Columns {
		schema.Columns = append(schema.Columns, TableColumn{
			Field:   col.Field,
			Type:    col.Type,
			Null:    col.Null,
			Key:     col.Key,
			Default: col.Default,
			Extra:   col.Extra,
		})
	}

	for _, idx := range structure.Indexes {
		schema.Indexes = append(schema.Indexes, TableIndex{
			KeyName:    idx.KeyName,
			ColumnName: idx.ColumnName,
			NonUnique:  idx.NonUnique != 0,
			IndexType:  idx.IndexType,
			SeqInIndex: idx.SeqInIndex,
		})
	}

	return schema, nil
}

func (e *mysqlEngine) QueryTableData(dbName, table string, page, pageSize int) (*QueryResult, error) {
	data, err := e.databaseSvc.ListTableData(e.instance.ID, dbName, table, page, pageSize)
	if err != nil {
		return nil, err
	}
	return &QueryResult{
		Columns:  data.Columns,
		Rows:     data.Rows,
		Total:    data.Total,
		Page:     data.Page,
		PageSize: data.PageSize,
	}, nil
}

func (e *mysqlEngine) ExecuteSQL(dbName, sql string, allowDestructive bool) (*SQLResult, error) {
	result, err := e.sqlSvc.ExecuteSQL(e.instance.ID, dbName, sql, allowDestructive)
	if err != nil {
		return nil, err
	}
	return &SQLResult{
		Columns:         result.Columns,
		Rows:            result.Rows,
		AffectedRows:    result.AffectedRows,
		ExecutionTimeMs: result.ExecutionTimeMs,
	}, nil
}

func (e *mysqlEngine) ListUsers() ([]UserInfo, error) {
	users, err := e.userSvc.ListUsers(e.instance.ID)
	if err != nil {
		return nil, err
	}
	result := make([]UserInfo, 0, len(users))
	for _, u := range users {
		result = append(result, UserInfo{User: u.User, Host: u.Host})
	}
	return result, nil
}

func (e *mysqlEngine) CreateUser(username, password, host string) error {
	return e.userSvc.CreateUser(e.instance.ID, username, password, host)
}

func (e *mysqlEngine) DropUser(username, host string) error {
	return e.userSvc.DeleteUser(e.instance.ID, username, host)
}

func (e *mysqlEngine) GrantPrivileges(username, host, dbName string, privileges []string) error {
	return e.userSvc.GrantPrivileges(e.instance.ID, username, host, dbName, privileges)
}

func (e *mysqlEngine) RevokePrivileges(username, host, dbName string) error {
	return e.userSvc.RevokePrivileges(e.instance.ID, username, host, dbName)
}
