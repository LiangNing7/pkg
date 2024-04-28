package log

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// loggerKeyForGin 定义在 Gin 中上下文存储 Logger 的键
const loggerKeyForGin = "OneXLogger"

// contextKey 定义在 `context.Context`中查找 `Logger`的类型
type contextKey struct{}

// WithContext returns a copy of context in which the log value is set.
func WithContext(ctx context.Context, keyvals ...any) context.Context {
	// 如果在上下文中找到了Logger，则使用其 WithContext方法
	if l := FromContext(ctx); l != nil {
		return l.(*zapLogger).WithContext(ctx, keyvals...)
	}

	// 否则，使用全局日志记录器的WithContext方法
	return std.WithContext(ctx, keyvals...)
}

func (l *zapLogger) WithContext(ctx context.Context, keyvals ...any) context.Context {
	// 定义一个函数with，用于在上下文中设置Logger
	with := func(l Logger) context.Context {
		return context.WithValue(ctx, contextKey{}, l)
	}

	// 如果上下文是 Gin 上下文，则使用 Gin 的 Set 方法设置Logger
	if c, ok := ctx.(*gin.Context); ok {
		with = func(l Logger) context.Context {
			c.Set(loggerKeyForGin, l)
			return c
		}
	}

	// 如果没有提供键值对，则返回设置了日志记录器的上下文。
	keylen := len(keyvals)
	if keylen == 0 || keylen%2 != 0 {
		return with(l)
	}

	// 如果提供了键值对，使用提供的键值对创建zap.Field切片
	data := make([]zap.Field, 0, (keylen/2)+1)
	for i := 0; i < keylen; i += 2 {
		data = append(data, zap.Any(fmt.Sprint(keyvals[i]), keyvals[i+1]))
	}

	return with(l.With(data...))
}

// FromContext 从上下文中检索具有预定义值的日志记录器。
func FromContext(ctx context.Context, keyvals ...any) Logger {
	// 根据上下文类型确定键。
	var key any = contextKey{}
	if _, ok := ctx.(*gin.Context); ok {
		key = loggerKeyForGin
	}

	// 将Logger初始化为Logger。
	var log Logger = std

	// 如果可用，则从上下文中检索Logger。
	if ctx != nil {
		if logger, ok := ctx.Value(key).(Logger); ok {
			log = logger
		}
	}

	// 如果没有提供额外的键值对，则返回检索到的Logger。
	keylen := len(keyvals)
	if keylen == 0 || keylen%2 != 0 {
		return log
	}

	// 否则，使用提供的键值对创建 zap.Field 切片。
	data := make([]zap.Field, 0, (keylen/2)+1)
	for i := 0; i < keylen; i += 2 {
		data = append(data, zap.Any(fmt.Sprint(keyvals[i]), keyvals[i+1]))
	}

	// 返回设置了附加字段的日志记录器。
	return log.With(data...)
}

func C(ctx context.Context) Logger {
	// 返回从上下文中检索到的Logger，-1表示跳过当前调用者
	return FromContext(ctx).AddCallerSkip(-1)
}
