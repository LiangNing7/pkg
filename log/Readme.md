# [日志包](https://github.com/LiangNing7/pkg/tree/main/log)

> 该 log 包主要基于`zap`包，并在其中集成了`gorm`和`kratos`的`log`

# log

定义字段类型，与zap日志库中的字段类型一致

```go
type Field = zapcore.Field
```



定义Logger，其为 onex 项目的日志接口，该接口只包含了支持的日志记录方法

```go
type Logger interface {
	Debugf(format string, args ...any)
	Debugw(msg string, keyvals ...any)
	Infof(format string, args ...any)
	Infow(msg string, keyvals ...any)
	Warnf(format string, args ...any)
	Warnw(msg string, keyvals ...any)
	Errorf(format string, args ...any)
	Errorw(err error, msg string, keyvals ...any)
	Panicf(format string, args ...any)
	Panicw(msg string, keyvals ...any)
	Fatalf(format string, args ...any)
	Fatalw(msg string, keyvals ...any)
	With(fields ...Field) Logger
	AddCallerSkip(skip int) Logger
	Sync()

	// integrate other loggers
	krtlog.Logger
	gormlogger.Interface
}
```



`zapLogger`是`Logger`接口的具体实现，底层封装了 `zap.Logger`：

```go
type zapLogger struct {
	z    *zap.Logger
	opts *Options
}

// 确保 zapLogger 实现了 Logger 接口. 以下变量赋值，可以使错误在编译期被发现.
var _ Logger = (*zapLogger)(nil)
```



定义全局锁以及默认的全局`Logger`：

```go
var (
	mu sync.Mutex

	// std 定义了默认的全局 Logger.
	std = NewLogger(NewOptions())
)
```



`Init`函数使用指定的选项初始化Logger、`Default`函数返回默认的Logger，以及`Options`函数返回该Logger的配置：

```go
// Init 使用指定的选项初始化 Logger.
func Init(opts *Options) {
	mu.Lock()
	defer mu.Unlock()

	std = NewLogger(opts)
}

// Default 返回默认的 Logger
func Default() Logger {
	return std
}

// Options 返回该 Logger 的配置
func (l *zapLogger) Options() *Options {
	return l.opts
}
```



`NewLogger`函数根据传入的 opts 创建 Logger：

其中 `Options` 如下：

```go 
type Options struct {
    DisableCaller     bool     `json:"disable-caller,omitempty" mapstructure:"disable-caller"`
    DisableStacktrace bool     `json:"disable-stacktrace,omitempty" mapstructure:"disable-stacktrace"`
    EnableColor       bool     `json:"enable-color"       mapstructure:"enable-color"`
    Level             string   `json:"level,omitempty" mapstructure:"level"`
    Format            string   `json:"format,omitempty" mapstructure:"format"`
    OutputPaths       []string `json:"output-paths,omitempty" mapstructure:"output-paths"`
}
```

`NewLogger` 函数如下：

```go 
// NewLogger 根据传入的 opts 创建 Logger.
func NewLogger(opts *Options) *zapLogger {
    // 如果传入的opts为nil，则调用 NewOptions 创建 opts
	if opts == nil {
		opts = NewOptions()
	}

	// 将文本格式的日志级别，例如 info 转换为 zapcore.Level 类型以供后面使用
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(opts.Level)); err != nil {
		// 如果指定了非法的日志级别，则默认使用 info 级别
		zapLevel = zapcore.InfoLevel
	}

	// 创建一个默认的 encoder 配置
	encoderConfig := zap.NewProductionEncoderConfig()
	// 自定义 MessageKey 为 message，message 语义更明确
	encoderConfig.MessageKey = "message"
	// 自定义 TimeKey 为 timestamp，timestamp 语义更明确
	encoderConfig.TimeKey = "timestamp"
	// 指定时间序列化函数，将时间序列化为 `2006-01-02 15:04:05.000` 格式，更易读
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	// 指定 time.Duration 序列化函数，将 time.Duration 序列化为经过的毫秒数的浮点数
	// 毫秒数比默认的秒数更精确
	encoderConfig.EncodeDuration = func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendFloat64(float64(d) / float64(time.Millisecond))
	}
	// when output to local path, with color is forbidden
	if opts.Format == "console" && opts.EnableColor {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 创建构建 zap.Logger 需要的配置
	cfg := &zap.Config{
		// 是否在日志中显示调用日志所在的文件和行号，例如：`"caller":"onex/onex.go:75"`
		DisableCaller: opts.DisableCaller,
		// 是否禁止在 panic 及以上级别打印堆栈信息
		DisableStacktrace: opts.DisableStacktrace,
		// 指定日志级别
		Level: zap.NewAtomicLevelAt(zapLevel),
		// 指定日志显示格式，可选值：console, json
		Encoding:      opts.Format,
		EncoderConfig: encoderConfig,
		// 指定日志输出位置
		OutputPaths: opts.OutputPaths,
		// 设置 zap 内部错误输出位置
		ErrorOutputPaths: []string{"stderr"},
	}

	// 使用 cfg 创建 *zap.Logger 对象
	z, err := cfg.Build(zap.AddStacktrace(zapcore.PanicLevel), zap.AddCallerSkip(2))
	if err != nil {
		panic(err)
	}
	logger := &zapLogger{z: z, opts: opts}

	// 把标准库的 log.Logger 的 info 级别的输出重定向到 zap.Logger
	zap.RedirectStdLog(z)

	return logger
}
```



