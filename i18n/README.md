# [i18n](https://github.com/LiangNing7/pkg/tree/main/i18n)

> 一个帮助您将 Go 程序翻译成多种语言的 Go 包

# `options.go`

定义 `I18n`  的配置：

```go
// Options 定义 i18n 的选项配置.
type Options struct {
	format   string       // 文件格式
	language language.Tag // 语言
	files    []string     // 文件列表
	fs       embed.FS     // 嵌入文件系统
}
```

使用选项模式构建Options：

```go
// WithFormat 设置 i18n 的格式
func WithFormat(format string) func(*Options) {
	return func(options *Options) {
		if format != "" {
			getOptionsOrSetDefault(options).format = format
		}
	}
}

// WithLanguage 设置要使用的语言。
func WithLanguage(lang language.Tag) func(*Options) {
	return func(options *Options) {
		if lang.String() != "und" {
			getOptionsOrSetDefault(options).language = lang
		}
	}
}

// WithFile 添加要加载的 i18n 文件。
func WithFile(f string) func(*Options) {
	return func(options *Options) {
		if f != "" {
			getOptionsOrSetDefault(options).files = append(getOptionsOrSetDefault(options).files, f)
		}
	}
}

// WithFS 设置嵌入文件系统。
func WithFS(fs embed.FS) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).fs = fs
	}
}
```

返回已有的选项或者默认选项：

```go
// getOptionsOrSetDefault 返回已有的选项或设置默认选项.
func getOptionsOrSetDefault(options *Options) *Options {
	// 如果配置为空，则返回默认配置.
	if options == nil {
		return &Options{
			format:   "yml",
			language: language.English, //设置默认为英语.
			files:    []string{},
		}
	}
	return options
}
```

# `i18n.go`

## `I18n`的构建

定义`I18n`的配置：

```go
// I18n 用于存储 I18n 的选项和配置
type I18n struct {
	ops       Options         // 配置
	bundle    *i18n.Bundle    // i18n包的 Bundle
	localizer *i18n.Localizer // i18n 包的 Localizer
	lang      language.Tag    // 语言
}
```

创建 `I18n`实例：

```go
// New 使用给定的参数创建一个新的 I18n 结构实例
// 接受一个可变参数的函数选项，用于选项模式的函数配置
// 最后返回 I18n 结构的指针
func New(options ...func(*Options)) (rp *I18n) {
	// 获取配置参数
	ops := getOptionsOrSetDefault(nil)
	// 使用配置函数进行配置
	for _, f := range options {
		f(ops)
	}

	// 创建一个 Bundle 以在应用程序的整个生命周期中使用
	bundle := i18n.NewBundle(ops.language)

	// 创建一个 Localizer 以便于一组首选语言
	localizer := i18n.NewLocalizer(bundle, ops.language.String())

	// 设置 I18n 的格式，可能为 toml,json，默认配置为 yaml
	switch ops.format {
	case "toml":
		bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	case "json":
		bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	default:
		bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	}

	// I18n 的实例
	rp = &I18n{
		ops:       *ops,
		bundle:    bundle,
		localizer: localizer,
		lang:      ops.language,
	}

	// 添加 I18n 的配置文件
	for _, item := range ops.files {
		rp.Add(item)
	}
	// 添加 I18n 的 FS
	rp.AddFS(ops.fs)
	return
}
```

## 功能列表

使用`Select`选择语言：

```go
// Select 选择语言
func (i *I18n) Select(lang language.Tag) *I18n {
	// 如果没有定义【und】，则设置 I18n 的语言为 I18n 配置的语言
	if lang.String() == "und" {
		lang = i.ops.language
	}
	return &I18n{
		ops:    i.ops,
		bundle: i.bundle,
		// 需要重新定义 localizer
		localizer: i18n.NewLocalizer(i.bundle, lang.String()),
		lang:      lang,
	}
}
```

使用`Language`获取当前语言：

```go
// Language 获取当前语言
func (i *I18n) Language() language.Tag {
	return i.lang
}
```

本地化给定的消息：

