package types

import (
	"io"
	"strings"
)

// LogWriter implements io.Writer for capturing logs
type LogWriter struct {
	Logs *[]string
}

// MaxLogs is the maximum number of log entries to keep
const MaxLogs = 100

// Write implements io.Writer
func (w LogWriter) Write(p []byte) (n int, err error) {
	logStr := string(p)
	if strings.Contains(logStr, `"level":"DEBUG"`) {
		return len(p), nil
	}
	*w.Logs = append(*w.Logs, logStr)
	if len(*w.Logs) > MaxLogs {
		*w.Logs = (*w.Logs)[len(*w.Logs)-MaxLogs:]
	}
	return len(p), nil
}

var _ io.Writer = LogWriter{}
