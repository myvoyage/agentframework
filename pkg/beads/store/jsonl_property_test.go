package store

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"AgentFramework/pkg/beads"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: beads-task-tracking, Property 4: Append-Only JSONL Semantics
// **Validates: Requirements 1.5**

// TestProperty4_AppendOnlyJSONLSemantics verifies that for any sequence of task operations,
// the JSONL store only appends new lines and never modifies or deletes existing lines.
func TestProperty4_AppendOnlyJSONLSemantics(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10
	properties := gopter.NewProperties(parameters)

	properties.Property("JSONL store only appends, never modifies or deletes", prop.ForAll(
		func(eventSequences [][]beads.Event) bool {
			// Create a temporary directory for this test
			tmpDir := t.TempDir()
			store, err := NewJSONLStore(tmpDir)
			if err != nil {
				t.Logf("Failed to create store: %v", err)
				return false
			}
			defer store.Close()

			ctx := context.Background()

			// Process each sequence of events
			for seqIdx, events := range eventSequences {
				if len(events) == 0 {
					continue
				}

				// Capture file contents before appending
				beforeContents, err := captureAllFileContents(tmpDir)
				if err != nil {
					t.Logf("Failed to capture file contents before append: %v", err)
					return false
				}

				// Append events
				for i := range events {
					if err := store.AppendEvent(ctx, &events[i]); err != nil {
						t.Logf("Failed to append event: %v", err)
						return false
					}
				}

				// Force flush to ensure events are written
				if err := store.ForceFlush(); err != nil {
					t.Logf("Failed to flush: %v", err)
					return false
				}

				// Capture file contents after appending
				afterContents, err := captureAllFileContents(tmpDir)
				if err != nil {
					t.Logf("Failed to capture file contents after append: %v", err)
					return false
				}

				// Verify append-only semantics:
				// 1. All previous content must still exist unchanged
				// 2. New content must be added at the end
				if !verifyAppendOnly(beforeContents, afterContents) {
					t.Logf("Append-only semantics violated at sequence %d", seqIdx)
					return false
				}
			}

			return true
		},
		genEventSequences(),
	))

	properties.TestingRun(t)
}

// genEventSequences generates sequences of event batches for property testing
func genEventSequences() gopter.Gen {
	return gen.SliceOf(gen.SliceOf(genEvent()))
}

// genEvent generates a random event for property testing
func genEvent() gopter.Gen {
	return gopter.CombineGens(
		genEventType(),
		gen.Identifier(),
		gen.Int64Range(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
			time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC).Unix()),
		gen.MapOf(gen.Identifier(), gen.OneConstOf("value1", "value2", "value3")),
	).Map(func(values []interface{}) beads.Event {
		eventType := values[0].(beads.EventType)
		taskID := values[1].(string)
		timestamp := values[2].(int64)
		data := values[3]

		// Convert to map[string]interface{}
		dataMap := make(map[string]interface{})
		
		// Handle different map types that gopter might generate
		switch v := data.(type) {
		case map[string]interface{}:
			dataMap = v
		case map[string]string:
			for k, val := range v {
				dataMap[k] = val
			}
		case map[interface{}]interface{}:
			for k, val := range v {
				if strKey, ok := k.(string); ok {
					dataMap[strKey] = val
				}
			}
		}

		return beads.Event{
			Type:      eventType,
			TaskID:    taskID,
			Timestamp: timestamp,
			Data:      dataMap,
		}
	})
}

// genEventType generates a random event type
func genEventType() gopter.Gen {
	return gen.OneConstOf(
		beads.EventTaskCreated,
		beads.EventTaskUpdated,
		beads.EventTaskClosed,
		beads.EventDependencyAdded,
		beads.EventDependencyRemoved,
	)
}

// captureAllFileContents reads all JSONL files in the directory and returns their contents
func captureAllFileContents(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var contents []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if len(entry.Name()) >= 6 && entry.Name()[len(entry.Name())-6:] == ".jsonl" {
			filePath := dir + string(os.PathSeparator) + entry.Name()
			data, err := os.ReadFile(filePath)
			if err != nil {
				return nil, err
			}
			contents = append(contents, string(data))
		}
	}

	return contents, nil
}

