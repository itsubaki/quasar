package execlog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

type StdOutLogger struct {
	Writer io.Writer
	mu     sync.Mutex
}

func (l *StdOutLogger) Log(_ context.Context, entry *Entry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	writer := l.Writer
	if writer == nil {
		writer = os.Stdout
	}

	if err := json.NewEncoder(writer).Encode(entry); err != nil {
		return fmt.Errorf("write execution log: %w", err)
	}

	return nil
}
