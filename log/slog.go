package log

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log/slog"
)

type ZapHandler struct {
	logger *zap.Logger
	level  slog.Level
	group  string
	attrs  []slog.Attr
}

var slogger *slog.Logger

func GetSlog() *slog.Logger {
	if slogger == nil {
		handler := NewZapHandler(GetLogger(), nil)
		slogger = slog.New(handler)
	}
	return slogger
}

// NewZapHandler 创建适配器实例
func NewZapHandler(z *zap.Logger, opts *slog.HandlerOptions) slog.Handler {
	level := slog.LevelInfo
	if opts != nil && opts.Level != nil {
		level = opts.Level.Level()
	}

	return &ZapHandler{
		logger: z,
		level:  level,
		attrs:  []slog.Attr{},
	}
}

func (h *ZapHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return l >= h.level
}

func (h *ZapHandler) Handle(ctx context.Context, r slog.Record) error {
	if !h.Enabled(ctx, r.Level) {
		return nil
	}

	// 转换日志级别
	var zapLevel zapcore.Level
	switch r.Level {
	case slog.LevelDebug:
		zapLevel = zapcore.DebugLevel
	case slog.LevelInfo:
		zapLevel = zapcore.InfoLevel
	case slog.LevelWarn:
		zapLevel = zapcore.WarnLevel
	default:
		zapLevel = zapcore.ErrorLevel
	}

	// 组装键值对
	fields := make([]zap.Field, 0, len(h.attrs)+r.NumAttrs())
	for _, a := range h.attrs {
		fields = append(fields, zap.Any(a.Key, a.Value))
	}

	r.Attrs(func(a slog.Attr) bool {
		fields = append(fields, zap.Any(a.Key, a.Value))
		return true
	})

	// 输出日志
	ce := h.logger.Check(zapLevel, r.Message)
	if ce != nil {
		ce.Write(fields...)
	}
	return nil
}

func (h *ZapHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs), len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	newAttrs = append(newAttrs, attrs...)

	return &ZapHandler{
		logger: h.logger,
		level:  h.level,
		group:  h.group,
		attrs:  newAttrs,
	}
}

func (h *ZapHandler) WithGroup(name string) slog.Handler {
	return &ZapHandler{
		logger: h.logger,
		level:  h.level,
		group:  name,
		attrs:  h.attrs,
	}
}
