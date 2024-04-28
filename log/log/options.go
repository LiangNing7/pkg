package log

import (
	"github.com/spf13/pflag"
	"go.uber.org/zap/zapcore"
)

type Options struct {
	// DisableCaller 指定是否在日志中包括调用方信息
	DisableCaller bool `json:"disable-caller,omitempty" mapstructure:"disable-caller"`
	// DisableStacktrace 指定是否为处于或高于紧急级别的所有消息记录堆栈跟踪
	DisableStacktrace bool `json:"disable-stacktrace,omitempty" mapstructure:"disable-stacktrace"`
	// EnableColor 指定是否输出彩色日志
	EnableColor bool `json:"enable-color" mapstructure:"enable-color"`
	// Level 指定最低日志级别。有效值为：debug、info、warn、error、dpanic、panic 和 fatal
	Level string `json:"level,omitempty" mapstructure:"level"`
	// Format 指定日志输出格式。有效值为：console和json
	Format string `json:"format,omitempty" mapstructure:"format"`
	// OutputPaths 指定日志的输出路径
	OutputPaths []string `json:"output-paths,omitempty" mapstructure:"output-paths"`
}

// NewOptions 创建一个具有默认值的 Options
func NewOptions() *Options {
	return &Options{
		Level:       zapcore.InfoLevel.String(),
		Format:      "console",
		OutputPaths: []string{"stdout"},
	}
}

// Validate 验证传递给LogsOptions的标志
func (o *Options) Validate() []error {
	errs := []error{}

	return errs
}

// AddFlags 为 Options 添加命令行标志
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Level, "log.level", o.Level, "Minimum log output `Level`.")
	fs.BoolVar(&o.DisableCaller, "log.disable-caller", o.DisableCaller, "Disable output of caller information in the log.")
	fs.BoolVar(&o.DisableStacktrace, "log.disable-stacktrace", o.DisableStacktrace, ""+
		"Disable the log to record a stack trace for all messages at or above panic level.")
	fs.BoolVar(&o.EnableColor, "log.enable-color", o.EnableColor, "Enable output ansi colors in plain format logs.")
	fs.StringVar(&o.Format, "log.format", o.Format, "Log output `FORMAT`, support plain or json format.")
	fs.StringSliceVar(&o.OutputPaths, "log.output-paths", o.OutputPaths, "Output paths of log.")
}
