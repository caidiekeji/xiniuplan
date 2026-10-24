package api

import (
	"errors"
	"fmt"
	"strconv"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	dbsvc "cloudpivot/internal/service/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// failDatabaseError 统一的数据库模块错误出口。
//
// service 层用 dbsvc.ErrUnsupportedType 标记「实例类型不支持当前操作」、
// dbsvc.ErrInvalidRequest 标记「参数不合法」这两类客户端语义错误，
// 这里翻译成 400 而不是 500 —— 否则前端会弹出系统异常提示，掩盖真实原因
// （用户对实例类型用错了功能入口，或输入了非法权限/标识符）。
func failDatabaseError(c *gin.Context, err error) {
	if errors.Is(err, dbsvc.ErrUnsupportedType) || errors.Is(err, dbsvc.ErrInvalidRequest) {
		common.Fail(c, common.CodeParamError, err.Error())
		return
	}
	common.Fail(c, common.CodeServerError, err.Error())
}

// DatabaseHandler 数据库管理 API Handler
type DatabaseHandler struct {
	instanceSvc *dbsvc.InstanceService
	databaseSvc *dbsvc.DatabaseService
	userSvc     *dbsvc.UserService
	sqlSvc      *dbsvc.SQLService
	tuningSvc   *dbsvc.TuningService
}

// NewDatabaseHandler 创建数据库管理 Handler
func NewDatabaseHandler(db *gorm.DB) *DatabaseHandler {
	return &DatabaseHandler{
		instanceSvc: dbsvc.NewInstanceService(db, "./runtime"),
		databaseSvc: dbsvc.NewDatabaseService(db),
		userSvc:     dbsvc.NewUserService(db),
		sqlSvc:      dbsvc.NewSQLService(db),
		tuningSvc:   dbsvc.NewTuningService(db),
	}
}

// RegisterDatabaseRoutes 注册数据库管理路由
func RegisterDatabaseRoutes(r *gin.RouterGroup) {
	h := NewDatabaseHandler(model.DB)

	// 实例管理
	r.GET("/databases", middleware.RequirePermission(model.PermDatabaseView), h.ListInstances)
	r.POST("/databases", middleware.RequirePermission(model.PermDatabaseCreate), h.CreateInstance)
	r.POST("/databases/import", middleware.RequirePermission(model.PermDatabaseCreate), h.ImportInstance)
	r.GET("/databases/discover", middleware.RequirePermission(model.PermDatabaseView), h.DiscoverInstances)
	r.GET("/databases/:id", middleware.RequirePermission(model.PermDatabaseView), h.GetInstance)
	r.DELETE("/databases/:id", middleware.RequirePermission(model.PermDatabaseDelete), h.DeleteInstance)
	r.POST("/databases/:id/start", middleware.RequirePermission(model.PermDatabaseUpdate), h.StartInstance)
	r.POST("/databases/:id/stop", middleware.RequirePermission(model.PermDatabaseUpdate), h.StopInstance)
	r.POST("/databases/:id/restart", middleware.RequirePermission(model.PermDatabaseUpdate), h.RestartInstance)

	// 库表管理
	r.GET("/databases/:id/dbs", middleware.RequirePermission(model.PermDatabaseView), h.ListDatabases)
	r.POST("/databases/:id/dbs", middleware.RequirePermission(model.PermDatabaseCreate), h.CreateDatabase)
	r.DELETE("/databases/:id/dbs/:name", middleware.RequirePermission(model.PermDatabaseDelete), h.DeleteDatabase)
	r.GET("/databases/:id/dbs/:name/tables", middleware.RequirePermission(model.PermDatabaseView), h.ListTables)
	r.GET("/databases/:id/dbs/:name/tables/:table", middleware.RequirePermission(model.PermDatabaseView), h.GetTableStructure)
	r.GET("/databases/:id/dbs/:name/tables/:table/data", middleware.RequirePermission(model.PermDatabaseView), h.ListTableData)

	// 用户权限
	r.GET("/databases/:id/users", middleware.RequirePermission(model.PermDatabaseView), h.ListUsers)
	r.POST("/databases/:id/users", middleware.RequirePermission(model.PermDatabaseCreate), h.CreateUser)
	r.DELETE("/databases/:id/users/:username", middleware.RequirePermission(model.PermDatabaseDelete), h.DeleteUser)
	r.POST("/databases/:id/users/:username/grant", middleware.RequirePermission(model.PermDatabaseUpdate), h.GrantPrivileges)
	r.POST("/databases/:id/users/:username/revoke", middleware.RequirePermission(model.PermDatabaseUpdate), h.RevokePrivileges)

	// SQL 执行
	r.POST("/databases/:id/sql", middleware.RequirePermission(model.PermDatabaseUpdate), h.ExecuteSQL)
	r.POST("/databases/:id/import", middleware.RequirePermission(model.PermDatabaseUpdate), h.ImportSQL)
	r.GET("/databases/:id/export", middleware.RequirePermission(model.PermDatabaseView), h.ExportSQL)

	// 参数配置（DB-002）与状态监控（DB-004）
	r.GET("/databases/:id/variables", middleware.RequirePermission(model.PermDatabaseView), h.ListVariables)
	r.PUT("/databases/:id/variables", middleware.RequirePermission(model.PermDatabaseUpdate), h.UpdateVariable)
	r.GET("/databases/:id/status", middleware.RequirePermission(model.PermDatabaseView), h.DBStatus)
}

// parseInstanceID 从路径参数解析实例 ID
func parseInstanceID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的实例 ID")
		return 0, false
	}
	return uint(id), true
}

