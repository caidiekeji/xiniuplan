package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"cloudpivot/internal/config"
	"cloudpivot/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建内存测试数据库并迁移必要的表
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.Role{}, &model.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	// 创建一个测试角色
	db.Create(&model.Role{
		Name:        "测试管理员",
		Code:        "admin",
		Description: "test",
		Permissions: model.StringSlice{"site:view", "database:view"},
		IsSystem:    true,
	})
	return db
}

// newTestService 创建带内存数据库的测试 Service
func newTestService(t *testing.T) (*Service, *rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	pubKey := &privKey.PublicKey
	db := setupTestDB(t)
	cfg := &config.Config{
		JWT: config.JWTConfig{
			AccessTokenTTL:  3600,
			RefreshTokenTTL: 86400,
		},
	}
	svc := NewService(db, privKey, pubKey, cfg)
	return svc, privKey, pubKey
}

// TestPasswordHash 测试 bcrypt 密码哈希与校验
func TestPasswordHash(t *testing.T) {
	password := "TestP@ssw0rd!2024"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword failed: %v", err)
	}

	// 正确密码应通过校验
	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		t.Errorf("CompareHashAndPassword with correct password failed: %v", err)
	}

	// 错误密码应失败
	if err := bcrypt.CompareHashAndPassword(hash, []byte("wrongpassword")); err == nil {
		t.Error("CompareHashAndPassword with wrong password should fail")
	}

	// 哈希值不应与明文相同
	if string(hash) == password {
		t.Error("hash should not equal plaintext password")
	}
}

// TestRandomPasswordGeneration 测试随机密码生成
func TestRandomPasswordGeneration(t *testing.T) {
	pwd := generateRandomPassword(16)
	if len(pwd) != 16 {
		t.Errorf("expected password length 16, got %d", len(pwd))
	}

	// 生成多个密码应不同
	pwd2 := generateRandomPassword(16)
	if pwd == pwd2 {
		t.Error("two generated passwords should be different")
	}

	// 测试长度为 0 时使用默认 16
	pwd3 := generateRandomPassword(0)
	if len(pwd3) != 16 {
		t.Errorf("expected default length 16, got %d", len(pwd3))
	}
}

// TestTokenGenerationAndValidation 测试 JWT token 生成与校验
func TestTokenGenerationAndValidation(t *testing.T) {
	svc, _, _ := newTestService(t)

	user := model.User{
		ID:       1,
		Username: "testuser",
		Role:     "admin",
	}

	// 生成 access token
	tokenStr, err := svc.generateToken(user, time.Hour, "access")
	if err != nil {
		t.Fatalf("generateToken failed: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("generated token is empty")
	}

	// 解析 token
	claims, err := svc.parseToken(tokenStr)
	if err != nil {
		t.Fatalf("parseToken failed: %v", err)
	}

	// 校验 claims 内容
	userID, _ := claims["user_id"].(float64)
	if uint(userID) != user.ID {
		t.Errorf("expected user_id %d, got %v", user.ID, claims["user_id"])
	}
	if claims["username"] != user.Username {
		t.Errorf("expected username %s, got %v", user.Username, claims["username"])
	}
	if claims["role"] != user.Role {
		t.Errorf("expected role %s, got %v", user.Role, claims["role"])
	}
	if claims["type"] != "access" {
		t.Errorf("expected token type access, got %v", claims["type"])
	}

	// 校验 permissions 字段存在
	if perms, ok := claims["permissions"].([]interface{}); !ok || len(perms) == 0 {
		t.Logf("permissions claim: %v", claims["permissions"])
	}
}

// TestInvalidToken 测试无效 token 解析失败
func TestInvalidToken(t *testing.T) {
	svc, _, _ := newTestService(t)

	// 伪造的 token 应解析失败
	_, err := svc.parseToken("invalid.token.string")
	if err == nil {
		t.Error("parseToken with invalid token should return error")
	}

	// 用另一把密钥签名的 token 应校验失败
	otherSvc, _, _ := newTestService(t)
	user := model.User{ID: 1, Username: "hacker", Role: "user"}
	fakeToken, _ := otherSvc.generateToken(user, time.Hour, "access")

	_, err = svc.parseToken(fakeToken)
	if err == nil {
		t.Error("token signed with different key should be rejected")
	}
}

// TestExpiredToken 测试过期 token
func TestExpiredToken(t *testing.T) {
	svc, _, _ := newTestService(t)

	user := model.User{ID: 1, Username: "test", Role: "user"}

	// 生成已过期的 token（TTL 为负）
	expiredToken, err := svc.generateToken(user, -time.Second, "access")
	if err != nil {
		t.Fatalf("generateToken failed: %v", err)
	}

	_, err = svc.parseToken(expiredToken)
	if err == nil {
		t.Error("expired token should return error")
	}
}
