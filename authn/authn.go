package authn

import (
	"context"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// IToken 定义了实现通用令牌的方法.
type IToken interface {
	GetToken() string              // 获取令牌字符串。
	GetTokenType() string          // 获取令牌类型。
	GetExpiresAt() int64           // 获取令牌过期时间戳。
	EncodeToJSON() ([]byte, error) // JSON 编码
}

// Authenticator 定义了用于令牌处理的方法.
type Authenticator interface {
	// Sign 用于生成一个令牌.
	Sign(ctx context.Context, userID string) (IToken, error)
	// Destroy 用于销毁一个令牌.
	Destroy(ctx context.Context, accessToken string) error
	// ParseClaims 解析令牌并返回声明.
	ParseClaims(ctx context.Context, accessToken string) (*jwt.RegisteredClaims, error)
	// Release 用于释放所请求的资源.
	Release() error
}

// Encrypt 使用bcrypt对纯文本进行加密.
func Encrypt(source string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(source), bcrypt.DefaultCost)
	return string(hashedBytes), err
}

// Compare 如果加密文本和纯文本相同，Compare会对其进行比较.
func Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