实现Logger的各种方法：

```go 
// Debugf 输出 debug 级别的日志.
func Debugf(format string, args ...any) {
	std.Debugf(format, args...)
}

func (l *zapLogger) Debugf(format string, args ...any) {
	l.z.Sugar().Debugf(format, args...)
}

// Debugw 输出 debug 级别的日志.
func Debugw(msg string, keyvals ...any) {
	std.Debugw(msg, keyvals...)
}

func (l *zapLogger) Debugw(msg string, keyvals ...any) {
	l.z.Sugar().Debugw(msg, keyvals...)
}

// Infof 输出 info 级别的日志.
func Infof(format string, args ...any) {
	std.Infof(format, args...)
}

func (l *zapLogger) Infof(msg string, keyvals ...any) {
	l.z.Sugar().Infof(msg, keyvals...)
}

// Infow 输出 info 级别的日志.
func Infow(msg string, keyvals ...any) {
	std.Infow(msg, keyvals...)
}

func (l *zapLogger) Infow(msg string, keyvals ...any) {
	l.z.Sugar().Infow(msg, keyvals...)
}

// Warnf 输出 warning 级别的日志.
func Warnf(format string, args ...any) {
	std.Warnf(format, args...)
}

func (l *zapLogger) Warnf(format string, args ...any) {
	l.z.Sugar().Warnf(format, args...)
}

// Warnw 输出 warning 级别的日志.
func Warnw(msg string, keyvals ...any) {
	std.Warnw(msg, keyvals...)
}

func (l *zapLogger) Warnw(msg string, keyvals ...any) {
	l.z.Sugar().Warnw(msg, keyvals...)
}

// Errorf 输出 error 级别的日志.
func Errorf(format string, args ...any) {
	std.Errorf(format, args...)
}

func (l *zapLogger) Errorf(format string, args ...any) {
	l.z.Sugar().Errorf(format, args...)
}

// Errorw 输出 error 级别的日志.
func Errorw(err error, msg string, keyvals ...any) {
	std.Errorw(err, msg, keyvals...)
}

func (l *zapLogger) Errorw(err error, msg string, keyvals ...any) {
	l.z.Sugar().Errorw(msg, append(keyvals, "err", err)...)
}

// Panicf 输出 panic 级别的日志.
func Panicf(format string, args ...any) {
	std.Panicf(format, args...)
}

func (l *zapLogger) Panicf(format string, args ...any) {
	l.z.Sugar().Panicf(format, args...)
}

// Panicw 输出 panic 级别的日志.
func Panicw(msg string, keyvals ...any) {
	std.Panicw(msg, keyvals...)
}

func (l *zapLogger) Panicw(msg string, keyvals ...any) {
	l.z.Sugar().Panicw(msg, keyvals...)
}

// Fatalf 输出 fatal 级别的日志.
func Fatalf(format string, args ...any) {
	std.Fatalf(format, args...)
}

func (l *zapLogger) Fatalf(format string, args ...any) {
	l.z.Sugar().Fatalf(format, args...)
}

// Fatalw 输出 fatal 级别的日志.
func Fatalw(msg string, keyvals ...any) {
	std.Fatalw(msg, keyvals...)
}

func (l *zapLogger) Fatalw(msg string, keyvals ...any) {
	l.z.Sugar().Fatalw(msg, keyvals...)
}

func With(fields ...Field) Logger {
	return std.With(fields...)
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func (l *zapLogger) With(fields ...Field) Logger {
	if len(fields) == 0 {
		return l
	}

	lc := l.clone()
	lc.z = lc.z.With(fields...)
	return lc
}

func AddCallerSkip(skip int) Logger {
	return std.AddCallerSkip(skip)
}

// AddCallerSkip increases the number of callers skipped by caller annotation
// (as enabled by the AddCaller option). When building wrappers around the
// Logger and SugaredLogger, supplying this Option prevents zap from always
// reporting the wrapper code as the caller.
func (l *zapLogger) AddCallerSkip(skip int) Logger {
	lc := l.clone()
	lc.z = lc.z.WithOptions(zap.AddCallerSkip(skip))
	return lc
}

// clone 深度拷贝 zapLogger.
func (l *zapLogger) clone() *zapLogger {
	copied := *l
	return &copied
}

// Sync 调用底层 zap.Logger 的 Sync 方法，将缓存中的日志刷新到磁盘文件中. 主程序需要在退出前调用 Sync.
func Sync() { std.Sync() }

func (l *zapLogger) Sync() {
	_ = l.z.Sync()
}
```



