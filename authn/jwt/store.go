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
