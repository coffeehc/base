package log

import "go.uber.org/zap/zapcore"

type PBWrite interface {
	zapcore.WriteSyncer
}

func newPBWrite() PBWrite {
	impl := &pbWriteImpl{}
	return impl
}

type pbWriteImpl struct {
}

func (impl *pbWriteImpl) Write(p []byte) (n int, err error) {
	panic("implement me")
}

func (impl *pbWriteImpl) Sync() error {
	panic("implement me")
}
