package logger

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, LevelInfo, false)
	
	assert.NotNil(t, log)
	assert.Equal(t, LevelInfo, log.level)
	assert.False(t, log.quiet)
}

func TestLogger_Debug(t *testing.T) {
	tests := []struct {
		name          string
		level         Level
		message       string
		args          []interface{}
		expectOutput  bool
	}{
		{
			name:         "debug enabled",
			level:        LevelDebug,
			message:      "debug message",
			expectOutput: true,
		},
		{
			name:         "debug disabled at info level",
			level:        LevelInfo,
			message:      "debug message",
			expectOutput: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log := New(&buf, tt.level, false)
			
			log.Debug(tt.message, tt.args...)
			
			output := buf.String()
			if tt.expectOutput {
				assert.Contains(t, output, tt.message)
				assert.Contains(t, output, "[DEBUG]")
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, LevelInfo, false)
	
	log.Info("info message")
	
	output := buf.String()
	assert.Contains(t, output, "info message")
	assert.Contains(t, output, "[INFO]")
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, LevelWarn, false)
	
	log.Warn("warning message")
	
	output := buf.String()
	assert.Contains(t, output, "warning message")
	assert.Contains(t, output, "[WARN]")
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, LevelError, false)
	
	log.Error("error message")
	
	output := buf.String()
	assert.Contains(t, output, "error message")
	assert.Contains(t, output, "[ERROR]")
}

func TestLogger_QuietMode(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, LevelInfo, true)
	
	log.Info("should not appear")
	log.Warn("should not appear")
	
	output := buf.String()
	assert.Empty(t, output)
	
	// Errors should still appear in quiet mode
	log.Error("should appear")
	output = buf.String()
	assert.Contains(t, output, "should appear")
}

func TestLogger_Formatting(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, LevelInfo, false)
	
	log.Info("formatted: %s %d", "test", 123)
	
	output := buf.String()
	assert.Contains(t, output, "formatted: test 123")
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Level
		wantErr bool
	}{
		{
			name:  "debug",
			input: "debug",
			want:  LevelDebug,
		},
		{
			name:  "info",
			input: "info",
			want:  LevelInfo,
		},
		{
			name:  "warn",
			input: "warn",
			want:  LevelWarn,
		},
		{
			name:  "error",
			input: "error",
			want:  LevelError,
		},
		{
			name:  "uppercase",
			input: "INFO",
			want:  LevelInfo,
		},
		{
			name:    "invalid",
			input:   "invalid",
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLevel(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
	}
	
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.level.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
