package i18n

import (
	"embed"
	"golang.org/x/text/language"
)

// Options 定义 i18n 的选项配置.
type Options struct {
	format   string       // 文件格式.
	language language.Tag // 语言.
	files    []string     // 文件列表.
	fs       embed.FS     // 嵌入文件系统.
}

// WithFormat 设置 i18n 的格式.
func WithFormat(format string) func(*Options) {
	return func(options *Options) {
		if format != "" {
			getOptionsOrSetDefault(options).format = format
		}
	}
}

// WithLanguage 设置要使用的语言.
func WithLanguage(lang language.Tag) func(*Options) {
	return func(options *Options) {
		if lang.String() != "und" {
			getOptionsOrSetDefault(options).language = lang
		}
	}
}

// WithFile 添加要加载的 i18n 文件.
func WithFile(f string) func(*Options) {
	return func(options *Options) {
		if f != "" {
			getOptionsOrSetDefault(options).files = append(getOptionsOrSetDefault(options).files, f)
		}
	}
}

// WithFS 设置嵌入文件系统.
func WithFS(fs embed.FS) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).fs = fs
	}
}

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
