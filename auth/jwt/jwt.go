package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey

	// 默认密钥路径
	defaultPrivateKeyPath = "/opt/oidc/private.pem"
	defaultPublicKeyPath  = "/opt/oidc/public.pem"
)

// InitKeys 初始化 RSA 密钥对
func InitKeys(privateKeyPath, publicKeyPath string) error {
	// 如果没有提供路径，使用默认路径
	if privateKeyPath == "" {
		privateKeyPath = defaultPrivateKeyPath
	}
	if publicKeyPath == "" {
		publicKeyPath = defaultPublicKeyPath
	}

	// 尝试加载私钥
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		// 如果私钥不存在，生成新的密钥对
		if err := generateKeyPair(privateKeyPath, publicKeyPath); err != nil {
			return fmt.Errorf("failed to generate key pair: %w", err)
		}
	}

	// 加载私钥
	privKey, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}
	privateKey = privKey
	publicKey = &privKey.PublicKey

	return nil
}

// InitKeysFromEnv 从环境变量初始化密钥，如果没有设置环境变量则使用默认路径
func InitKeysFromEnv() error {
	privateKeyPath := os.Getenv("OIDC_PRIVATE_KEY_PATH")
	publicKeyPath := os.Getenv("OIDC_PUBLIC_KEY_PATH")

	return InitKeys(privateKeyPath, publicKeyPath)
}

// generateKeyPair 生成 RSA 密钥对
func generateKeyPair(privateKeyPath, publicKeyPath string) error {
	// 生成 2048 位的 RSA 密钥
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// 保存私钥
	privFile, err := os.Create(privateKeyPath)
	if err != nil {
		return err
	}
	defer privFile.Close()

	privPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	}
	if err := pem.Encode(privFile, privPEM); err != nil {
		return err
	}

	// 保存公钥
	pubFile, err := os.Create(publicKeyPath)
	if err != nil {
		return err
	}
	defer pubFile.Close()

	pubPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privKey.PublicKey),
	}
	if err := pem.Encode(pubFile, pubPEM); err != nil {
		return err
	}

	return nil
}

// loadPrivateKey 加载私钥
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

// LoadPublicKey 加载公钥
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

// Claims 自定义 JWT Claims
type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成 Access Token
func GenerateAccessToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(1 * time.Hour).Unix(),
		"typ": "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

// GenerateIDToken 生成 ID Token
func GenerateIDToken(userID, email, name string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			Issuer:    "oidc-server",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privateKey)
}

// ValidateToken 验证 Token 并返回 Claims
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateAccessToken 验证 Access Token
func ValidateAccessToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, ok := claims["sub"].(string); ok {
			return userID, nil
		}
	}

	return "", errors.New("invalid token claims")
}

// GenerateState 生成 state 参数
func GenerateState() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

// GenerateAuthCode 生成授权码
func GenerateAuthCode() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

// GetPublicKey 获取公钥（用于 JWKS 端点）
func GetPublicKey() *rsa.PublicKey {
	return publicKey
}
