# [authn](https://github.com/LiangNing7/pkg/tree/main/authn)

> 用于 `JWT` 的实现

# `authn.go`

定义了如下接口与函数：

```go
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
```

# jwt

> 包装了`IToken`与`Authenticator`接口的实现

## IToken

`token.go`：

```go
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
```

## Authenticator

常量与变量定义：

```go
const (
	// reason 保存错误原因.
	reason string = "Unauthorized"

	// defaultKey 保证用于签署 jwt 令牌的默认密钥.
	defaultKey = "onex(#)666"
)

// 定义错误类型
var (
	// ErrTokenInvalid 表示令牌无效
	ErrTokenInvalid = errors.Unauthorized(reason, "Token is invalid")
	//ErrTokenExpired 表示令牌已过期
	ErrTokenExpired = errors.Unauthorized(reason, "Token is expired")
	// ErrTokenParseFail 表示解析令牌失败
	ErrTokenParseFail = errors.Unauthorized(reason, "Fail to parse token")
	// ErrUnSupportSigningMethod 表示不支持的签名方法
	ErrUnSupportSigningMethod = errors.Unauthorized(reason, "Wrong signing method")
	// ErrSignTokenFailed 表示签署令牌失败
	ErrSignTokenFailed = errors.Unauthorized(reason, "Failed to sign token")
)

// 定义 I18n 的消息
var (
	MessageTokenInvalid           = &goi18n.Message{ID: "jwt.token.invalid", Other: ErrTokenInvalid.Error()}
	MessageTokenExpired           = &goi18n.Message{ID: "jwt.token.expired", Other: ErrTokenExpired.Error()}
	MessageTokenParseFail         = &goi18n.Message{ID: "jwt.token.parse.failed", Other: ErrTokenParseFail.Error()}
	MessageUnSupportSigningMethod = &goi18n.Message{ID: "jwt.token.signing.method", Other: ErrUnSupportSigningMethod.Error()}
	MessageSignTokenFailed        = &goi18n.Message{ID: "jwt.token.sign.failed", Other: ErrSignTokenFailed.Error()}
)
```

`JWT` 的配置：

```go
// 定义 JWT 的配置
type options struct {
	signingMethod jwt.SigningMethod // 签名算法
	signingKey    any               //签名密钥
	keyfunc       jwt.Keyfunc       // 密钥验证回调函数
	issuer        string            // 签发者
	expired       time.Duration     // 过期时间
	tokenType     string            // 令牌类型
	tokenHeader   map[string]any    // 令牌头部信息
}
```

定义默认配置：

```go
// 定义默认配置
var defaultOptions = options{
	tokenType:     "Bearer",               // 设置 token 类型为 Bearer
	expired:       2 * time.Hour,          // 过期时间为 2 小时
	signingMethod: jwt.SigningMethodHS256, // 设置签名算法为 HS256
	signingKey:    []byte(defaultKey),     // 设置默认 Key
	keyfunc: func(t *jwt.Token) (any, error) {
		// 检查 token 的签名方法是否是 HMAC
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			// 若不是 HMAC，返回错误信息 ErrTokenInvalid
			return nil, ErrTokenInvalid
		}
		// 返回默认的密钥
		return []byte(defaultKey), nil
	},
}
```

定义一些函数用于选项模式进行配置：

```go
// Option 定义配置函数，用于选项模式
type Option func(*options)

// WithSigningMethod 设置签名方法。
func WithSigningMethod(method jwt.SigningMethod) Option {
	return func(o *options) {
		o.signingMethod = method
	}
}

// WithIssuer 设置令牌签发者。
func WithIssuer(issuer string) Option {
	return func(o *options) {
		o.issuer = issuer
	}
}

// WithSigningKey 设置签名密钥。
func WithSigningKey(key any) Option {
	return func(o *options) {
		o.signingKey = key
	}
}

// WithKeyfunc 设置验证密钥的回调函数。
func WithKeyfunc(keyFunc jwt.Keyfunc) Option {
	return func(o *options) {
		o.keyfunc = keyFunc
	}
}

// WithExpired 设置令牌过期时间（默认 2 小时）。
func WithExpired(expired time.Duration) Option {
	return func(o *options) {
		o.expired = expired
	}
}

// WithTokenHeader 设置客户端的自定义 tokenHeader。
func WithTokenHeader(header map[string]any) Option {
	return func(o *options) {
		o.tokenHeader = header
	}
}
```

