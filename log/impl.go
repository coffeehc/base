package log

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var HiddenCall = false
var disableLog = false

func SetDisableLog(disable bool) {
	service.SetLevel("fatal")
	disableLog = disable
}

func init() {
	service = newService()
	service.InitLogger(true)
}

var DefaultLevel = "info"

type Service interface {
	GetLogger() *zap.Logger
	InitLogger(force bool)
	ResetLogger(fields ...zap.Field)
	SendLog(level zapcore.Level, msg string, fields ...zap.Field)
	LoadConfig()
	SetLevel(level string)
	PrintLog(write io.Writer)

	RegisterAccept(logWrite chan<- []byte) int64
	UnRegisterAccept(id int64)
}

var service Service

func newService() Service {
	defaultConfig := &Config{
		Level: DefaultLevel,
		FileConfig: FileLogConfig{
			FileName:   "./logs/service.log",
			Enable:     false,
			Maxsize:    100,
			MaxBackups: 5,
			MaxAge:     7,
			Compress:   true,
		},
		EnableConsole: true,
		EnableColor:   true,
		EnableSampler: true,
	}
	viper.SetDefault("logger", defaultConfig)
	logFileWrite := &lumberjack.Logger{
		Filename:   defaultConfig.FileConfig.FileName,
		LocalTime:  true,
		MaxSize:    defaultConfig.FileConfig.Maxsize,    // megabytes
		MaxBackups: defaultConfig.FileConfig.MaxBackups, // 最多保留3个备份
		MaxAge:     defaultConfig.FileConfig.MaxAge,     // days
		Compress:   defaultConfig.FileConfig.Compress,   // 是否压缩 disabled by default
	}
	encodeConfig := newEncodeConfig()
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encodeConfig), zapcore.AddSync(os.Stdout), zap.InfoLevel)
	rootLogger := zap.New(zapcore.NewTee(core), zap.AddStacktrace(zapcore.DPanicLevel), zap.AddCaller())
	impl := &serviceImpl{
		level:         zap.NewAtomicLevel(),
		logFileWrite:  logFileWrite,
		rootLogger:    rootLogger,
		conf:          defaultConfig,
		_logWritePipe: new(logWritePipe),
	}
	impl.LoadConfig()
	impl.ResetLogger()
	return impl
}

type serviceImpl struct {
	rootLogger     *zap.Logger
	defaultLogger  *zap.Logger
	internalLogger *zap.Logger
	level          zap.AtomicLevel
	mutex          sync.Mutex
	baseFields     []zap.Field
	configHash     string
	logFileWrite   *lumberjack.Logger
	conf           *Config
	_logWritePipe  *logWritePipe
}

func (impl *serviceImpl) RegisterAccept(logWrite chan<- []byte) int64 {
	return impl._logWritePipe.RegisterAccept(logWrite)
}

func (impl *serviceImpl) UnRegisterAccept(id int64) {
	impl._logWritePipe.UnRegisterAccept(id)
}

func (impl *serviceImpl) PrintLog(write io.Writer) {

}

func (impl *serviceImpl) SetLevel(level string) {
	var logLevel = zap.InfoLevel
	switch strings.ToLower(level) {
	case "debug":
		logLevel = zap.DebugLevel
	case "info":
		logLevel = zap.InfoLevel
	case "error":
		logLevel = zap.ErrorLevel
	case "dPanic":
		logLevel = zap.DPanicLevel
	case "panic":
		logLevel = zap.PanicLevel
	case "fatal":
		logLevel = zap.FatalLevel
	default:
		Warn("无法识别的日志级别", zap.String("level", level))
		return
	}
	// zap.In.Info("设置日志级别", zap.String("level", level))
	//fmt.Println("设置日志级别:" + level)
	impl.level.SetLevel(logLevel)
}