```go
// LocalizeT 本地化给定的消息并返回本地化的字符串
// 如果无法翻译，则将消息作为默认消息返回
func (i I18n) LocalizeT(message *i18n.Message) (rp string) {
	// 如果消息为空，则直接返回空
	if message == nil {
		return ""
	}

	// 本地化给定的消息
	var err error
	rp, err = i.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: message,
	})
	if err != nil {
		// 当无法翻译时，使用 ID 作为消息
		rp = message.ID
	}
	return
}

// LocalizeE 是 LocalizeT 方法的包装，将本地化的字符串转换为错误类型并返回
func (i I18n) LocalizeE(message *i18n.Message) error {
	return errors.New(i.LocalizeT(message))
}

// T 本地化具有给定 ID 的消息并返回本地化的字符串
// 它使用 LocalizerT 方法执行翻译
func (i I18n) T(id string) (rp string) {
	return i.LocalizeT(&i18n.Message{ID: id})
}

// E 是对 T 的包装，将本地化字符串转换为错误类型并返回
func (i I18n) E(id string) error {
	return errors.New(i.T(id))
}
```

## 加载翻译的配置文件

```go
// Add 添加语言文件或目录
func (i *I18n) Add(f string) {
	info, err := os.Stat(f) // os.Stat 返回 f 的 fileinfo
	// 如果有错误，则直接返回
	if err != nil {
		return
	}
	// 如果为文件夹，则进行遍历
	if info.IsDir() {
		filepath.Walk(f, func(path string, fi os.FileInfo, errBack error) (err error) {
			// 不是文件夹则添加
			if !fi.IsDir() {
				i.bundle.LoadMessageFile(path)
			}
			return
		})
	} else {
		// 如果是文件，则直接添加
		i.bundle.LoadMessageFile(f)
	}
}

// readFS 用于递归地读取 embed.FS 中的文件
func readFS(fs embed.FS, dir string) (rp []string) {
	rp = make([]string, 0)
	// 使用 fs.ReadDir 方法读取 embed.FS 中指定目录的文件信息
	dirs, err := fs.ReadDir(dir)
	if err != nil {
		return
	}
	// 遍历读取的文件信息
	for _, item := range dirs {
		// 构建文件的完整路径
		name := dir + string(os.PathSeparator) + item.Name()
		// 如果是当前文件夹，则值使用文件名
		if dir == "." {
			name = item.Name()
		}
		// 如果当前项是一个目录，则递归调用 readFS 函数读取其下文件
		if item.IsDir() {
			rp = append(rp, readFS(fs, name)...)
		} else {
			// 如果当前项是一个文件，则将其路径添加到返回的文件列表中。
			rp = append(rp, name)
		}
	}
	return
}

// AddFS 方法用于向 I18n 实例添加 embed.FS 文件。
func (i *I18n) AddFS(fs embed.FS) {
	// 调用 readFS 函数获取 embed.FS 中的文件列表
	files := readFS(fs, ".")
	// 遍历文件列表，将每个文件加载到 i18n 包的 Bundle 中
	for _, name := range files {
		i.bundle.LoadMessageFileFS(fs, name)
	}
}
```

# `Context.go`

> 该文件用于在上下文中传递和获取 `I18n` 实例的函数，在函数调用链中共享 `I18n` 实例比较有用

```go
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
```

# Usage

在`example`文件夹下创建`locales`文件夹：其中包含用于 `I18n` 的文件：`en.yaml、zh.yaml、locales.go`：

其中`en.yaml`如下：

```yaml
no.permission: 'No permission to access this resource'
```

`zh.yaml`如下：

```yaml
no.permission: '无权访问'
```

而`locales.go`如下：

```go
package locales

import "embed"

//go:embed en.yaml zh.yaml
var Locales embed.FS

const (
	NoPermission = "no.permission"
)

```

最后在`example`文件夹中创建`main.go`文件来完成调用：

```go
// 运行文件夹和文件的情况需要进入到 example 文件夹
package main

import (
	"fmt"
	"github.com/LiangNing7/onex/pkg/i18n"
	"github.com/LiangNing7/onex/pkg/i18n/example/locales"
	"golang.org/x/text/language"
)

func main() {
    i := i18n.New()
    // 1. add dir
    //i.Add("./locales")

    // 2. add files
    //i.Add("./locales/en.yaml")
    //i.Add("./locales/zh.yaml")

    // 3. add embed fs
    i.AddFS(locales.Locales)

    fmt.Println(i.T("no.permission"))
    fmt.Println(i.Select(language.Chinese).T("no.permission"))
}
```

