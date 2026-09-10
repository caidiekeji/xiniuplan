// Package web 嵌入前端构建产物（frontend/dist → web/dist）
// 构建流水线会把前端 dist 复制到本目录；本地开发若未构建，
// dist/index.html 占位文件保证 embed 可编译（访问时返回占位提示）。
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// Dist 返回前端静态文件系统（根为 dist/）
func Dist() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
