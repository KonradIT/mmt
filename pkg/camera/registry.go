package camera

import (
	"fmt"
	"sort"
	"sync"
)

var (
	mu       sync.RWMutex
	registry = map[string]Camera{}
)

// Register adds a camera backend to the global registry.
// Typically called from a package's init function.
func Register(c Camera) {
	mu.Lock()
	defer mu.Unlock()
	registry[c.Name()] = c
}

// Get returns the camera registered under the given name.
func Get(name string) (Camera, error) {
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
func Guess(root string) Camera {
	mu.RLock()
	defer mu.RUnlock()
	for _, c := range registry {
		if c.GuessFromPath(root) {
			return c
		}
	}
	return nil
}

// All returns every registered camera, sorted by name.
func All() []Camera {
	mu.RLock()
	defer mu.RUnlock()
	cams := make([]Camera, 0, len(registry))
	for _, c := range registry {
		cams = append(cams, c)
	}
	sort.Slice(cams, func(i, j int) bool { return cams[i].Name() < cams[j].Name() })
	return cams
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
