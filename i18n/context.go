// 该文件用于在上下文中传递和获取 I18n 实例的函数，在函数调用链中共享 I18n 实例比较有用

package i18n

import "context"

type translator struct{}

// NewContext 创建一个新的 Context，函数接受一个现有的 Context 和一个 I18n 实例作为参数，
// 然后将 I18n 实例存储在 Context。这样，可以将 I18n 实例传递给函数调用链中的其他函数，
// 使它们能够在需要时访问到 I18n 实例
func NewContext(ctx context.Context, i *I18n) context.Context {
	return context.WithValue(ctx, translator{}, i)
}

// FromContext 用于从 Context 中获取 I18n 实例
// 它接受一个 Context 作为参数，并尝试从该上下文中提取 I18n 实例
// 如果成功提取到了 I18n 实例，则返回该实例；否则，会创建一个新的 I18n 实例并返回
func FromContext(ctx context.Context) *I18n {
	if i, ok := ctx.Value(translator{}).(*I18n); ok {
		return i
	}
	return New()
}
