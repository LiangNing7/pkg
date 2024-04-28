package log

import (
	"context"
	"fmt"
	gormlogger "gorm.io/gorm/logger"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	infoStr       = "%s[info] "
	warnStr       = "%s[warn] "
	errStr        = "%s[error] "
	traceStr      = "[%s][%.3fms] [rows:%v] %s"
	traceWarnStr  = "%s %s[%.3fms] [rows:%v] %s"
	traceErrStr   = "%s %s[%.3fms] [rows:%v] %s"
	slowThreshold = 200 * time.Millisecond
)

var levelM = map[string]gormlogger.LogLevel{
	"panic": gormlogger.Silent,
	"error": gormlogger.Error,
	"warn":  gormlogger.Warn,
	"info":  gormlogger.Info,
}

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

func fileWithLineNum() string {
	for i := 4; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)

		if ok && !strings.HasSuffix(file, "_test.go") {
			dir, f := filepath.Split(file)
			return filepath.Join(filepath.Base(dir), f) + ":" + strconv.FormatInt(int64(line), 10)
		}
	}
	return ""
}

func (l *zapLogger) Info(ctx context.Context, msg string, keyvals ...any) {
	l.z.Sugar().Infof(infoStr+msg, append([]any{fileWithLineNum()}, keyvals)...)
}
func (l *zapLogger) Warn(ctx context.Context, msg string, keyvals ...any) {
	l.z.Sugar().Warnf(warnStr+msg, append([]any{fileWithLineNum()}, keyvals)...)
}
func (l *zapLogger) Error(ctx context.Context, msg string, keyvals ...any) {
	l.z.Sugar().Errorf(errStr+msg, append([]any{fileWithLineNum()}, keyvals)...)
}

func (l *zapLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// 检查日志级别是否低于静默级别，如果是，则不记录日志，直接返回
	if levelM[l.opts.Level] <= gormlogger.Silent {
		return
	}

	// 计算操作所花费时间
	elapsed := time.Since(begin)
	switch {
	case err != nil && levelM[l.opts.Level] >= gormlogger.Error:
		// 如果有错误并且日志级别高于或等于 Error 级别，则记录错误日志
		sql, rows := fc()
		if rows == -1 {
			l.z.Sugar().Errorf(traceErrStr, fileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", "sql")
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
