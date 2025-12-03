package logward

import (
	"context"
	"sync"
	"time"
)

// FlushFunc is a function that flushes a batch of logs.
type FlushFunc func(ctx context.Context, logs []Log) error

// Batcher handles automatic batching of logs with size and time-based flushing.
type Batcher struct {
	mu          sync.Mutex
	logs        []Log
	maxSize     int
	flushInterval time.Duration
	flushFunc   FlushFunc

	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	flushChan  chan struct{}
	stopped    bool
}

// BatcherConfig holds the configuration for a batcher.
type BatcherConfig struct {
	MaxSize       int
	FlushInterval time.Duration
	FlushFunc     FlushFunc
}

// DefaultBatcherConfig returns the default batcher configuration.
func DefaultBatcherConfig(flushFunc FlushFunc) *BatcherConfig {
	return &BatcherConfig{
		MaxSize:       100,
		FlushInterval: 5 * time.Second,
		FlushFunc:     flushFunc,
	}
}

// NewBatcher creates a new batcher with the specified configuration.
func NewBatcher(config *BatcherConfig) *Batcher {
	if config == nil {
		panic("batcher config cannot be nil")
	}
	if config.FlushFunc == nil {
		panic("flush function cannot be nil")
	}
	if config.MaxSize <= 0 {
		config.MaxSize = 100
	}
	if config.FlushInterval <= 0 {
		config.FlushInterval = 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	b := &Batcher{
		logs:          make([]Log, 0, config.MaxSize),
		maxSize:       config.MaxSize,
		flushInterval: config.FlushInterval,
		flushFunc:     config.FlushFunc,
		ctx:           ctx,
		cancel:        cancel,
		flushChan:     make(chan struct{}, 1),
	}

	// Start background flusher
	b.wg.Add(1)
	go b.backgroundFlusher()

	return b
}

// Add adds a log to the batch. If the batch size reaches maxSize, it triggers a flush.
func (b *Batcher) Add(log Log) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.stopped {
		return ErrClientClosed
	}

	// Add log to batch
	b.logs = append(b.logs, log)

	// Check if we need to flush based on size
	if len(b.logs) >= b.maxSize {
		// Trigger immediate flush
		select {
		case b.flushChan <- struct{}{}:
		default:
			// Flush already pending
		}
	}

	return nil
}

// Flush immediately flushes all pending logs.
func (b *Batcher) Flush(ctx context.Context) error {
	b.mu.Lock()

	if len(b.logs) == 0 {
		b.mu.Unlock()
		return nil
	}

	// Take logs and reset batch
	logs := make([]Log, len(b.logs))
	copy(logs, b.logs)
	b.logs = b.logs[:0] // Reset slice but keep capacity

	b.mu.Unlock()

	// Flush logs
	return b.flushFunc(ctx, logs)
}

// Stop stops the batcher and flushes any remaining logs.
func (b *Batcher) Stop() error {
	b.mu.Lock()
	if b.stopped {
		b.mu.Unlock()
		return nil
	}
	b.stopped = true
	b.mu.Unlock()

	// Cancel background goroutine
	b.cancel()

	// Wait for background goroutine to finish
	b.wg.Wait()

	// Flush remaining logs
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return b.Flush(ctx)
}

// backgroundFlusher runs in a goroutine and periodically flushes logs.
func (b *Batcher) backgroundFlusher() {
	defer b.wg.Done()

	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			// Batcher stopped
			return

		case <-ticker.C:
			// Time-based flush
			if err := b.Flush(b.ctx); err != nil {
				// TODO: Consider adding error callback
				// For now, silently continue
			}

		case <-b.flushChan:
			// Size-based flush
			if err := b.Flush(b.ctx); err != nil {
				// TODO: Consider adding error callback
				// For now, silently continue
			}
		}
	}
}

// Size returns the current number of logs in the batch.
func (b *Batcher) Size() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.logs)
}
