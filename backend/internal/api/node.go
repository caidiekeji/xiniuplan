package api

import (
	"strconv"

	"cloudpivot/internal/common"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	nodesvc "cloudpivot/internal/service/node"

	"github.com/gin-gonic/gin"
)

// RegisterNodeRoutes 注册多节点管理路由（roadmap #13）
func RegisterNodeRoutes(r *gin.RouterGroup, svc *nodesvc.Service) {
	g := r.Group("/nodes", middleware.RequirePermission(model.PermMonitorView))

	g.GET("", func(c *gin.Context) {
		list, err := svc.List()
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"items": list})
	})

	g.POST("", middleware.RequirePermission(model.PermSettingUpdate), func(c *gin.Context) {
		var n model.Node
		if err := c.ShouldBindJSON(&n); err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		n.ID = 0
		if n.Name == "" {
			n.Name = n.BaseURL
		}
		if err := svc.Create(&n); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, n)
	})

	g.DELETE("/:id", middleware.RequirePermission(model.PermSettingUpdate), func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := svc.Delete(uint(id)); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"deleted": true})
	})

	g.POST("/:id/refresh", func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		n, err := svc.Refresh(uint(id))
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, n)
	})

	g.POST("/:id/enable", middleware.RequirePermission(model.PermSettingUpdate), func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := svc.SetEnabled(uint(id), true); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"enabled": true})
	})

	g.POST("/:id/disable", middleware.RequirePermission(model.PermSettingUpdate), func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := svc.SetEnabled(uint(id), false); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, gin.H{"enabled": false})
	})
}