定义`JWTAuth`，用于实现`authn.Authenticator` 接口：

```go
// JWTAuth implement the authn.Authenticator interface.
type JWTAuth struct {
	opts  *options
	store Storer
}
```

使用`New`方法可以创建一个`JWTAuth`实例：

```go
// New 创建一个新的 JWTAuth 实例
func New(store Storer, opts ...Option) *JWTAuth {
	// 使用选项模式进行配置
	o := defaultOptions
	for _, opt := range opts {
		opt(&o)
	}
	// 返回 JWTAuth 实例
	return &JWTAuth{opts: &o, store: store}
}
```

实现`authn.Authenticator`接口需要的函数以及一些辅助函数：

```go
// Sign 用于生成一个新的 Token
func (a *JWTAuth) Sign(ctx context.Context, userID string) (authn.IToken, error) {
	// 获取当前时间
	now := time.Now()

	// 计算令牌过期时间
	expiresAt := now.Add(a.opts.expired)

	// 创建新的令牌
	token := jwt.NewWithClaims(a.opts.signingMethod, &jwt.RegisteredClaims{
		// Issuer = iss,令牌颁发者。它表示该令牌是由谁创建的
		Issuer: a.opts.issuer,
		// IssuedAt = iat,令牌颁发时的时间戳。它表示令牌是何时被创建的
		IssuedAt: jwt.NewNumericDate(now),
		// ExpiresAt = exp,令牌的过期时间戳。它表示令牌将在何时过期
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		// NotBefore = nbf,令牌的生效时的时间戳。它表示令牌从什么时候开始生效
		NotBefore: jwt.NewNumericDate(now),
		// Subject = sub,令牌的主体。它表示该令牌是关于谁的
		Subject: userID,
	})

	// 添加 tokenHeader
	if a.opts.tokenHeader != nil {
		for k, v := range a.opts.tokenHeader {
			token.Header[k] = v
		}
	}

	// 使用签名密钥对令牌进行签名
	refreshToken, err := token.SignedString(a.opts.signingKey)
	if err != nil {
		// 签名失败，返回错误信息
		return nil, errors.Unauthorized(reason, i18n.FromContext(ctx).LocalizeT(MessageSignTokenFailed))
	}

	// 创建 tokenInfo
	tokenInfo := &tokenInfo{
		// 设置令牌过期时间戳
		ExpiresAt: expiresAt.Unix(),
		// 设置令牌类型
		Type: a.opts.tokenType,
		// 设置令牌内容
		Token: refreshToken,
	}
	return tokenInfo, nil
}

// parseToken 用于解析输入的 refreshToken
func (a *JWTAuth) parseToken(ctx context.Context, refreshToken string) (*jwt.RegisteredClaims, error) {
	// 使用提供的 keyfunc 解析令牌
	token, err := jwt.ParseWithClaims(refreshToken, &jwt.RegisteredClaims{}, a.opts.keyfunc)
	if err != nil {
		// 解析错误
		ve, ok := err.(*jwt.ValidationError)
		if !ok {
			// 如果错误不是 ValidationError 类型，则返回未授权的错误
			return nil, errors.Unauthorized(reason, err.Error())
		}
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			// 令牌格式错误
			return nil, errors.Unauthorized(reason, i18n.FromContext(ctx).LocalizeT(MessageTokenInvalid))
		}
		if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			// 令牌已经过期或尚未生效
			return nil, errors.Unauthorized(reason, i18n.FromContext(ctx).LocalizeT(MessageTokenExpired))
		}
		// 其他解析错误
		return nil, errors.Unauthorized(reason, i18n.FromContext(ctx).LocalizeT(MessageTokenParseFail))
	}

	// 验证令牌是否有效
	if !token.Valid {
		return nil, errors.Unauthorized(reason, i18n.FromContext(ctx).LocalizeT(MessageTokenInvalid))
	}

	// 检查签名算法是否与配置一致
	if token.Method != a.opts.signingMethod {
		return nil, errors.Unauthorized(reason, i18n.FromContext(ctx).LocalizeT(MessageUnSupportSigningMethod))
	}
	return token.Claims.(*jwt.RegisteredClaims), nil
}

// callStore 执行传入的存储函数
func (a *JWTAuth) callStore(fn func(Storer) error) error {
	// 检查存储是否存在，如果存在则执行传入的函数 fn
	if store := a.store; store != nil {
		return fn(store)
	}
	// 如果存储不存在，则返回 nil
	return nil
}

// Destroy 用于销毁令牌
func (a *JWTAuth) Destroy(ctx context.Context, refreshToken string) error {
	// 解析令牌声明
	claims, err := a.parseToken(ctx, refreshToken)
	if err != nil {
		return err
	}

	// 如果设置了 storage，将未过期的令牌放入
	store := func(store Storer) error {
		// 设置令牌剩余时间
		expired := time.Until(claims.ExpiresAt.Time)
		// 将令牌放入Store
		return store.Set(ctx, refreshToken, expired)
	}
	// 调用存储函数
	return a.callStore(store)
}

// ParseClaims 解析令牌并返回声明
func (a *JWTAuth) ParseClaims(ctx context.Context, refreshToken string) (*jwt.RegisteredClaims, error) {
	// 如果令牌为空，则返回 TokenInvalid 错误
	if refreshToken == "" {
		return nil, errors.Unauthorized(reason, i18n.FromContext(ctx).LocalizeT(MessageTokenInvalid))
	}
	// 解析令牌声明
	claims, err := a.parseToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	// 检查存储中是否存在该令牌
	store := func(store Storer) error {
		exists, err := store.Check(ctx, refreshToken)
		if err != nil {
			return err
		}
		// 如果存在令牌，则返回未授权的错误，【因为销毁令牌是放入存储中】
		if exists {
			return errors.Unauthorized(reason, i18n.FromContext(ctx).LocalizeT(MessageTokenInvalid))
		}
		return nil
	}
	// 执行调用函数
	if err := a.callStore(store); err != nil {
		return nil, err
	}
	// 返回解析后的声明
	return claims, nil
}

// Release 用于释放请求的资源
func (a *JWTAuth) Release() error {
	// 调用存储的 Close 方法释放资源
	return a.callStore(func(store Storer) error {
		return store.Close()
	})
}
```