// verifyAppendOnly checks that afterContents contains all of beforeContents
// at the beginning, with only new content appended at the end
func verifyAppendOnly(before, after []string) bool {
	// If we have more files after than before, that's okay (new month files)
	// But all previous files must still exist with their content intact

	// Create a map of before contents for easy lookup
	beforeMap := make(map[string]bool)
	for _, content := range before {
		beforeMap[content] = false // false means not yet found in after
	}

	// Check that all before contents exist in after
	for _, afterContent := range after {
		// Check if this after content starts with any before content
		for beforeContent := range beforeMap {
			if len(beforeContent) > 0 && len(afterContent) >= len(beforeContent) {
				// The after content should either be identical or have the before content as a prefix
				if afterContent[:len(beforeContent)] == beforeContent {
					beforeMap[beforeContent] = true
					// Verify that only content was appended (no modification in the middle)
					if afterContent != beforeContent {
						// New content was added - verify it's only at the end
						// The new content should start right after the old content
						newContent := afterContent[len(beforeContent):]
						// Verify new content doesn't contain any deletions or modifications
						// by checking that it's valid JSONL (starts with { or is empty)
						if len(newContent) > 0 && newContent[0] != '{' && newContent[0] != '\n' {
							return false
						}
					}
				}
			}
		}
	}

	// Verify all before contents were found
	for _, found := range beforeMap {
		if !found && len(before) > 0 {
			// A previous file's content was not found - this violates append-only
			return false
		}
	}

	return true
}

// Feature: beads-task-tracking, Property 23: JSONL Single-Line Format
// **Validates: Requirements 7.4**

// TestProperty23_JSONLSingleLineFormat verifies that for any event written to JSONL_Store,
// the event occupies exactly one line (no embedded newlines) to enable line-based Git merging.
func TestProperty23_JSONLSingleLineFormat(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 10
	properties := gopter.NewProperties(parameters)

	properties.Property("each event occupies exactly one line with no embedded newlines", prop.ForAll(
		func(events []beads.Event) bool {
			if len(events) == 0 {
				return true // Empty case is valid
			}

			// Create a temporary directory for this test
			tmpDir := t.TempDir()
			store, err := NewJSONLStore(tmpDir)
			if err != nil {
				t.Logf("Failed to create store: %v", err)
				return false
			}
			defer store.Close()

			ctx := context.Background()

			// Append all events
			for i := range events {
				if err := store.AppendEvent(ctx, &events[i]); err != nil {
					t.Logf("Failed to append event: %v", err)
					return false
				}
			}

			// Force flush to ensure events are written
			if err := store.ForceFlush(); err != nil {
				t.Logf("Failed to flush: %v", err)
				return false
			}

			// Read all JSONL files and verify single-line format
			files, err := os.ReadDir(tmpDir)
			if err != nil {
				t.Logf("Failed to read directory: %v", err)
				return false
			}

			eventCount := 0
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				if len(file.Name()) < 6 || file.Name()[len(file.Name())-6:] != ".jsonl" {
					continue
				}

				filePath := tmpDir + string(os.PathSeparator) + file.Name()
				f, err := os.Open(filePath)
				if err != nil {
					t.Logf("Failed to open file %s: %v", filePath, err)
					return false
				}

				scanner := bufio.NewScanner(f)
				lineNum := 0
				for scanner.Scan() {
					lineNum++
					line := scanner.Text()
					
					if line == "" {
						continue // Empty lines are okay
					}

					eventCount++

					// Verify the line is valid JSON
					var event beads.Event
					if err := json.Unmarshal([]byte(line), &event); err != nil {
						t.Logf("Line %d in %s is not valid JSON: %v", lineNum, file.Name(), err)
						f.Close()
						return false
					}

					// Verify no embedded newlines in the original line
					// The scanner already splits by newlines, so if we got here,
					// the line itself doesn't contain newlines. But let's verify
					// the marshaled JSON also doesn't contain newlines.
					marshaled, err := json.Marshal(event)
					if err != nil {
						t.Logf("Failed to re-marshal event: %v", err)
						f.Close()
						return false
					}

					// Check that the marshaled JSON contains no newlines
					for i, b := range marshaled {
						if b == '\n' || b == '\r' {
							t.Logf("Event contains embedded newline at position %d: %s", i, string(marshaled))
							f.Close()
							return false
						}
					}

					// Verify the line matches the expected format (JSON object on a single line)
					if line[0] != '{' {
						t.Logf("Line %d in %s does not start with '{': %s", lineNum, file.Name(), line)
						f.Close()
						return false
					}
				}

				if err := scanner.Err(); err != nil {
					t.Logf("Scanner error for %s: %v", filePath, err)
					f.Close()
					return false
				}

				f.Close()
			}

			// Verify we found all events
			if eventCount != len(events) {
				t.Logf("Expected %d events, found %d", len(events), eventCount)
				return false
			}

			return true
		},
		gen.SliceOf(genEvent()),
	))

	properties.TestingRun(t)
}
