package logward

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBatcherSizeBasedFlushing(t *testing.T) {
	var flushedCount int32
	var mu sync.Mutex
	var allLogs []Log

	flushFunc := func(ctx context.Context, logs []Log) error {
		atomic.AddInt32(&flushedCount, 1)
		mu.Lock()
		allLogs = append(allLogs, logs...)
		mu.Unlock()
		return nil
	}

	config := &BatcherConfig{
		MaxSize:       3,
		FlushInterval: 1 * time.Minute, // Long interval to test size-based flushing
		FlushFunc:     flushFunc,
	}

	batcher := NewBatcher(config)
	defer batcher.Stop()

	// Add logs
	for i := 0; i < 10; i++ {
		err := batcher.Add(Log{
			Time:    time.Now(),
			Service: "test",
			Level:   LogLevelInfo,
			Message: "test message",
		})
		if err != nil {
			t.Fatalf("Add() error = %v", err)
		}
	}

	// Wait for flushes to complete
	time.Sleep(300 * time.Millisecond)

	// Stop will flush remaining logs
	batcher.Stop()

	mu.Lock()
	totalLogs := len(allLogs)
	mu.Unlock()

	// All 10 logs should be flushed eventually
	if totalLogs != 10 {
		t.Errorf("total flushed logs = %d, want 10", totalLogs)
	}
}

func TestBatcherTimeBasedFlushing(t *testing.T) {
	var flushedCount int32

	flushFunc := func(ctx context.Context, logs []Log) error {
		atomic.AddInt32(&flushedCount, 1)
		return nil
	}

	config := &BatcherConfig{
		MaxSize:       100,
		FlushInterval: 100 * time.Millisecond,
		FlushFunc:     flushFunc,
	}

	batcher := NewBatcher(config)
	defer batcher.Stop()

	// Add a few logs (not enough to trigger size-based flush)
	for i := 0; i < 5; i++ {
		batcher.Add(Log{
			Time:    time.Now(),
			Service: "test",
			Level:   LogLevelInfo,
			Message: "test message",
		})
	}

	// Wait for time-based flush
	time.Sleep(150 * time.Millisecond)

	count := atomic.LoadInt32(&flushedCount)
	if count < 1 {
		t.Errorf("flushed count = %d, want >= 1", count)
	}
}

func TestBatcherManualFlush(t *testing.T) {
	var flushedLogs []Log
	var mu sync.Mutex

	flushFunc := func(ctx context.Context, logs []Log) error {
		mu.Lock()
		flushedLogs = append(flushedLogs, logs...)
		mu.Unlock()
		return nil
	}

	config := &BatcherConfig{
		MaxSize:       100,
		FlushInterval: 1 * time.Minute,
		FlushFunc:     flushFunc,
	}

	batcher := NewBatcher(config)
	defer batcher.Stop()

	// Add logs
	for i := 0; i < 5; i++ {
		batcher.Add(Log{
			Time:    time.Now(),
			Service: "test",
			Level:   LogLevelInfo,
			Message: "test message",
		})
	}

	// Manual flush
	ctx := context.Background()
	err := batcher.Flush(ctx)
	if err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	mu.Lock()
	count := len(flushedLogs)
	mu.Unlock()

	if count != 5 {
		t.Errorf("flushed logs = %d, want 5", count)
	}

	// Batch should be empty now
	if batcher.Size() != 0 {
		t.Errorf("batch size after flush = %d, want 0", batcher.Size())
	}
}

func TestBatcherStop(t *testing.T) {
	var flushedLogs []Log
	var mu sync.Mutex

	flushFunc := func(ctx context.Context, logs []Log) error {
		mu.Lock()
		flushedLogs = append(flushedLogs, logs...)
		mu.Unlock()
		return nil
	}

	config := &BatcherConfig{
		MaxSize:       100,
		FlushInterval: 1 * time.Minute,
		FlushFunc:     flushFunc,
	}

	batcher := NewBatcher(config)

	// Add logs
	for i := 0; i < 10; i++ {
		batcher.Add(Log{
			Time:    time.Now(),
			Service: "test",
			Level:   LogLevelInfo,
			Message: "test message",
		})
	}

	// Stop should flush remaining logs
	err := batcher.Stop()
	if err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	mu.Lock()
	count := len(flushedLogs)
	mu.Unlock()

	if count != 10 {
		t.Errorf("flushed logs on stop = %d, want 10", count)
	}

	// Adding logs after stop should fail
	err = batcher.Add(Log{
		Time:    time.Now(),
		Service: "test",
		Level:   LogLevelInfo,
		Message: "test message",
	})
	if err != ErrClientClosed {
		t.Errorf("Add() after stop error = %v, want %v", err, ErrClientClosed)
	}
}

func TestBatcherConcurrentAdds(t *testing.T) {
	var totalFlushed int32

	flushFunc := func(ctx context.Context, logs []Log) error {
		atomic.AddInt32(&totalFlushed, int32(len(logs)))
		return nil
	}

	config := &BatcherConfig{
		MaxSize:       50,
		FlushInterval: 100 * time.Millisecond,
		FlushFunc:     flushFunc,
	}

	batcher := NewBatcher(config)
	defer batcher.Stop()

	// Concurrent adds
	var wg sync.WaitGroup
	numGoroutines := 10
	logsPerGoroutine := 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				batcher.Add(Log{
					Time:    time.Now(),
					Service: "test",
					Level:   LogLevelInfo,
					Message: "concurrent message",
				})
			}
		}()
	}

	wg.Wait()

	// Stop and flush
	batcher.Stop()

	total := atomic.LoadInt32(&totalFlushed)
	expected := int32(numGoroutines * logsPerGoroutine)

	if total != expected {
		t.Errorf("total flushed logs = %d, want %d", total, expected)
	}
}

func TestBatcherEmptyFlush(t *testing.T) {
	called := false

	flushFunc := func(ctx context.Context, logs []Log) error {
		called = true
		return nil
	}

	config := &BatcherConfig{
		MaxSize:       100,
		FlushInterval: 1 * time.Minute,
		FlushFunc:     flushFunc,
	}

	batcher := NewBatcher(config)
	defer batcher.Stop()

	// Flush empty batch
	err := batcher.Flush(context.Background())
	if err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	if called {
		t.Error("flush function should not be called for empty batch")
	}
}

func TestBatcherSize(t *testing.T) {
	flushFunc := func(ctx context.Context, logs []Log) error {
		return nil
	}

	config := &BatcherConfig{
		MaxSize:       100,
		FlushInterval: 1 * time.Minute,
		FlushFunc:     flushFunc,
	}

	batcher := NewBatcher(config)
	defer batcher.Stop()

	if batcher.Size() != 0 {
		t.Errorf("initial size = %d, want 0", batcher.Size())
	}

	// Add logs
	for i := 0; i < 5; i++ {
		batcher.Add(Log{
			Time:    time.Now(),
			Service: "test",
			Level:   LogLevelInfo,
			Message: "test",
		})
	}

	if batcher.Size() != 5 {
		t.Errorf("size after adding 5 logs = %d, want 5", batcher.Size())
	}
}
