package habit

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"
)

// A store provides a concurrency-safe store for Habits that is persisted to a
// local file.
type store struct {
	path string
	data map[string]Habit
	mtx  sync.Mutex
}

// Get returns the habit with the given name and a bool indicating if the habit
// exists in the store.
func (s *store) Get(name string) (Habit, bool) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	h, ok := s.data[name]
	return h, ok
}

// Add adds or updates the given habit in the store.
func (s *store) Add(h Habit) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.data[h.Name] = h
}

// Delete deletes the habit with the given name from the store. If the
// habit does not exist in the store, then the delete is a no-op.
func (s *store) Delete(name string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	delete(s.data, name)
}

// All returns a list of all habits contained in the store.
func (s *store) All() []Habit {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	var habits []Habit
	for _, hbt := range s.data {
		habits = append(habits, hbt)
	}
	return habits
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
		data: map[string]Habit{},
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
