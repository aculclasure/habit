package habit_test

import (
	"os"
	"testing"

	"github.com/aculclasure/habit"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestStore_GetReturnsHabitAndOkGivenExistingHabit(t *testing.T) {
	t.Parallel()
	store, err := habit.OpenStore("")
	if err != nil {
		t.Fatal(err)
	}
	store.Add(habit.Habit{Name: "habit1"})
	got, ok := store.Get("habit1")
	if !ok {
		t.Fatal("expected ok to be true when getting habit that exists")
	}
	want := habit.Habit{Name: "habit1"}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestStore_GetReturnsNotOkGivenNonExistentHabit(t *testing.T) {
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

func TestStore_AddUpdatesExistingHabit(t *testing.T) {
	t.Parallel()
	store, err := habit.OpenStore("")
	if err != nil {
		t.Fatal(err)
	}
	store.Add(habit.Habit{
		Name:          "habit",
		CurrentStreak: 1,
	})
	store.Add(habit.Habit{
		Name:          "habit",
		CurrentStreak: 2,
	})
	got, ok := store.Get("habit")
	if !ok {
		t.Fatal("wanted ok to be true when getting habit that exists")
	}
	want := habit.Habit{
		Name:          "habit",
		CurrentStreak: 2,
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

// habitSliceCmpOpt provides a comparison option that allows 2 different slices
// of Habit structs to be compared for equality.
var habitSliceCmpOpt = cmpopts.SortSlices(func(h1, h2 habit.Habit) bool {
	return h1.Name < h2.Name
})

func TestStore_DeleteCorrectlyDeletesExistingHabit(t *testing.T) {
	t.Parallel()
	store, err := habit.OpenStore("")
	if err != nil {
		t.Fatal(err)
	}
	store.Add(habit.Habit{Name: "habit1"})
	store.Add(habit.Habit{Name: "habit2"})
	store.Add(habit.Habit{Name: "habit3"})
	store.Delete("habit2")
	want := []habit.Habit{
		{Name: "habit1"},
		{Name: "habit3"},
	}
	got := store.All()
	if !cmp.Equal(want, got, habitSliceCmpOpt) {
		t.Error(cmp.Diff(want, got, habitSliceCmpOpt))
	}
}

func TestStore_DeleteDoesNotModifyStoreGivenNonExistentHabit(t *testing.T) {
	t.Parallel()
	store, err := habit.OpenStore("")
	if err != nil {
		t.Fatal(err)
	}
	store.Add(habit.Habit{Name: "habit1"})
	store.Add(habit.Habit{Name: "habit2"})
	store.Add(habit.Habit{Name: "habit3"})
	store.Delete("habit4")
	want := []habit.Habit{
		{Name: "habit1"},
		{Name: "habit2"},
		{Name: "habit3"},
	}
	got := store.All()
	if !cmp.Equal(want, got, habitSliceCmpOpt) {
		t.Error(cmp.Diff(want, got, habitSliceCmpOpt))
	}
}

func TestStore_AllReturnsAllHabits(t *testing.T) {
	t.Parallel()
	store, err := habit.OpenStore("")
	if err != nil {
		t.Fatal(err)
	}
	store.Add(habit.Habit{Name: "habit1"})
	store.Add(habit.Habit{Name: "habit2"})
	store.Add(habit.Habit{Name: "habit3"})
	want := []habit.Habit{
		{Name: "habit1"},
		{Name: "habit2"},
		{Name: "habit3"},
	}
	got := store.All()
	if !cmp.Equal(want, got, habitSliceCmpOpt) {
		t.Error(cmp.Diff(want, got, habitSliceCmpOpt))
	}
}

func TestStore_SaveSavesStorePersistently(t *testing.T) {
	t.Parallel()
	path := t.TempDir() + "/temp.store"
	store, err := habit.OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	store.Add(habit.Habit{Name: "habit1"})
	store.Add(habit.Habit{Name: "habit2"})
	store.Add(habit.Habit{Name: "habit3"})
	err = store.Save()
	if err != nil {
		t.Fatal(err)
	}
	store2, err := habit.OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []habit.Habit{
		{Name: "habit1"},
		{Name: "habit2"},
		{Name: "habit3"},
	}
	got := store2.All()
	if !cmp.Equal(want, got, habitSliceCmpOpt) {
		t.Error(cmp.Diff(want, got, habitSliceCmpOpt))
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
		t.Error("expected an error when saving to unwritable path")
	}
}

func TestOpenStoreReturnsErrorForInvalidData(t *testing.T) {
	t.Parallel()
	_, err := habit.OpenStore("testdata/empty.store")
	if err == nil {
		t.Error("expected an error when opening empty store file")
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
		t.Error("expected an error when opening unreadable path")
	}
}
