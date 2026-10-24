package api

import (
	"log/slog"
	"strconv"
	"strings"

	"cloudpivot/internal/common"
	"cloudpivot/internal/dto"
	"cloudpivot/internal/middleware"
	"cloudpivot/internal/model"
	filesvc "cloudpivot/internal/service/file"

	"github.com/gin-gonic/gin"
)

var fileManager *filesvc.Manager

// initFileManager 懒初始化文件管理器
func initFileManager() *filesvc.Manager {
	if fileManager == nil {
		// 生产环境使用根目录 /，允许管理整个服务器文件系统
		root := "/"
		m, err := filesvc.NewManager(root)
		if err != nil {
			slog.Error("init file manager failed", "error", err)
			panic(err)
		}
		fileManager = m
	}
	return fileManager
}

// SharedFileManager 供 main.go 装配新功能路由复用同一个文件管理器实例
func SharedFileManager() *filesvc.Manager { return initFileManager() }

// RegisterFileRoutes 注册文件管理器路由
func RegisterFileRoutes(r *gin.RouterGroup) {
	m := initFileManager()

	r.GET("/files", middleware.RequirePermission(model.PermFileView), func(c *gin.Context) {
		path := c.DefaultQuery("path", "/")
		data, err := m.ListDir(path)
		if err != nil {
			common.Fail(c, common.CodeParamError, err.Error())
			return
		}
		common.Success(c, data)
	})

	// 文件权限管理（F-005）：chmod / chown
	r.POST("/files/permission", middleware.RequirePermission(model.PermFileUpdate), func(c *gin.Context) {
		var req struct {
			Path      string `json:"path" binding:"required"`
			Mode      string `json:"mode"` // 八进制权限，如 644
			Owner     string `json:"owner"`
			Group     string `json:"group"`
			Recursive bool   `json:"recursive"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
			return
		}
		if req.Mode != "" {
			if err := m.SetPermission(req.Path, req.Mode, req.Recursive); err != nil {
				common.Fail(c, common.CodeServerError, err.Error())
				return
			}
		}
		if req.Owner != "" || req.Group != "" {
			if err := m.SetOwner(req.Path, req.Owner, req.Group, req.Recursive); err != nil {
				common.Fail(c, common.CodeServerError, err.Error())
				return
			}
		}
		common.Success(c, gin.H{"updated": true})
	})

	r.POST("/files/mkdir", middleware.RequirePermission(model.PermFileCreate), func(c *gin.Context) {
		var req struct {
			Path string `json:"path"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
			return
		}
		if req.Path == "" {
			common.Fail(c, common.CodeParamError, "path 不能为空")
			return
		}
		info, err := m.CreateDir(req.Path)
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, info)
	})

	r.POST("/files/create", middleware.RequirePermission(model.PermFileCreate), func(c *gin.Context) {
		var req dto.CreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
			return
		}
		if req.Path == "" {
			common.Fail(c, common.CodeParamError, "path 不能为空")
			return
		}
		if req.Type == "dir" {
			info, err := m.CreateDir(req.Path)
			if err != nil {
				common.Fail(c, common.CodeServerError, err.Error())
				return
			}
			common.Success(c, info)
			return
		}
		info, err := m.CreateFile(req.Path, req.Content)
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, info)
	})

	r.GET("/files/content", middleware.RequirePermission(model.PermFileView), func(c *gin.Context) {
		path := c.Query("path")
		if path == "" {
			common.Fail(c, common.CodeParamError, "path 不能为空")
			return
		}
		content, err := m.ReadFile(path)
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, dto.ContentResponse{
			Path:    path,
			Content: content,
		})
	})

	r.PUT("/files/content", middleware.RequirePermission(model.PermFileUpdate), func(c *gin.Context) {
		var req dto.WriteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
			return
		}
		if req.Path == "" {
			common.Fail(c, common.CodeParamError, "path 不能为空")
			return
		}
		if err := m.WriteFile(req.Path, req.Content); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, nil)
	})

	r.POST("/files/rename", middleware.RequirePermission(model.PermFileUpdate), func(c *gin.Context) {
		var req dto.RenameRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
			return
		}
		if req.OldPath == "" || req.NewPath == "" {
			common.Fail(c, common.CodeParamError, "old_path 和 new_path 不能为空")
			return
		}
		if err := m.Rename(req.OldPath, req.NewPath); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, nil)
	})

	r.POST("/files/copy", middleware.RequirePermission(model.PermFileCreate), func(c *gin.Context) {
		var req dto.CopyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
			return
		}
		if req.Src == "" || req.Dst == "" {
			common.Fail(c, common.CodeParamError, "src 和 dst 不能为空")
			return
		}
		if err := m.Copy(req.Src, req.Dst); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, nil)
	})

	r.POST("/files/move", middleware.RequirePermission(model.PermFileUpdate), func(c *gin.Context) {
		var req dto.MoveRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
			return
		}
		if req.Src == "" || req.Dst == "" {
			common.Fail(c, common.CodeParamError, "src 和 dst 不能为空")
			return
		}
		if err := m.Move(req.Src, req.Dst); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, nil)
	})

	r.DELETE("/files", middleware.RequirePermission(model.PermFileDelete), func(c *gin.Context) {
		pathsStr := c.Query("paths")
		if pathsStr == "" {
			common.Fail(c, common.CodeParamError, "paths 不能为空")
			return
		}
		paths := strings.Split(pathsStr, ",")
		if c.Query("forever") == "1" {
			if err := m.DeleteForever(paths); err != nil {
				common.Fail(c, common.CodeServerError, err.Error())
				return
			}
			common.Success(c, gin.H{"recycled": false})
			return
		}
		if err := m.Delete(paths); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		m.RecordRecycle(paths, model.DB)
		common.Success(c, gin.H{"recycled": true})
	})

	r.POST("/files/compress", middleware.RequirePermission(model.PermFileCreate), func(c *gin.Context) {
		var req dto.CompressRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
			return
		}
		if len(req.SrcPaths) == 0 || req.DestPath == "" {
			common.Fail(c, common.CodeParamError, "src_paths 和 dest_path 不能为空")
			return
		}
		if req.Format == "" {
			req.Format = "zip"
		}
		if err := m.Compress(req.SrcPaths, req.DestPath, req.Format); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, nil)
	})

	r.POST("/files/decompress", middleware.RequirePermission(model.PermFileUpdate), func(c *gin.Context) {
		var req dto.DecompressRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.Fail(c, common.CodeParamError, "参数错误: "+err.Error())
			return
		}
		if req.ArchivePath == "" || req.DestPath == "" {
			common.Fail(c, common.CodeParamError, "archive_path 和 dest_path 不能为空")
			return
		}
		if err := m.Decompress(req.ArchivePath, req.DestPath); err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, nil)
	})

	r.GET("/files/search", middleware.RequirePermission(model.PermFileView), func(c *gin.Context) {
		root := c.DefaultQuery("root", "/")
		keyword := c.Query("keyword")
		if keyword == "" {
			common.Fail(c, common.CodeParamError, "keyword 不能为空")
			return
		}
		useRegex, _ := strconv.ParseBool(c.DefaultQuery("regex", "false"))
		results, err := m.Search(root, keyword, useRegex)
		if err != nil {
			common.Fail(c, common.CodeServerError, err.Error())
			return
		}
		common.Success(c, results)
	})

	r.POST("/files/upload", middleware.RequirePermission(model.PermFileCreate), func(c *gin.Context) {
		destDir := c.DefaultPostForm("path", "/")
		info, err := m.UploadFile(c.Request, destDir)
		if err != nil {
			common.Fail(c, common.CodeServerError, "上传失败: "+err.Error())
			return
		}
		common.Success(c, info)
	})

	r.GET("/files/download", middleware.RequirePermission(model.PermFileView), func(c *gin.Context) {
		path := c.Query("path")
		if path == "" {
			common.Fail(c, common.CodeParamError, "path 不能为空")
			return
		}
		if err := m.DownloadFile(c.Writer, path); err != nil {
			common.Fail(c, common.CodeServerError, "下载失败: "+err.Error())
			return
		}
	})
}
