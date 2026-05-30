package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type FileObserver struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
}

func NewFileObserver(filePath string) (*FileObserver, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path cannot be empty")
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit file: %w", err)
	}

	return &FileObserver{
		filePath: filePath,
		file:     file,
	}, nil
}

func (f *FileObserver) LogEvent(event Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.file == nil {
		file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to reopen audit file: %w", err)
		}
		f.file = file
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if _, err := f.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write to audit file: %w", err)
	}

	return nil
}

func (f *FileObserver) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.file != nil {
		return f.file.Close()
	}
	return nil
}