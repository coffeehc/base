package log

import (
	"io"
	"sync"
	"sync/atomic"
)

type logWritePipe struct {
	io.Writer
	accepts      sync.Map
	acceptSeq    atomic.Int64
	historyMutex sync.RWMutex
	// history 采用环形缓冲，避免调试旁路无限占用内存。
	history       [][]byte
	historyCap    int
	historySize   int
	historyOffset int
}

const defaultLogHistoryCap = 1024

func newLogWritePipe(historyCap int) *logWritePipe {
	if historyCap <= 0 {
		historyCap = defaultLogHistoryCap
	}
	return &logWritePipe{
		history:    make([][]byte, historyCap),
		historyCap: historyCap,
	}
}

func (impl *logWritePipe) RegisterAccept(logWrite chan<- []byte) int64 {
	id := impl.acceptSeq.Add(1)
	impl.accepts.Store(id, logWrite)
	return id
}

func (impl *logWritePipe) UnRegisterAccept(id int64) {
	impl.accepts.Delete(id)
}

func (impl *logWritePipe) GetRecentLogs(limit int) [][]byte {
	impl.historyMutex.RLock()
	defer impl.historyMutex.RUnlock()
	if impl.historySize == 0 {
		return nil
	}
	if limit <= 0 || limit > impl.historySize {
		limit = impl.historySize
	}
	out := make([][]byte, limit)
	start := impl.historySize - limit
	// 从最老到最新返回，方便直接按顺序展示。
	for i := 0; i < limit; i++ {
		idx := (impl.historyOffset + start + i) % impl.historyCap
		out[i] = cloneLogEntry(impl.history[idx])
	}
	return out
}

func (impl *logWritePipe) AppendHistory(p []byte) {
	if len(p) == 0 {
		return
	}
	entry := cloneLogEntry(p)
	impl.historyMutex.Lock()
	defer impl.historyMutex.Unlock()
	if impl.historySize < impl.historyCap {
		idx := (impl.historyOffset + impl.historySize) % impl.historyCap
		impl.history[idx] = entry
		impl.historySize++
		return
	}
	impl.history[impl.historyOffset] = entry
	impl.historyOffset = (impl.historyOffset + 1) % impl.historyCap
}

func cloneLogEntry(p []byte) []byte {
	out := make([]byte, len(p))
	copy(out, p)
	return out
}

func (impl *logWritePipe) Write(p []byte) (n int, err error) {
	impl.AppendHistory(p)
	impl.accepts.Range(func(key, value any) bool {
		id, ok := key.(int64)
		if !ok {
			return true
		}
		writer := value.(chan<- []byte)
		// 每个订阅方独立拷贝，避免共享底层切片引发并发污染。
		payload := cloneLogEntry(p)
		defer func() {
			// 如果订阅方通道被外部关闭，自动摘除该订阅，避免持续 panic。
			if r := recover(); r != nil {
				impl.accepts.Delete(id)
			}
		}()
		select {
		case writer <- payload:
		default:
		}
		return true
	})
	return len(p), nil
}
