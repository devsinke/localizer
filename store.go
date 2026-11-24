package localizer

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Store holds indexed translation strings with thread-safe read/write operations.
type Store struct {
	mu      sync.RWMutex
	catalog map[string]string
}

// NewStore creates a new thread-safe translation catalog.
func NewStore() *Store {
	return &Store{
		catalog: make(map[string]string),
	}
}

// Set stores a key-value translation string.
func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.catalog[key] = value
}

// Get retrieves a key translation string.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.catalog[key]
	return val, ok
}

// Merge copies entries from another map into the store.
func (s *Store) Merge(entries map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range entries {
		s.catalog[k] = v
	}
}

// Snapshot returns a shallow copy of the current catalog.
func (s *Store) Snapshot() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]string, len(s.catalog))
	for k, v := range s.catalog {
		out[k] = v
	}
	return out
}

// LoadDirectory scans a directory recursively and loads .yaml, .yml, and .json files.
func (s *Store) LoadDirectory(dirPath string) error {
	cleanDir := filepath.Clean(dirPath)
	return filepath.WalkDir(cleanDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			return nil
		}

		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		rel, relErr := filepath.Rel(cleanDir, path)
		if relErr != nil {
			return relErr
		}

		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) == 0 {
			return nil
		}

		locale := parts[0]
		prefixParts := []string{locale}
		for _, segment := range parts[1:] {
			fileNameWithoutExt := strings.TrimSuffix(segment, filepath.Ext(segment))
			if fileNameWithoutExt != "" {
				prefixParts = append(prefixParts, fileNameWithoutExt)
			}
		}
		prefix := strings.Join(prefixParts, ".")

		parsed := make(map[string]interface{})
		if ext == ".json" {
			if parseErr := json.Unmarshal(content, &parsed); parseErr != nil {
				return parseErr
			}
		} else {
			if parseErr := yaml.Unmarshal(content, &parsed); parseErr != nil {
				return parseErr
			}
		}

		flattened := make(map[string]string)
		flattenHierarchy(prefix, parsed, flattened)
		s.Merge(flattened)
		return nil
	})
}

func flattenHierarchy(prefix string, current interface{}, target map[string]string) {
	switch val := current.(type) {
	case map[string]interface{}:
		for k, v := range val {
			nextPrefix := k
			if prefix != "" {
				nextPrefix = fmt.Sprintf("%s.%s", prefix, k)
			}
			flattenHierarchy(nextPrefix, v, target)
		}
	case map[interface{}]interface{}:
		for k, v := range val {
			nextPrefix := fmt.Sprint(k)
			if prefix != "" {
				nextPrefix = fmt.Sprintf("%s.%v", prefix, k)
			}
			flattenHierarchy(nextPrefix, v, target)
		}
	default:
		target[prefix] = fmt.Sprint(val)
	}
}
