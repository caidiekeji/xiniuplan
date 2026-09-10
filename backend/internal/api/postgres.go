package api

import (
	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	dbsvc "cloudpivot/internal/service/database"

	"github.com/gin-gonic/gin"
)

// PostgresHandler PostgreSQL 管理 API Handler
type PostgresHandler struct {
	pgSvc *dbsvc.PostgresService
}

// NewPostgresHandler 创建 PostgreSQL 管理 Handler
func NewPostgresHandler() *PostgresHandler {
	return &PostgresHandler{
		pgSvc: dbsvc.NewPostgresService(model.DB),
	}
}

// RegisterPostgresRoutes 注册 PostgreSQL 管理路由
func RegisterPostgresRoutes(r *gin.RouterGroup) {
	h := NewPostgresHandler()

	// 库管理
	r.GET("/postgres/:id/dbs", middleware.RequirePermission(model.PermPostgresView), h.ListDatabases)
	r.POST("/postgres/:id/dbs", middleware.RequirePermission(model.PermPostgresManage), h.CreateDatabase)
	r.DELETE("/postgres/:id/dbs/:name", middleware.RequirePermission(model.PermPostgresManage), h.DropDatabase)

	// 表管理
	r.GET("/postgres/:id/dbs/:db/tables", middleware.RequirePermission(model.PermPostgresView), h.ListTables)
	r.GET("/postgres/:id/dbs/:db/tables/:table", middleware.RequirePermission(model.PermPostgresView), h.GetTableSchema)
	r.GET("/postgres/:id/dbs/:db/tables/:table/data", middleware.RequirePermission(model.PermPostgresView), h.QueryTableData)

	// SQL 执行
	r.POST("/postgres/:id/sql", middleware.RequirePermission(model.PermPostgresManage), h.ExecuteSQL)

	// 用户权限
	r.GET("/postgres/:id/users", middleware.RequirePermission(model.PermPostgresView), h.ListUsers)
	r.POST("/postgres/:id/users", middleware.RequirePermission(model.PermPostgresManage), h.CreateUser)
	r.DELETE("/postgres/:id/users/:name", middleware.RequirePermission(model.PermPostgresManage), h.DropUser)
	r.POST("/postgres/:id/users/:name/grant", middleware.RequirePermission(model.PermPostgresManage), h.GrantPrivileges)
	r.POST("/postgres/:id/users/:name/revoke", middleware.RequirePermission(model.PermPostgresManage), h.RevokePrivileges)
}

// ========== 库管理 ==========

// ListDatabases GET /postgres/:id/dbs
func (h *PostgresHandler) ListDatabases(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	engine, err := h.pgSvc.GetEngine(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	dbs, err := engine.ListDatabases()
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, dbs)
}

// CreateDatabase POST /postgres/:id/dbs
func (h *PostgresHandler) CreateDatabase(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	var req dto.CreatePostgresDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	engine, err := h.pgSvc.GetEngine(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	if err := engine.CreateDatabaseEx(req.Name, req.Encoding, req.Locale, req.Owner); err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, nil)
}

// DropDatabase DELETE /postgres/:id/dbs/:name
func (h *PostgresHandler) DropDatabase(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	name := c.Param("name")
	engine, err := h.pgSvc.GetEngine(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	if err := engine.DropDatabase(name); err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, nil)
}

// ========== 表管理 ==========

// ListTables GET /postgres/:id/dbs/:db/tables
func (h *PostgresHandler) ListTables(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	dbName := c.Param("db")
	engine, err := h.pgSvc.GetEngine(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	tables, err := engine.ListTablesEx(dbName)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, tables)
}

// GetTableSchema GET /postgres/:id/dbs/:db/tables/:table
func (h *PostgresHandler) GetTableSchema(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	dbName := c.Param("db")
	tableName := c.Param("table")
	engine, err := h.pgSvc.GetEngine(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	schema, err := engine.GetTableSchema(dbName, tableName)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, schema)
}

// QueryTableData GET /postgres/:id/dbs/:db/tables/:table/data
func (h *PostgresHandler) QueryTableData(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	dbName := c.Param("db")
	tableName := c.Param("table")
	page, pageSize := parsePagination(c)

	engine, err := h.pgSvc.GetEngine(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	result, err := engine.QueryTableData(dbName, tableName, page, pageSize)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, result)
}

// ========== SQL 执行 ==========

// ExecuteSQL POST /postgres/:id/sql
func (h *PostgresHandler) ExecuteSQL(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	var req dto.ExecuteSQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	engine, err := h.pgSvc.GetEngine(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	result, err := engine.ExecuteSQL(req.DBName, req.SQL, req.AllowDestructive)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, result)
}

// ========== 用户权限 ==========

// ListUsers GET /postgres/:id/users
func (h *PostgresHandler) ListUsers(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	engine, err := h.pgSvc.GetEngine(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	users, err := engine.ListUsersEx()
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, users)
}

// CreateUser POST /postgres/:id/users
func (h *PostgresHandler) CreateUser(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	var req dto.CreatePostgresUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	// 默认允许登录（Go bool 零值为 false，创建用户通常需要登录权限）
	req.Login = true
	engine, err := h.pgSvc.GetEngine(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	if err := engine.CreateUserEx(req); err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, nil)
}

// DropUser DELETE /postgres/:id/users/:name
func (h *PostgresHandler) DropUser(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	name := c.Param("name")
	engine, err := h.pgSvc.GetEngine(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	if err := engine.DropUser(name, ""); err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, nil)
}

// GrantPrivileges POST /postgres/:id/users/:name/grant
func (h *PostgresHandler) GrantPrivileges(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	username := c.Param("name")
	var req dto.PostgresGrantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	dbName := c.Query("db_name")
	if dbName == "" {
		common.Fail(c, common.CodeParamError, "缺少 db_name 查询参数")
		return
	}
	engine, err := h.pgSvc.GetEngine(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	if err := engine.GrantPrivileges(username, "", dbName, req.Privileges); err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, nil)
}

// RevokePrivileges POST /postgres/:id/users/:name/revoke
func (h *PostgresHandler) RevokePrivileges(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	username := c.Param("name")
	dbName := c.Query("db_name")
	if dbName == "" {
		common.Fail(c, common.CodeParamError, "缺少 db_name 查询参数")
		return
	}
	engine, err := h.pgSvc.GetEngine(id)
	if err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	if err := engine.RevokePrivileges(username, "", dbName); err != nil {
		common.Fail(c, common.CodeServerError, err.Error())
		return
	}
	common.Success(c, nil)
}