// parsePagination 解析分页参数
func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	return page, pageSize
}

// ========== 实例管理 ==========

// ListInstances GET /databases
func (h *DatabaseHandler) ListInstances(c *gin.Context) {
	page, pageSize := parsePagination(c)
	instances, total, err := h.instanceSvc.ListInstances(page, pageSize)
	if err != nil {
		failDatabaseError(c, fmt.Errorf("获取实例列表失败: %w", err))
		return
	}

	// 转换为响应（补充 container_status）
	responses := make([]dto.DBInstanceResponse, 0, len(instances))
	for _, inst := range instances {
		responses = append(responses, toInstanceResponse(&inst, ""))
	}
	common.Page(c, responses, total, page, pageSize)
}

// CreateInstance POST /databases
func (h *DatabaseHandler) CreateInstance(c *gin.Context) {
	var req dto.CreateDBInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	instance, err := h.instanceSvc.CreateInstance(req)
	if err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, toInstanceResponse(instance, ""))
}

// ImportInstance POST /databases/import 导入已有数据库实例
func (h *DatabaseHandler) ImportInstance(c *gin.Context) {
	var req dto.ImportDBInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}

	instance, err := h.instanceSvc.ImportInstance(req)
	if err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, toInstanceResponse(instance, ""))
}

// DiscoverInstances GET /databases/discover 自动发现本机已安装的数据库
func (h *DatabaseHandler) DiscoverInstances(c *gin.Context) {
	instances, err := h.instanceSvc.DiscoverInstances()
	if err != nil {
		failDatabaseError(c, fmt.Errorf("发现数据库失败: "+"%w", err))
		return
	}
	common.Success(c, instances)
}

// GetInstance GET /databases/:id
func (h *DatabaseHandler) GetInstance(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}

	instance, err := h.instanceSvc.GetInstance(id)
	if err != nil {
		common.Fail(c, common.CodeNotFound, err.Error())
		return
	}

	// 获取真实容器状态
	containerStatus, _ := h.instanceSvc.GetInstanceStatus(id)
	common.Success(c, toInstanceResponse(instance, containerStatus))
}

// DeleteInstance DELETE /databases/:id
func (h *DatabaseHandler) DeleteInstance(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}

	cleanup := c.DefaultQuery("cleanup", "false") == "true"
	if err := h.instanceSvc.DeleteInstance(id, cleanup); err != nil {
		failDatabaseError(c, fmt.Errorf("删除实例失败: "+"%w", err))
		return
	}
	common.Success(c, nil)
}

// StartInstance POST /databases/:id/start
func (h *DatabaseHandler) StartInstance(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	if err := h.instanceSvc.StartInstance(id); err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, nil)
}

// StopInstance POST /databases/:id/stop
func (h *DatabaseHandler) StopInstance(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	if err := h.instanceSvc.StopInstance(id); err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, nil)
}

// RestartInstance POST /databases/:id/restart
func (h *DatabaseHandler) RestartInstance(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	if err := h.instanceSvc.RestartInstance(id); err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, nil)
}

// ========== 库表管理 ==========

// ListDatabases GET /databases/:id/dbs
func (h *DatabaseHandler) ListDatabases(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	dbs, err := h.databaseSvc.ListDatabases(id)
	if err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, dbs)
}

// CreateDatabase POST /databases/:id/dbs
func (h *DatabaseHandler) CreateDatabase(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	var req dto.CreateDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if err := h.databaseSvc.CreateDatabase(id, req.Name, req.Charset, req.Collation); err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, nil)
}

// DeleteDatabase DELETE /databases/:id/dbs/:name
func (h *DatabaseHandler) DeleteDatabase(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	name := c.Param("name")
	if err := h.databaseSvc.DeleteDatabase(id, name); err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, nil)
}

// ListTables GET /databases/:id/dbs/:name/tables
func (h *DatabaseHandler) ListTables(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	dbName := c.Param("name")
	tables, err := h.databaseSvc.ListTables(id, dbName)
	if err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, tables)
}

// GetTableStructure GET /databases/:id/dbs/:name/tables/:table
func (h *DatabaseHandler) GetTableStructure(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	dbName := c.Param("name")
	tableName := c.Param("table")
	structure, err := h.databaseSvc.GetTableStructure(id, dbName, tableName)
	if err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, structure)
}

// ListTableData GET /databases/:id/dbs/:name/tables/:table/data
func (h *DatabaseHandler) ListTableData(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	dbName := c.Param("name")
	tableName := c.Param("table")
	page, pageSize := parsePagination(c)

	data, err := h.databaseSvc.ListTableData(id, dbName, tableName, page, pageSize)
	if err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, data)
}

