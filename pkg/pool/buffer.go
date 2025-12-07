package pool

import "sync"

// BufferPool представляет пул буферов для снижения нагрузки на GC
// Использование пула буферов критично для высоконагруженных систем
type BufferPool struct {
	pool sync.Pool
}

// NewBufferPool создает новый пул буферов указанного размера
func NewBufferPool(size int) *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				// Создаем буфер указанного размера
				buf := make([]byte, size)
				return &buf
			},
		},
	}
}

// Get получает буфер из пула
func (bp *BufferPool) Get() []byte {
	return *bp.pool.Get().(*[]byte)
}

// Put возвращает буфер обратно в пул для переиспользования
func (bp *BufferPool) Put(buf []byte) {
	// Сбрасываем длину буфера, но сохраняем capacity
	buf = buf[:cap(buf)]
	bp.pool.Put(&buf)
}
