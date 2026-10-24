package api

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"cloudpivot/internal/common"
	dbsvc "cloudpivot/internal/service/database"

	"github.com/gin-gonic/gin"
)

// TestFailDatabaseError 验证数据库模块统一错误出口的映射规则。
//
// 2026-10-24 事故：批量替换错误出口时，误把 failDatabaseError 函数体内部的
// common.Fail 也替换成了 failDatabaseError 自身，形成无限递归 —— 任何走这个
// 出口的错误都会触发 stack overflow 打挂整个后端进程。回归测试锁死映射规则。
func TestFailDatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   int
	}{
		{
			name:       "实例类型不支持属于客户端错误应返回 400",
			err:        dbsvc.NewUnsupportedTypeError("redis", "该操作仅支持 MySQL 实例"),
			wantStatus: 400,
			wantCode:   common.CodeParamError,
		},
		{
			name:       "包装过的哨兵错误仍应识别为 400",
			err:        wrapErr("查询用户列表失败: %w", dbsvc.ErrUnsupportedType),
			wantStatus: 400,
			wantCode:   common.CodeParamError,
		},
		{
			name:       "普通服务端错误返回 500",
			err:        errors.New("数据库连接失败"),
			wantStatus: 500,
			wantCode:   common.CodeServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/api/v1/databases/1/users", nil)

			failDatabaseError(c, tc.err)

			if w.Code != tc.wantStatus {
				t.Errorf("HTTP 状态码 = %d, 期望 %d", w.Code, tc.wantStatus)
			}
			resp := decodeResponse(t, w.Body.Bytes())
			if resp.Code != tc.wantCode {
				t.Errorf("业务码 = %d, 期望 %d", resp.Code, tc.wantCode)
			}
		})
	}
}

// TestFailDatabaseErrorNoRecursion 直接调用错误出口，验证不会栈溢出。
//
// 若 failDatabaseError 内部误写成自调用，本测试会以 stack overflow 崩溃
// （Go 默认栈上限 1GB，测试进程会直接 fatal）。
func TestFailDatabaseErrorNoRecursion(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/databases/1/users", nil)

	failDatabaseError(c, errors.New("普通错误"))

	if w.Code != 500 {
		t.Errorf("HTTP 状态码 = %d, 期望 500", w.Code)
	}
}

// wrapErr 用 %w 包装错误，验证 errors.Is 能穿透包装
func wrapErr(format string, err error) error {
	return &wrappedError{msg: format, err: err}
}

type wrappedError struct {
	msg string
	err error
}

func (w *wrappedError) Error() string { return w.msg }
func (w *wrappedError) Unwrap() error { return w.err }

// decodeResponse 解析统一响应体
func decodeResponse(t *testing.T, body []byte) common.Response {
	t.Helper()
	var resp common.Response
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("解析响应体失败: %v", err)
	}
	return resp
}
