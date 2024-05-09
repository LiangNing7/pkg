package jwt

import "encoding/json"

// tokenInfo authn.IToken 接口的实现
type tokenInfo struct {
	Token     string `json:"token"`    // 令牌字符串
	Type      string `json:"type"`     // 令牌类型
	ExpiresAt int64  `json:"expireAt"` // 令牌过期时间
}

// GetToken 获取 Token
func (t *tokenInfo) GetToken() string {
	return t.Token
}

// GetTokenType 获取 TokenType
func (t *tokenInfo) GetTokenType() string {
	return t.Type
}

// GetExpiresAt 获取过期时间
func (t *tokenInfo) GetExpiresAt() int64 {
	return t.ExpiresAt
}

// EncodeToJSON JSON 编码
func (t *tokenInfo) EncodeToJSON() ([]byte, error) {
	return json.Marshal(t)
}
