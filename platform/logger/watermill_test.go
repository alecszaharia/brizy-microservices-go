package logger

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
)

// captureLogger captures log calls for verification (thread-safe)
type captureLogger struct {
	mu    sync.Mutex
	calls []logCall
}

type logCall struct {
	level   log.Level
	keyvals []interface{}
}

func (c *captureLogger) Log(level log.Level, keyvals ...interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, logCall{
		level:   level,
		keyvals: keyvals,
	})
	return nil
}

func (c *captureLogger) lastCall() *logCall {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.calls) == 0 {
		return nil
	}
	return &c.calls[len(c.calls)-1]
}

func (c *captureLogger) getMessage() string {
	call := c.lastCall()
	if call == nil {
		return ""
	}
	// Kratos log.Helper formats as: [key, value, key, value, ...]
	// Find the "msg" key
	for i := 0; i < len(call.keyvals)-1; i += 2 {
		if key, ok := call.keyvals[i].(string); ok && key == "msg" {
			if val, ok := call.keyvals[i+1].(string); ok {
				return val
			}
		}
	}
	return ""
}

func TestNewWatermillLogger(t *testing.T) {
	t.Run("creates logger successfully", func(t *testing.T) {
		wl := NewWatermillLogger(log.DefaultLogger)

		assert.NotNil(t, wl)
		assert.NotNil(t, wl.log)
	})
}

func TestWatermillLogger_Error(t *testing.T) {
	capture := &captureLogger{}
	wl := NewWatermillLogger(capture)

	wl.Error("test error", fmt.Errorf("error"), watermill.LogFields{"key": "value"})

	call := capture.lastCall()
	assert.NotNil(t, call)
	assert.Equal(t, log.LevelError, call.level)
	assert.Contains(t, capture.getMessage(), "test error")
}

func TestWatermillLogger_Info(t *testing.T) {
	capture := &captureLogger{}
	wl := NewWatermillLogger(capture)

	wl.Info("test info", watermill.LogFields{"key": "value"})

	call := capture.lastCall()
	assert.NotNil(t, call)
	assert.Equal(t, log.LevelInfo, call.level)
	assert.Contains(t, capture.getMessage(), "test info")
}

func TestWatermillLogger_Debug(t *testing.T) {
	capture := &captureLogger{}
	wl := NewWatermillLogger(capture)

	wl.Debug("test debug", watermill.LogFields{"key": "value"})

	call := capture.lastCall()
	assert.NotNil(t, call)
	assert.Equal(t, log.LevelDebug, call.level)
	assert.Contains(t, capture.getMessage(), "test debug")
}

func TestWatermillLogger_Trace(t *testing.T) {
	capture := &captureLogger{}
	wl := NewWatermillLogger(capture)

	wl.Trace("test trace", watermill.LogFields{"key": "value"})

	call := capture.lastCall()
	assert.NotNil(t, call)
	assert.Equal(t, log.LevelDebug, call.level)
	msg := capture.getMessage()
	assert.Contains(t, msg, "[TRACE]")
	assert.Contains(t, msg, "test trace")
}

func TestWatermillLogger_With(t *testing.T) {
	capture := &captureLogger{}
	wl := NewWatermillLogger(capture)

	result := wl.With(watermill.LogFields{"key": "value"})

	// Current implementation returns the same instance
	assert.Equal(t, wl, result)
}

func TestWatermillLogger_InterfaceCompliance(t *testing.T) {
	wl := NewWatermillLogger(log.DefaultLogger)

	// Compile-time check: verify it implements watermill.LoggerAdapter
	var _ watermill.LoggerAdapter = wl
}
