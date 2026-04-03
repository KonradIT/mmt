package camera

import (
	"fmt"
	"sort"
	"sync"
)

// entry is the minimal interface the registry needs to index and match cameras.
type entry interface {
	Name() string
	GuessFromPath(root string) bool
}

var (
	mu       sync.RWMutex
	registry = map[string]any{}
)

// Register adds a camera backend to the global registry.
// c must implement Name() string and GuessFromPath(root string) bool.
// Typically called from a package's init function.
func Register(c entry) {
	mu.Lock()
	defer mu.Unlock()

	registry[c.Name()] = c
}

// Get returns the camera registered under the given name.
func Get(name string) (any, error) {
	mu.RLock()
	defer mu.RUnlock()

	c, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("camera %q is not supported", name)
	}

	return c, nil
}

// Guess returns the first registered camera whose GuessFromPath matches root,
// or nil if no camera matches.
func Guess(root string) any {
	mu.RLock()
	defer mu.RUnlock()

	for _, c := range registry {
		if e, ok := c.(entry); ok && e.GuessFromPath(root) {
			return c
		}
	}

	return nil
}

// Names returns the sorted list of registered camera names.
func Names() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(registry))

	for name := range registry {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
}