func (impl *serviceImpl) LoadConfig() {
	conf := impl.conf
	err := viper.UnmarshalKey("logger", conf)
	if err != nil {
		impl.rootLogger.Error("解析日志配置失败", zap.Error(err))
		return
	}
	data, _ := json.Marshal(conf)
	hash := md5.New()
	hash.Write(data)
	configHash := hex.EncodeToString(hash.Sum(nil))
	if configHash == impl.configHash {
		return
	}
	if disableLog {
		conf.Level = "fatal"
	}
	impl.SetLevel(conf.Level)
	logCores := make([]zapcore.Core, 0)
	fileLogConfig := conf.FileConfig
	if fileLogConfig.Enable {
		if fileLogConfig.FileName != "" {
			impl.logFileWrite.Filename = fileLogConfig.FileName
		}
		if fileLogConfig.MaxAge > 0 {
			impl.logFileWrite.MaxAge = fileLogConfig.MaxAge
		}
		if fileLogConfig.MaxBackups > 0 {
			impl.logFileWrite.MaxBackups = fileLogConfig.MaxBackups
		}
		if fileLogConfig.Maxsize > 0 {
			impl.logFileWrite.MaxSize = fileLogConfig.Maxsize
		}
		if fileLogConfig.Compress != impl.logFileWrite.Compress {
			impl.logFileWrite.Compress = fileLogConfig.Compress
		}
		impl.logFileWrite.Rotate()
		encoder := zapcore.NewJSONEncoder(newEncodeConfig())
		core := zapcore.NewCore(encoder, zapcore.AddSync(impl.logFileWrite), impl.level)
		if conf.EnableSampler {
			core = zapcore.NewSamplerWithOptions(core, time.Second*5, 100, 10)
		}
		logCores = append(logCores, core)
	}
	if conf.EnableConsole {
		encodeConfig := newEncodeConfig()
		if conf.EnableColor {
			//encodeConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
			encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		}
		core := zapcore.NewCore(zapcore.NewConsoleEncoder(encodeConfig), zapcore.AddSync(os.Stdout), impl.level)
		if conf.EnableSampler {
			core = zapcore.NewSamplerWithOptions(core, time.Second*5, 100, 5)
		}
		logCores = append(logCores, core)
	}
	encodeConfig := newEncodeConfig()
	encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encodeConfig), zapcore.AddSync(impl._logWritePipe), impl.level)
	if conf.EnableSampler {
		core = zapcore.NewSamplerWithOptions(core, time.Second*5, 100, 5)
	}
	logCores = append(logCores, core)
	impl.rootLogger.Sync()
	impl.rootLogger = zap.New(zapcore.NewTee(logCores...), zap.AddStacktrace(zapcore.DPanicLevel), zap.AddCaller())
	zap.ReplaceGlobals(impl.rootLogger)
	impl.ResetLogger(impl.baseFields...)
	impl.configHash = configHash
}

func (impl *serviceImpl) SendLog(level zapcore.Level, msg string, fields ...zap.Field) {
	if disableLog {
		return
	}
	if ce := impl.internalLogger.Check(level, msg); ce != nil {
		ce.Write(fields...)
	}
}

func (impl *serviceImpl) ResetLogger(fields ...zap.Field) {
	impl.baseFields = fields
	impl.defaultLogger = impl.rootLogger.With(fields...)
	impl.internalLogger = impl.defaultLogger.WithOptions(zap.AddCallerSkip(2))
}

func (impl *serviceImpl) InitLogger(force bool) {
	impl.mutex.Lock()
	defer impl.mutex.Unlock()
	if !force && impl.rootLogger != nil {
		return
	}
	//  初始化本地化的日志
	encodeConfig := newEncodeConfig()
	encodeConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encodeConfig), zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
	impl.rootLogger = zap.New(core, zap.AddStacktrace(zapcore.DPanicLevel), zap.AddCaller())
	LoadConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		LoadConfig()
	})
	viper.WatchConfig()
}

func (impl *serviceImpl) GetLogger() *zap.Logger {
	return impl.defaultLogger
}
