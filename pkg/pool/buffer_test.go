package pool

import (
	"bytes"
	"testing"
)

func TestBufferPool(t *testing.T) {
	// Get buffer from pool
	buf := GetBuffer()
	if buf == nil {
		t.Fatal("GetBuffer returned nil")
	}

	// Check initial state
	if buf.Len() != 0 {
		t.Errorf("New buffer should be empty, got length %d", buf.Len())
	}

	if buf.Cap() < 32*1024 {
		t.Errorf("Buffer capacity should be at least 32KB, got %d", buf.Cap())
	}

	// Use buffer
	data := []byte("test data")
	_, err := buf.Write(data)
	if err != nil {
		t.Fatalf("Failed to write to buffer: %v", err)
	}

	if buf.Len() != len(data) {
		t.Errorf("Expected buffer length %d, got %d", len(data), buf.Len())
	}

	// Return to pool
	PutBuffer(buf)

	// Get again - should be reset
	buf2 := GetBuffer()
	if buf2.Len() != 0 {
		t.Errorf("Recycled buffer should be empty, got length %d", buf2.Len())
	}

	PutBuffer(buf2)
}

func TestBufferPoolConcurrency(t *testing.T) {
	const concurrency = 100
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			// Get buffer
			buf := GetBuffer()
			if buf == nil {
				t.Error("GetBuffer returned nil")
				done <- false
				return
			}

			// Use buffer
			data := bytes.Repeat([]byte{byte(id)}, 1024)
			_, err := buf.Write(data)
			if err != nil {
				t.Errorf("Failed to write to buffer: %v", err)
				done <- false
				return
			}

			// Verify
			if buf.Len() != len(data) {
				t.Errorf("Buffer length mismatch: expected %d, got %d", len(data), buf.Len())
			}

			// Return to pool
			PutBuffer(buf)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < concurrency; i++ {
		if !<-done {
			t.Fatal("Concurrent test failed")
		}
	}
}

func TestBufferPoolReset(t *testing.T) {
	buf := GetBuffer()

	// Write some data
	originalData := []byte("original data that should be cleared")
	_, err := buf.Write(originalData)
	if err != nil {
		t.Fatalf("Failed to write to buffer: %v", err)
	}

	if buf.Len() != len(originalData) {
		t.Fatalf("Expected length %d, got %d", len(originalData), buf.Len())
	}

	// Return to pool (should reset)
	PutBuffer(buf)

	// Get new buffer
	newBuf := GetBuffer()
	if newBuf.Len() != 0 {
		t.Errorf("Buffer should be reset, but has length %d", newBuf.Len())
	}

	// Verify it doesn't contain old data
	if bytes.Contains(newBuf.Bytes(), originalData) {
		t.Error("Buffer contains old data after reset")
	}

	PutBuffer(newBuf)
}

func BenchmarkBufferPoolGet(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := GetBuffer()
		PutBuffer(buf)
	}
}

func BenchmarkBufferPoolGetPut(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := GetBuffer()
			_, _ = buf.WriteString("benchmark data")
			PutBuffer(buf)
		}
	})
}

func BenchmarkBufferPoolVsNew(b *testing.B) {
	b.Run("Pool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := GetBuffer()
			_, _ = buf.WriteString("test")
			PutBuffer(buf)
		}
	})

	b.Run("New", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(make([]byte, 0, 32*1024))
			_, _ = buf.WriteString("test")
		}
	})
}