// ========== 用户权限 ==========

// ListUsers GET /databases/:id/users
func (h *DatabaseHandler) ListUsers(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	users, err := h.userSvc.ListUsers(id)
	if err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, users)
}

// CreateUser POST /databases/:id/users
func (h *DatabaseHandler) CreateUser(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	var req dto.CreateDatabaseUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if err := h.userSvc.CreateUser(id, req.Username, req.Password, req.Host); err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, nil)
}

// DeleteUser DELETE /databases/:id/users/:username?host=xxx
func (h *DatabaseHandler) DeleteUser(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	username := c.Param("username")
	host := c.DefaultQuery("host", "%")
	if err := h.userSvc.DeleteUser(id, username, host); err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, nil)
}

// GrantPrivileges POST /databases/:id/users/:username/grant
func (h *DatabaseHandler) GrantPrivileges(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	username := c.Param("username")
	var req dto.GrantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	req.Username = username
	if err := h.userSvc.GrantPrivileges(id, req.Username, req.Host, req.DBName, req.Privileges); err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, nil)
}

// RevokePrivileges POST /databases/:id/users/:username/revoke
func (h *DatabaseHandler) RevokePrivileges(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	username := c.Param("username")
	host := c.DefaultQuery("host", "%")
	dbName := c.Query("db_name")
	if dbName == "" {
		common.Fail(c, common.CodeParamError, "缺少 db_name 参数")
		return
	}
	if err := h.userSvc.RevokePrivileges(id, username, host, dbName); err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, nil)
}

// ========== SQL 执行 ==========

// ExecuteSQL POST /databases/:id/sql
func (h *DatabaseHandler) ExecuteSQL(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	var req dto.ExecuteSQLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	result, err := h.sqlSvc.ExecuteSQL(id, req.DBName, req.SQL, req.AllowDestructive)
	if err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, result)
}

// ImportSQL POST /databases/:id/import
func (h *DatabaseHandler) ImportSQL(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	var req dto.ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	result, err := h.sqlSvc.ImportSQL(id, req.DBName, req.Content)
	if err != nil {
		failDatabaseError(c, err)
		return
	}
	common.Success(c, result)
}

// ExportSQL GET /databases/:id/export?db_name=xxx&table_name=yyy
func (h *DatabaseHandler) ExportSQL(c *gin.Context) {
	id, ok := parseInstanceID(c)
	if !ok {
		return
	}
	dbName := c.Query("db_name")
	tableName := c.Query("table_name")
	if dbName == "" {
		common.Fail(c, common.CodeParamError, "缺少 db_name 参数")
		return
	}

	sqlText, err := h.sqlSvc.ExportSQL(id, dbName, tableName)
	if err != nil {
		failDatabaseError(c, fmt.Errorf("导出失败: "+"%w", err))
		return
	}

	// 设置下载响应头
	fileName := dbName
	if tableName != "" {
		fileName = dbName + "_" + tableName
	}
	c.Header("Content-Type", "application/sql; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+fileName+".sql")
	c.String(200, sqlText)
}

// ========== 辅助函数 ==========

// toInstanceResponse 转换模型为响应 DTO
func toInstanceResponse(inst *model.DatabaseInstance, containerStatus string) dto.DBInstanceResponse {
	if containerStatus == "" {
		containerStatus = inst.Status
	}
	return dto.DBInstanceResponse{
		ID:              inst.ID,
		Type:            inst.Type,
		Version:         inst.Version,
		Host:            inst.Host,
		Port:            inst.Port,
		Username:        inst.Username,
		ContainerID:     inst.ContainerID,
		IsExternal:      inst.IsExternal,
		Status:          inst.Status,
		ContainerStatus: containerStatus,
		CreatedAt:       inst.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ListVariables GET /databases/:id/variables?search= - 参数列表（DB-002）
func (h *DatabaseHandler) ListVariables(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的实例 ID")
		return
	}
	vars, err := h.tuningSvc.GetVariables(uint(id), c.Query("search"))
	if err != nil {
		failDatabaseError(c, fmt.Errorf("获取参数列表失败: "+"%w", err))
		return
	}
	common.Success(c, vars)
}

// UpdateVariable PUT /databases/:id/variables - 修改参数（DB-002）
func (h *DatabaseHandler) UpdateVariable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的实例 ID")
		return
	}
	var req struct {
		Name  string `json:"name" binding:"required"`
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
		return
	}
	if err := h.tuningSvc.UpdateVariable(uint(id), req.Name, req.Value); err != nil {
		failDatabaseError(c, fmt.Errorf("修改参数失败: "+"%w", err))
		return
	}
	common.Success(c, gin.H{"updated": true})
}

// DBStatus GET /databases/:id/status - 实例状态监控（DB-004）
func (h *DatabaseHandler) DBStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.Fail(c, common.CodeParamError, "无效的实例 ID")
		return
	}
	status, err := h.tuningSvc.GetStatus(uint(id))
	if err != nil {
		failDatabaseError(c, fmt.Errorf("获取实例状态失败: "+"%w", err))
		return
	}
	common.Success(c, status)
}
