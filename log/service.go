package log

import (
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

// func Fatal(msg string, fields ...zap.Field) {
// 	service.SendLog(zap.FatalLevel, msg, fields...)
// }
