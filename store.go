package habit

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"sync"
)

// A store provides a concurrency-safe, key-value store for Habits that is
// persisted to a local file.
type store struct {
	path string
	data map[string]*Habit
	mtx  sync.Mutex
}

// Get returns the given Habit and a bool indicating if the key exists in the
// store.
func (s *store) Get(key string) (Habit, bool) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	h, ok := s.data[key]
	if !ok {
		return Habit{}, ok
	}
	return *h, ok
}

// Set adds or updates the given key in the store.
func (s *store) Set(key string, hb *Habit) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.data[key] = hb
}

// Delete deletes the given key from the store.
func (s *store) Delete(key string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.data, key)
}

// All returns all the keys and values in the store.
func (s *store) All() map[string]*Habit {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return maps.Clone(s.data)
}

// Save saves the store to a GOB-encoded file. An error is returned if there is
// a problem encoding the store's data or saving the store's data to a local
// file.
func (s *store) Save() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	f, err := os.Create(s.path)
	if err != nil {
		return fmt.Errorf("error creating store %q: %w", s.path, err)
	}
	err = gob.NewEncoder(f).Encode(&s.data)
	if err != nil {
		return fmt.Errorf("error encoding habit data to store %q: %w", s.path, err)
	}
	return nil
}

// OpenStore opens the store file at the given path and returns a store
// initialized with the key-value data contained in the file. An error is
// returned if there is a problem opening the store file or decoding its data.
func OpenStore(path string) (*store, error) {
	s := &store{
		path: path,
		data: map[string]*Habit{},
	}
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error opening store %q: %w", path, err)
	}
	defer f.Close()
	err = gob.NewDecoder(f).Decode(&s.data)
	if err != nil {
		return nil, fmt.Errorf("error decoding store data: %w", err)
	}
	return s, nil
}