## 由于JWT需要进行存储，则包装store

`Store.go`：定义了一些可能用到的方法

```go
package jwt

import (
	"context"
	"time"
)

// Storer token 存储接口.
type Storer interface {
	// Set 存储令牌数据并指定过期时间
	Set(ctx context.Context, assessToken string, expirationTime time.Duration) error
	// Delete 从存储中删除令牌数据
	Delete(ctx context.Context, assessToken string) (bool, error)
	// Check 检查令牌是否存在
	Check(ctx context.Context, assessToken string) (bool, error)
	// Close 关闭存储
	Close() error
}
```

`store/redis/redis.go`：实现需要用到的存储方法：

```go
package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// Config 包含了必要的 Redis 配置选项
type Config struct {
	Addr      string // 地址
	Username  string // 用户名
	Password  string // 密码
	Database  int    // 数据库编号
	KeyPrefix string // 存储键的前缀
}

// Store 用于实现 store.Storer 接口
type Store struct {
	cli    *redis.Client // redis 客户端
	prefix string        // 前缀
}

// NewStore 根据 Config 创建一个 *Store 实例
func NewStore(cfg Config) *Store {
	cli := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		DB:       cfg.Database,
		Username: cfg.Username,
		Password: cfg.Password,
	})
	return &Store{cli: cli, prefix: cfg.KeyPrefix}
}

// wrapperKey 用于构建 Redis 中的键名
func (s *Store) wrapperKey(key string) string {
	return fmt.Sprintf("%s%s", s.prefix, key)
}

// Set 调用 Redis 设置具有过期时间的键值对
// 键的格式为 <prefix><accessToken>
func (s *Store) Set(ctx context.Context, accessToken string, expiration time.Duration) error {
	cmd := s.cli.Set(ctx, s.wrapperKey(accessToken), "1", expiration)
	return cmd.Err()
}

// Delete 删除 Redis 中指定的 JWT 令牌
func (s *Store) Delete(ctx context.Context, accessToken string) (bool, error) {
	cmd := s.cli.Del(ctx, s.wrapperKey(accessToken))
	if err := cmd.Err(); err != nil {
		return false, err
	}
	return cmd.Val() > 0, nil
}

// Check 检查 Redis 中指定的 JWT 令牌是否存在
func (s *Store) Check(ctx context.Context, accessToken string) (bool, error) {
	cmd := s.cli.Exists(ctx, s.wrapperKey(accessToken))
	if err := cmd.Err(); err != nil {
		return false, err
	}
	return cmd.Val() > 0, nil
}

// Close 用于关闭 Redis client
func (s *Store) Close() error {
	return s.cli.Close()
}
```

