package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hedeqiang/skeleton/internal/config"
)

// CustomClaims 定义了自定义的 JWT 声明
type CustomClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWT 定义了 JWT 工具
type JWT struct {
	secret []byte
	config *config.JWT
}

// NewJWT 创建一个新的 JWT 工具实例
func NewJWT(cfg *config.Config) *JWT {
	return &JWT{
		secret: []byte(cfg.JWT.Secret),
		config: &cfg.JWT,
	}
}

// GenerateToken 生成一个新的 JWT Token
func (j *JWT) GenerateToken(userID uint, username string) (string, error) {
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.ExpireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "go-skeleton", // It's better to get this from config as well
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ParseToken 解析并验证一个 JWT Token
func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}
