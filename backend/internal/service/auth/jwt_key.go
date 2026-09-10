package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

const (
	privateKeyFile = "private.pem"
	publicKeyFile  = "public.pem"
)

// LoadOrGenerateKeys 加载或生成 RSA 密钥对
func LoadOrGenerateKeys(keyDir string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	if err := os.MkdirAll(keyDir, 0700); err != nil {
		return nil, nil, fmt.Errorf("create key dir: %w", err)
	}

	privPath := filepath.Join(keyDir, privateKeyFile)
	pubPath := filepath.Join(keyDir, publicKeyFile)

	// 尝试加载已有密钥
	if _, err := os.Stat(privPath); err == nil {
		privKey, err := loadPrivateKey(privPath)
		if err != nil {
			return nil, nil, err
		}
		return privKey, &privKey.PublicKey, nil
	}

	// 生成新密钥对
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("generate rsa key: %w", err)
	}

	// 保存私钥
	privBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})
	if err := os.WriteFile(privPath, privPEM, 0600); err != nil {
		return nil, nil, fmt.Errorf("write private key: %w", err)
	}

	// 保存公钥
	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal public key: %w", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})
	if err := os.WriteFile(pubPath, pubPEM, 0644); err != nil {
		return nil, nil, fmt.Errorf("write public key: %w", err)
	}

	return privKey, &privKey.PublicKey, nil
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("decode private key pem failed")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}
	return key, nil
}
