package log

import (
	"io"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Level         string        `mapstructure:"level,omitempty" json:"level,omitempty"`
	FileConfig    FileLogConfig `mapstructure:"file_config,omitempty" json:"file_config,omitempty"`
	EnableConsole bool          `mapstructure:"enable_console,omitempty" json:"enable_console,omitempty"`
	EnableColor   bool          `mapstructure:"enable_color,omitempty" json:"enable_color,omitempty"`
	EnableSampler bool          `mapstructure:"enable_sampler,omitempty" json:"enable_sampler,omitempty"`
}

type FileLogConfig struct {
	FileName   string `mapstructure:"file_name,omitempty" json:"file_name,omitempty"`
	Enable     bool   `mapstructure:"enable,omitempty" json:"enable,omitempty"`
	Maxsize    int    `mapstructure:"maxsize,omitempty" json:"maxsize,omitempty"`
	MaxBackups int    `mapstructure:"max_backups,omitempty" json:"max_backups,omitempty"`
	MaxAge     int    `mapstructure:"max_age,omitempty" json:"max_age,omitempty"`
	Compress   bool   `mapstructure:"compress,omitempty" json:"compress,omitempty"`
}

// 远程日志存储

func GetService() Service {
	return service
}

func GetLogger() *zap.Logger {
	return service.GetLogger()
}

func SetLevel(level string) {
	service.SetLevel(level)
}

func InitLogger(force bool) {
	service.InitLogger(force)
}

func ResetLogger(fields ...zap.Field) {
	service.ResetLogger(fields...)
}

func LoadConfig() {
	service.LoadConfig()
}

func RegisterAccept(logWrite chan<- []byte) int64 {
	return service.RegisterAccept(logWrite)
}

func UnRegisterAccept(id int64) {
	service.UnRegisterAccept(id)
}

func PrintLog(write io.Writer) {
	service.PrintLog(write)
}

func GetRecentLogs(limit int) [][]byte {
	impl, ok := service.(*serviceImpl)
	if !ok {
		return nil
	}
	return impl.GetRecentLogs(limit)
}

// SubscribeLogs 返回旁路日志通道，cancel 只做注销，不会 close 通道。
func SubscribeLogs(bufferSize int, replay int) (<-chan []byte, func()) {
	return SubscribeLogsWithLevel(bufferSize, replay, "")
}

// SubscribeLogsWithLevel 支持按最小级别过滤旁路日志，level 为空表示不过滤。
func SubscribeLogsWithLevel(bufferSize int, replay int, level string) (<-chan []byte, func()) {
	if bufferSize <= 0 {
		bufferSize = 128
	}
	minLevel, hasLevel := parseMinLevel(level)
	rawCh := make(chan []byte, bufferSize)
	out := make(chan []byte, bufferSize)
	id := RegisterAccept(rawCh)
	stopCh := make(chan struct{})
	var once sync.Once
	go func() {
		defer close(out)
		for {
			select {
			case <-stopCh:
				return
			case line := <-rawCh:
				if shouldPassByLevel(line, minLevel, hasLevel) {
					select {
					case out <- line:
					default:
					}
				}
			}
		}
	}()
	if replay > 0 {
		// 先回放最近日志，便于调试端接入后立即看到上下文。
		for _, line := range GetRecentLogs(replay) {
			if !shouldPassByLevel(line, minLevel, hasLevel) {
				continue
			}
			select {
			case out <- line:
			default:
				// 旁路消费慢时不阻塞主链路，直接丢弃回放数据。
			}
		}
	}
	cancel := func() {
		once.Do(func() {
			UnRegisterAccept(id)
			close(stopCh)
		})
	}
	return out, cancel
}

func parseMinLevel(level string) (zapcore.Level, bool) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "":
		return zapcore.InfoLevel, false
	case "debug":
		return zapcore.DebugLevel, true
	case "info":
		return zapcore.InfoLevel, true
	case "warn", "warning":
		return zapcore.WarnLevel, true
	case "error":
		return zapcore.ErrorLevel, true
	case "dpanic":
		return zapcore.DPanicLevel, true
	case "panic":
		return zapcore.PanicLevel, true
	case "fatal":
		return zapcore.FatalLevel, true
	default:
		return zapcore.InfoLevel, false
	}
}

func shouldPassByLevel(line []byte, minLevel zapcore.Level, hasLevel bool) bool {
	if !hasLevel {
		return true
	}
	lvl, ok := parseLevelFromConsoleLine(line)
	if !ok {
		return true
	}
	return lvl >= minLevel
}

func parseLevelFromConsoleLine(line []byte) (zapcore.Level, bool) {
	parts := strings.SplitN(strings.TrimSpace(string(line)), "\t", 4)
	if len(parts) < 2 {
		return zapcore.InfoLevel, false
	}
	return parseMinLevel(parts[1])
}

var TimeLocation = time.FixedZone("CST", 8*3600)

func newEncodeConfig() zapcore.EncoderConfig {
	callerKey := "caller"
	stacktraceKey := "stacktrace"
	if HiddenCall {
		callerKey = ""
		stacktraceKey = ""
	}
	return zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     callerKey,
		MessageKey:    "msg",
		StacktraceKey: stacktraceKey,
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.In(TimeLocation).Format("2006-01-02T15:04:05.000"))
		}, // zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 径编码器
	}
}

func Debug(msg string, fields ...zap.Field) {
	service.SendLog(zap.DebugLevel, msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	service.SendLog(zap.InfoLevel, msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	service.SendLog(zap.WarnLevel, msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	service.SendLog(zap.ErrorLevel, msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	service.SendLog(zap.PanicLevel, msg, fields...)
}

func DPanic(msg string, fields ...zap.Field) {
	service.SendLog(zap.DPanicLevel, msg, fields...)
}

func SendLog(level zapcore.Level, msg string, fields ...zap.Field) {
	service.SendLog(level, msg, fields...)
}

// func Fatal(msg string, fields ...zap.Field) {
// 	service.SendLog(zap.FatalLevel, msg, fields...)
// }
