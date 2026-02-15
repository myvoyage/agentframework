package store

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"AgentFramework/pkg/beads"
)

// JSONLStore implements the beads.JSONLStore interface
// It provides Git-tracked append-only event log with file partitioning by month
type JSONLStore struct {
	basePath string
	mu       sync.RWMutex
	// Buffer for batched writes
	buffer       []*beads.Event
	bufferSize   int
	bufferMu     sync.Mutex
	flushTicker  *time.Ticker
	stopFlush    chan struct{}
	flushStarted bool
}

// NewJSONLStore creates a new JSONL store with the specified base path
func NewJSONLStore(basePath string) (*JSONLStore, error) {
	// Create the base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create JSONL directory: %w", err)
	}

	store := &JSONLStore{
		basePath:   basePath,
		buffer:     make([]*beads.Event, 0, 100),
		bufferSize: 100,
		stopFlush:  make(chan struct{}),
	}

	// Start background flush goroutine
	store.startFlushRoutine()

	return store, nil
}

// startFlushRoutine starts a background goroutine that periodically flushes the buffer
func (s *JSONLStore) startFlushRoutine() {
	s.bufferMu.Lock()
	if s.flushStarted {
		s.bufferMu.Unlock()
		return
	}
	s.flushStarted = true
	s.bufferMu.Unlock()

	s.flushTicker = time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-s.flushTicker.C:
				s.flush()
			case <-s.stopFlush:
				s.flush() // Final flush before stopping
				return
			}
		}
	}()
}

// AppendEvent appends an event to the JSONL store with buffered writes
func (s *JSONLStore) AppendEvent(ctx context.Context, event *beads.Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	// Add to buffer
	s.bufferMu.Lock()
	s.buffer = append(s.buffer, event)
	shouldFlush := len(s.buffer) >= s.bufferSize
	s.bufferMu.Unlock()

	// Flush if buffer is full
	if shouldFlush {
		return s.flush()
	}

	return nil
}

// flush writes buffered events to disk
func (s *JSONLStore) flush() error {
	s.bufferMu.Lock()
	if len(s.buffer) == 0 {
		s.bufferMu.Unlock()
		return nil
	}

	// Take ownership of current buffer
	eventsToWrite := s.buffer
	s.buffer = make([]*beads.Event, 0, s.bufferSize)
	s.bufferMu.Unlock()

	// Group events by month
	eventsByMonth := make(map[string][]*beads.Event)
	for _, event := range eventsToWrite {
		monthKey := s.getMonthKey(time.Unix(event.Timestamp, 0))
		eventsByMonth[monthKey] = append(eventsByMonth[monthKey], event)
	}

	// Write each month's events
	s.mu.Lock()
	defer s.mu.Unlock()

	for monthKey, events := range eventsByMonth {
		if err := s.appendEventsToFile(monthKey, events); err != nil {
			// Put events back in buffer on error
			s.bufferMu.Lock()
			s.buffer = append(events, s.buffer...)
			s.bufferMu.Unlock()
			return err
		}
	}

	return nil
}

// appendEventsToFile appends events to the appropriate monthly file
func (s *JSONLStore) appendEventsToFile(monthKey string, events []*beads.Event) error {
	filePath := s.getFilePath(monthKey)

	// Open file in append mode, create if doesn't exist
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open JSONL file: %w", err)
	}
	defer file.Close()

	// Write each event as a single line
	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		// Ensure single line (no embedded newlines)
		if _, err := file.Write(data); err != nil {
			return fmt.Errorf("failed to write event: %w", err)
		}
		if _, err := file.Write([]byte("\n")); err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
	}

	return nil
}

// ReadEvents reads events from the JSONL store since the specified time
func (s *JSONLStore) ReadEvents(ctx context.Context, since time.Time) ([]*beads.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get all JSONL files
	files, err := s.getJSONLFiles()
	if err != nil {
		return nil, err
	}

	var events []*beads.Event

	// Read events from each file
	for _, file := range files {
		fileEvents, err := s.readEventsFromFile(file, since)
		if err != nil {
			return nil, err
		}
		events = append(events, fileEvents...)
	}

	// Sort events by timestamp
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp < events[j].Timestamp
	})

	return events, nil
}

// ReadAllEvents reads all events from the JSONL store
func (s *JSONLStore) ReadAllEvents(ctx context.Context) ([]*beads.Event, error) {
	return s.ReadEvents(ctx, time.Time{})
}

// GetLatestTimestamp returns the timestamp of the most recent event
func (s *JSONLStore) GetLatestTimestamp(ctx context.Context) (time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get all JSONL files
	files, err := s.getJSONLFiles()
	if err != nil {
		return time.Time{}, err
	}

	if len(files) == 0 {
		return time.Time{}, nil
	}

	// Start from the most recent file and work backwards
	for i := len(files) - 1; i >= 0; i-- {
		timestamp, err := s.getLastTimestampFromFile(files[i])
		if err != nil {
			continue // Skip files with errors
		}
		if !timestamp.IsZero() {
			return timestamp, nil
		}
	}

	return time.Time{}, nil
}

// getLastTimestampFromFile reads the last event from a file and returns its timestamp
func (s *JSONLStore) getLastTimestampFromFile(filePath string) (time.Time, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return time.Time{}, err
	}
	defer file.Close()

	var lastLine string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		return time.Time{}, err
	}

	if lastLine == "" {
		return time.Time{}, nil
	}

	var event beads.Event
	if err := json.Unmarshal([]byte(lastLine), &event); err != nil {
		return time.Time{}, err
	}

	return time.Unix(event.Timestamp, 0), nil
}

// readEventsFromFile reads events from a single JSONL file
func (s *JSONLStore) readEventsFromFile(filePath string, since time.Time) ([]*beads.Event, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to open JSONL file: %w", err)
	}
	defer file.Close()

	var events []*beads.Event
	scanner := bufio.NewScanner(file)

	// Increase buffer size for large lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event beads.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Log error but continue processing other events
			fmt.Fprintf(os.Stderr, "Warning: failed to unmarshal event at %s:%d: %v\n", filePath, lineNum, err)
			continue
		}

		// Filter by timestamp
		eventTime := time.Unix(event.Timestamp, 0)
		if eventTime.After(since) || eventTime.Equal(since) {
			events = append(events, &event)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading JSONL file: %w", err)
	}

	return events, nil
}

// getJSONLFiles returns a sorted list of all JSONL files in the store
func (s *JSONLStore) getJSONLFiles() ([]string, error) {
	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read JSONL directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".jsonl" {
			files = append(files, filepath.Join(s.basePath, entry.Name()))
		}
	}

	// Sort files by name (which sorts by date due to YYYY-MM format)
	sort.Strings(files)

	return files, nil
}

// getMonthKey returns the month key for a given time (format: YYYY-MM)
func (s *JSONLStore) getMonthKey(t time.Time) string {
	return t.UTC().Format("2006-01")
}

// getFilePath returns the file path for a given month key
func (s *JSONLStore) getFilePath(monthKey string) string {
	return filepath.Join(s.basePath, monthKey+".jsonl")
}

// Close closes the JSONL store and flushes any remaining buffered events
func (s *JSONLStore) Close() error {
	// Stop flush routine
	s.bufferMu.Lock()
	if s.flushStarted {
		close(s.stopFlush)
		if s.flushTicker != nil {
			s.flushTicker.Stop()
		}
		s.flushStarted = false
	}
	s.bufferMu.Unlock()

	// Final flush
	return s.flush()
}

// ForceFlush forces an immediate flush of buffered events (useful for testing)
func (s *JSONLStore) ForceFlush() error {
	return s.flush()
}
