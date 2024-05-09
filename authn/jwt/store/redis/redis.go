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