# options

`Options`结构体：

```go 
// Options contains configuration options for logging.
type Options struct {
	// DisableCaller specifies whether to include caller information in the log.
	DisableCaller bool `json:"disable-caller,omitempty" mapstructure:"disable-caller"`
	// DisableStacktrace specifies whether to record a stack trace for all messages at or above panic level.
	DisableStacktrace bool `json:"disable-stacktrace,omitempty" mapstructure:"disable-stacktrace"`
	// EnableColor specifies whether to output colored logs.
	EnableColor bool `json:"enable-color"       mapstructure:"enable-color"`
	// Level specifies the minimum log level. Valid values are: debug, info, warn, error, dpanic, panic, and fatal.
	Level string `json:"level,omitempty" mapstructure:"level"`
	// Format specifies the log output format. Valid values are: console and json.
	Format string `json:"format,omitempty" mapstructure:"format"`
	// OutputPaths specifies the output paths for the logs.
	OutputPaths []string `json:"output-paths,omitempty" mapstructure:"output-paths"`
}
```



`NewOptions`函数：创建默认的`Options`

```go 
// NewOptions creates a new Options object with default values.
func NewOptions() *Options {
	return &Options{
		Level:       zapcore.InfoLevel.String(),
		Format:      "console",
		OutputPaths: []string{"stdout"},
	}
}
```



`Validate`函数：验证传递给LogsOptions的标志：

```go 
// Validate verifies flags passed to LogsOptions.
func (o *Options) Validate() []error {
	errs := []error{}

	return errs
}
```



`AddFlags`：为`Options`添加命令行标志

```go 
// AddFlags adds command line flags for the configuration.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Level, "log.level", o.Level, "Minimum log output `LEVEL`.")
	fs.BoolVar(&o.DisableCaller, "log.disable-caller", o.DisableCaller, "Disable output of caller information in the log.")
	fs.BoolVar(&o.DisableStacktrace, "log.disable-stacktrace", o.DisableStacktrace, ""+
		"Disable the log to record a stack trace for all messages at or above panic level.")
	fs.BoolVar(&o.EnableColor, "log.enable-color", o.EnableColor, "Enable output ansi colors in plain format logs.")
	fs.StringVar(&o.Format, "log.format", o.Format, "Log output `FORMAT`, support plain or json format.")
	fs.StringSliceVar(&o.OutputPaths, "log.output-paths", o.OutputPaths, "Output paths of log.")
}
```



# 集成 gorm log

定义日志格式字符串和慢查询的阈值时间：

```go
// 定义日志格式字符串
var (
	infoStr       = "%s[info] "
	warnStr       = "%s[warn] "
	errStr        = "%s[error] "
	traceStr      = "[%s][%.3fms] [rows:%v] %s"
	traceWarnStr  = "%s %s[%.3fms] [rows:%v] %s"
	traceErrStr   = "%s %s[%.3fms] [rows:%v] %s"
	slowThreshold = 200 * time.Millisecond
)
```



设置日志级别映射表，将字符串表示的日志级别映射为 GORM 中的日志级别常量：

```go
// 日志级别映射表
var levelM = map[string]gormlogger.LogLevel{
	"panic": gormlogger.Silent,
	"error": gormlogger.Error,
	"warn":  gormlogger.Warn,
	"info":  gormlogger.Info,
}
```



`LogMode`函数，用于设置日志的输出级别：

```go
func (l *zapLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	opts := *l.opts
	switch {
	case level <= gormlogger.Silent:
		opts.Level = "panic"
	case level <= gormlogger.Error:
		opts.Level = "error"
	case level <= gormlogger.Warn:
		opts.Level = "warn"
	case level <= gormlogger.Info:
		opts.Level = "info"
	default:
	}

	return NewLogger(&opts)
}
```



输出不同级别日志的方法，分别用于输出信息，警告和错误日志。它们会根据传入的消息和额外的参数进行格式化，并调用底层的日志库输出日志：

