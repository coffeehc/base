package log

import (
	"io"
	"sync"
	"time"
)

type logWritePipe struct {
	io.Writer
	accepts sync.Map
	mutex   sync.Mutex
}

func (impl *logWritePipe) RegisterAccept(logWrite chan<- []byte) int64 {
	impl.mutex.Lock()
	time.Sleep(time.Nanosecond * 5)
	id := time.Now().UnixNano()
	impl.mutex.Unlock()
	impl.accepts.Store(id, logWrite)
	return id
}

func (impl *logWritePipe) UnRegisterAccept(id int64) {
	impl.accepts.Delete(id)
}

func (impl *logWritePipe) Write(p []byte) (n int, err error) {
	impl.accepts.Range(func(key, value any) bool {
		//np := make([]byte, len(p))
		//copy(np, p)
		writer := value.(chan<- []byte)
		select {
		case writer <- p:
		default:

		}
		return true
	})
	return len(p), nil
}
