// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025 Agent Framework Contributors

// SPDX-License-Identifier: AGPL-3.0-or-later

package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"

	"AgentFramework/agent/skills"
)

// MarkdownSkillDiscoverer discovers skills from Markdown files
type MarkdownSkillDiscoverer struct {
	directories []string
	cache       map[string]*skills.SkillDefinition
	mu          sync.RWMutex
	watcher     *fsnotify.Watcher
}

// NewMarkdownSkillDiscoverer creates a new MarkdownSkillDiscoverer
func NewMarkdownSkillDiscoverer(directories []string) (*MarkdownSkillDiscoverer, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	discoverer := &MarkdownSkillDiscoverer{
		directories: directories,
		cache:       make(map[string]*skills.SkillDefinition),
		watcher:     watcher,
	}

	// Watch directories for changes
	for _, dir := range directories {
		if err := discoverer.watchDirectory(dir); err != nil {
			return nil, err
		}
	}

	// Start watching goroutine
	go discoverer.watchLoop()

	return discoverer, nil
}

// Discover discovers all skills in configured directories
func (d *MarkdownSkillDiscoverer) Discover() ([]*skills.SkillDefinition, error) {
	var allSkills []*skills.SkillDefinition

	for _, dir := range d.directories {
		skillsInDir, err := d.discoverInDirectory(dir)
		if err != nil {
			return nil, err
		}
		allSkills = append(allSkills, skillsInDir...)
	}

	return allSkills, nil
}

// DiscoverWithPriority discovers skills with priority levels
// Priority: lower number = higher priority
func (d *MarkdownSkillDiscoverer) DiscoverWithPriority() (map[int][]*skills.SkillDefinition, error) {
	results := make(map[int][]*skills.SkillDefinition)

	for priority, dir := range d.directories {
		skillsInDir, err := d.discoverInDirectory(dir)
		if err != nil {
			return nil, err
		}
		results[priority] = skillsInDir
	}

	return results, nil
}

// Watch starts watching for changes and calls callback when a skill is updated
func (d *MarkdownSkillDiscoverer) Watch(callback func(*skills.SkillDefinition, string)) error {
	go func() {
		for {
			select {
			case event, ok := <-d.watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					if strings.HasSuffix(event.Name, "SKILL.md") {
						def, parseErr := NewMarkdownSkillParser().Parse(event.Name)
						if parseErr == nil {
							d.mu.Lock()
							d.cache[event.Name] = def
							d.mu.Unlock()
							callback(def, event.Name)
						}
					}
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					d.mu.Lock()
					delete(d.cache, event.Name)
					d.mu.Unlock()
					callback(nil, event.Name)
				}

			case _, ok := <-d.watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()

	return nil
}

// Stop stops watching for changes
func (d *MarkdownSkillDiscoverer) Stop() {
	if d.watcher != nil {
		d.watcher.Close()
	}
}

// watchLoop processes watch events
func (d *MarkdownSkillDiscoverer) watchLoop() {
	for {
		select {
		case event, ok := <-d.watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				if strings.HasSuffix(event.Name, "SKILL.md") {
					// File was modified - reparse and update cache
					if def, parseErr := NewMarkdownSkillParser().Parse(event.Name); parseErr == nil {
						d.mu.Lock()
						d.cache[event.Name] = def
						d.mu.Unlock()
					}
				}
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				// File was removed - remove from cache
				d.mu.Lock()
				delete(d.cache, event.Name)
				d.mu.Unlock()
			}

		case _, ok := <-d.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

// watchDirectory adds a directory to watch list
func (d *MarkdownSkillDiscoverer) watchDirectory(dirPath string) error {
	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return err
		}
	}

	if err := d.watcher.Add(dirPath); err != nil {
		return err
	}

	// Recursively watch subdirectories
	return filepath.WalkDir(dirPath, func(path string, dEntry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if dEntry.IsDir() && path != dirPath {
			return d.watcher.Add(path)
		}
		return nil
	})
}

// discoverInDirectory discovers skills in a single directory
func (d *MarkdownSkillDiscoverer) discoverInDirectory(dirPath string) ([]*skills.SkillDefinition, error) {
	var foundSkills []*skills.SkillDefinition

	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return foundSkills, nil
	}

	err := filepath.WalkDir(dirPath, func(path string, dEntry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !dEntry.IsDir() && strings.EqualFold(filepath.Base(path), "SKILL.md") {
			// Parse SKILL.md file
			def, err := NewMarkdownSkillParser().Parse(path)
			if err != nil {
				return fmt.Errorf("parse %s failed: %w", path, err)
			}

			d.mu.Lock()
			d.cache[path] = def
			d.mu.Unlock()

			foundSkills = append(foundSkills, def)
		}

		return nil
	})

	return foundSkills, err
}