```go
func (l *zapLogger) Info(ctx context.Context, msg string, keyvals ...any) {
	l.z.Sugar().Infof(infoStr+msg, append([]any{fileWithLineNum()}, keyvals...)...)
}

func (l *zapLogger) Warn(ctx context.Context, msg string, keyvals ...any) {
	l.z.Sugar().Warnf(warnStr+msg, append([]any{fileWithLineNum()}, keyvals...)...)
}

func (l *zapLogger) Error(ctx context.Context, msg string, keyvals ...any) {
	l.z.Sugar().Errorf(errStr+msg, append([]any{fileWithLineNum()}, keyvals...)...)
}
```



`Trace`函数，根据执行时间和日志级别来决定是否输出追踪日志：

```go
// Trace 方法用于记录数据库操作的跟踪信息，并根据日志级别进行相应的处理。
func (l *zapLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// 检查日志级别是否低于静默级别，如果是，则不记录日志，直接返回
	if levelM[l.opts.Level] <= gormlogger.Silent {
		return
	}

	// 计算操作所花费的时间
	elapsed := time.Since(begin)
	switch {
	case err != nil && levelM[l.opts.Level] >= gormlogger.Error:
		// 如果有错误并且日志级别高于或等于 Error 级别，则记录错误日志
		sql, rows := fc()
		if rows == -1 {
			l.z.Sugar().Errorf(traceErrStr, fileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.z.Sugar().Errorf(traceErrStr, fileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > slowThreshold && slowThreshold != 0 && levelM[l.opts.Level] >= gormlogger.Warn:
		// 如果操作时间超过慢查询阈值且日志级别高于或等于 Warn 级别，则记录慢查询日志
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", slowThreshold)
		if rows == -1 {
			l.z.Sugar().Warnf(traceWarnStr, fileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.z.Sugar().Warnf(traceWarnStr, fileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case levelM[l.opts.Level] >= gormlogger.Info:
		// 如果日志级别高于或等于 Info 级别，则记录普通信息日志
		sql, rows := fc()
		if rows == -1 {
			l.z.Sugar().Infof(traceStr, fileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.z.Sugar().Infof(traceStr, fileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}
```



`fileWithLineNum`函数，用于获取调用此函数的文件名和行号，用于输出日志时显示源代码位置：

```go
func fileWithLineNum() string {
	for i := 4; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)

		// 如果成功获取了文件名和行号，并且不是测试文件，则返回文件名和行号
		if ok && !strings.HasSuffix(file, "_test.go") {
			dir, f := filepath.Split(file)

			return filepath.Join(filepath.Base(dir), f) + ":" + strconv.FormatInt(int64(line), 10)
		}
	}

	return ""
}
```

# 集成 krator log

定义一个`KratosLogger`接口，该接口包含一个Log方法，用于输出日志：

```go
type KratosLogger interface {
	// Log implements is used to github.com/go-kratos/kratos/v2/log.Logger interface.
	Log(level krtlog.Level, keyvals ...any) error
}
```



`Log`函数：

```go
// 接受日志级别和键值对参数[可变列表]
func (l *zapLogger) Log(level krtlog.Level, keyvals ...any) error {
	keylen := len(keyvals)
	if keylen == 0 || keylen%2 != 0 {
		l.z.Warn(fmt.Sprint("Keyvalues must appear in pairs: ", keyvals))
		return nil
	}

	switch level {
	case krtlog.LevelDebug:
		l.z.Sugar().Debugw("", keyvals...)
	case krtlog.LevelInfo:
		l.z.Sugar().Infow("", keyvals...)
	case krtlog.LevelWarn:
		l.z.Sugar().Warnw("", keyvals...)
	case krtlog.LevelError:
		l.z.Sugar().Errorw("", keyvals...)
	case krtlog.LevelFatal:
		l.z.Sugar().Fatalw("", keyvals...)
	}

	return nil
}
```



# context

> 提供了一些函数用于在上下文中存储和获取日志记录器对象，并提供了与Gin框架集成的功能

定义在 Gin 中上下文存储 Logger 的键

```go 
const loggerKeyForGin = "OneXLogger"
```



定义在 `context.Context`中查找 `Logger`的类型

```go 
type contextKey struct{}
```



`WithContext`函数，返回设置了日志值的上下文的副本：

```go 
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
```



`FromContext`函数，从上下文中检索具有预定义值的Logger：

```go 
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
```



`C`函数，不带额外键值对的 FromContext 的快捷方式：

```go 
func C(ctx context.Context) Logger {
	// 返回从上下文中检索到的Logger，-1表示跳过当前调用者
	return FromContext(ctx).AddCallerSkip(-1)
}
```

