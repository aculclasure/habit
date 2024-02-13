package habit_test

import (
	"os"
	"testing"

	"github.com/aculclasure/habit"
	"github.com/google/go-cmp/cmp"
)

func TestStore_GetReturnsHabitAndOkIfKeyDoesExist(t *testing.T) {
	t.Parallel()
	store, err := habit.OpenStore("")
	if err != nil {
		t.Fatal(err)
	}
	store.Set("habit1", &habit.Habit{Name: "habit1"})
	got, ok := store.Get("habit1")
	if !ok {
		t.Fatal("wanted ok to be true when getting habit that exists")
	}
	want := &habit.Habit{Name: "habit1"}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestStore_GetReturnsNotOkIfKeyDoesNotExist(t *testing.T) {
	t.Parallel()
	store, err := habit.OpenStore("")
	if err != nil {
		t.Fatal(err)
	}
	_, ok := store.Get("nonexistent-key")
	if ok {
		t.Error("wanted ok to be false when getting non-existent key")
	}
}

func TestStore_SetUpdatesKeyToNewHabit(t *testing.T) {
	t.Parallel()
	store, err := habit.OpenStore("")
	if err != nil {
		t.Fatal(err)
	}
	store.Set("programming-habit", &habit.Habit{
		Name: "beginner-programming-habit",
	})
	store.Set("programming-habit", &habit.Habit{
		Name: "intermediate-programming-habit",
	})
	got, ok := store.Get("programming-habit")
	if !ok {
		t.Fatal("wanted ok to be true when getting habit that exists")
	}
	want := &habit.Habit{
		Name: "intermediate-programming-habit",
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestStore_DeleteDeletesHabitFromStoreIfKeyExists(t *testing.T) {
	t.Parallel()
	store, err := habit.OpenStore("")
	if err != nil {
		t.Fatal(err)
	}
	store.Set("habit1", &habit.Habit{Name: "habit1"})
	store.Set("habit2", &habit.Habit{Name: "habit2"})
	store.Set("habit3", &habit.Habit{Name: "habit3"})
	store.Delete("habit2")
	want := map[string]*habit.Habit{
		"habit1": {Name: "habit1"},
		"habit3": {Name: "habit3"},
	}
	got := store.All()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestStore_DeleteDoesNotModifyStoreGivenNonExistentKey(t *testing.T) {
	t.Parallel()
	store, err := habit.OpenStore("")
	if err != nil {
		t.Fatal(err)
	}
	store.Set("habit1", &habit.Habit{Name: "habit1"})
	store.Set("habit2", &habit.Habit{Name: "habit2"})
	store.Set("habit3", &habit.Habit{Name: "habit3"})
	store.Delete("habit4")
	want := map[string]*habit.Habit{
		"habit1": {Name: "habit1"},
		"habit2": {Name: "habit2"},
		"habit3": {Name: "habit3"},
	}
	got := store.All()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestStore_AllReturnsAllKeysAndHabits(t *testing.T) {
	t.Parallel()
	store, err := habit.OpenStore("")
	if err != nil {
		t.Fatal(err)
	}
	store.Set("habit1", &habit.Habit{Name: "habit1"})
	store.Set("habit2", &habit.Habit{Name: "habit2"})
	store.Set("habit3", &habit.Habit{Name: "habit3"})
	want := map[string]*habit.Habit{
		"habit1": {Name: "habit1"},
		"habit2": {Name: "habit2"},
		"habit3": {Name: "habit3"},
	}
	got := store.All()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestStore_SaveSavesStorePersistently(t *testing.T) {
	t.Parallel()
	path := t.TempDir() + "/temp.store"
	store, err := habit.OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	store.Set("habit1", &habit.Habit{Name: "habit1"})
	store.Set("habit2", &habit.Habit{Name: "habit2"})
	store.Set("habit3", &habit.Habit{Name: "habit3"})
	err = store.Save()
	if err != nil {
		t.Fatal(err)
	}
	store2, err := habit.OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]*habit.Habit{
		"habit1": {Name: "habit1"},
		"habit2": {Name: "habit2"},
		"habit3": {Name: "habit3"},
	}
	got := store2.All()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestStore_SaveReturnsErrorForUnwritablePath(t *testing.T) {
	t.Parallel()
	store, err := habit.OpenStore("fakedir/unwritable.store")
	if err != nil {
		t.Fatal(err)
	}
	err = store.Save()
	if err == nil {
		t.Error("no error")
	}
}

func TestOpenStoreReturnsErrorForInvalidData(t *testing.T) {
	t.Parallel()
	_, err := habit.OpenStore("testdata/empty.store")
	if err == nil {
		t.Error("no error")
	}
}

func TestOpenStoreReturnsErrorForUnreadablePath(t *testing.T) {
	t.Parallel()
	path := t.TempDir() + "/unreadable.store"
	_, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chmod(path, 0000)
	if err != nil {
		t.Fatal(err)
	}
	_, err = habit.OpenStore(path)
	if err == nil {
		t.Error("no error")
	}
}
