package jwt

import (
	"context"
	"github.com/LiangNing7/onex/pkg/authn"
	"github.com/LiangNing7/onex/pkg/i18n"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/golang-jwt/jwt/v4"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"time"
)

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

// JWTAuth implement the authn.Authenticator interface.
type JWTAuth struct {
	opts  *options
	store Storer
}

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

// 下面实现 authn.Authenticator interface 的方法: Sign Destroy ParseClaims Release

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
