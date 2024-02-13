package habit_test

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/aculclasure/habit"
	"github.com/google/go-cmp/cmp"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestTracker_UpsertReturnsErrorForHabitLastUpdatedInTheFuture(t *testing.T) {
	lastDone, err := time.Parse(time.RFC3339, "2024-02-06T13:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	store, err := habit.OpenStore("")
	if err != nil {
		t.Fatal(err)
	}
	store.Set("programming", &habit.Habit{
		Name:     "programming",
		LastDone: lastDone,
	})
	tracker, err := habit.NewTracker(habit.WithStore(store))
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = getTimeFunc(t, "2024-02-05T13:00:00Z")
	err = tracker.Upsert("programming")
	if err == nil {
		t.Error("expected an error when upserting a habit that takes place in the future")
	}
}

func TestTracker_UpsertDoesNotModifyStreakMoreThanOnceOnSameCalendarDay(t *testing.T) {
	lastDone, err := time.Parse(time.RFC3339, "2024-02-06T13:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/test.store"
	store, err := habit.OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	store.Set("programming", &habit.Habit{
		Name:          "programming",
		CurrentStreak: 7,
		LastDone:      lastDone,
	})
	output := io.Discard
	tracker, err := habit.NewTracker(habit.WithStore(store), habit.WithOutput(output))
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = getTimeFunc(t, "2024-02-06T13:05:00Z")
	err = tracker.Upsert("programming")
	if err != nil {
		t.Fatal(err)
	}
	want := &habit.Habit{
		Name:          "programming",
		CurrentStreak: 7,
		LastDone:      habit.Now(),
	}
	got, ok := store.Get("programming")
	if !ok {
		t.Fatal("expected habit 'programming' to be present in store")
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
}

func TestTracker_UpsertResetsStreakForHabitsOneOrMoreDaysOld(t *testing.T) {
	habit.Now = getTimeFunc(t, "2024-02-06T13:05:00Z")
	programmingLastDone, err := time.Parse(time.RFC3339, "2024-02-04T13:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	exercisingLastDone, err := time.Parse(time.RFC3339, "2024-02-05T13:05:00Z")
	if err != nil {
		t.Fatal(err)
	}
	testCases := map[string]struct {
		input      *habit.Habit
		wantHabit  *habit.Habit
		wantOutput string
	}{
		"Habit last done more than 1 day ago resets streak": {
			input: &habit.Habit{
				Name:          "programming",
				CurrentStreak: 5,
				LastDone:      programmingLastDone,
			},
			wantHabit: &habit.Habit{
				Name:          "programming",
				CurrentStreak: 1,
				LastDone:      habit.Now(),
			},
			wantOutput: "You last did the habit 'programming' 2 days ago, so you're starting a new streak today. Good luck!\n",
		},
		"Habit last done exactly 1 day ago resets streak": {
			input: &habit.Habit{
				Name:          "exercising",
				CurrentStreak: 5,
				LastDone:      exercisingLastDone,
			},
			wantHabit: &habit.Habit{
				Name:          "exercising",
				CurrentStreak: 1,
				LastDone:      habit.Now(),
			},
			wantOutput: "You last did the habit 'exercising' 1 day ago, so you're starting a new streak today. Good luck!\n",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			store, err := habit.OpenStore(t.TempDir() + "/test.store")
			if err != nil {
				t.Fatal(err)
			}
			store.Set(tc.input.Name, tc.input)
			output := new(bytes.Buffer)
			tracker, err := habit.NewTracker(habit.WithOutput(output), habit.WithStore(store))
			if err != nil {
				t.Fatal(err)
			}
			err = tracker.Upsert(tc.input.Name)
			if err != nil {
				t.Fatal(err)
			}
			gotHabit, ok := store.Get(tc.input.Name)
			if !ok {
				t.Fatalf("expected habit with name '%s' to be present in store", tc.input.Name)
			}
			if !cmp.Equal(tc.wantHabit, gotHabit) {
				t.Fatal(cmp.Diff(tc.wantHabit, gotHabit))
			}
			gotOutput := output.String()
			if tc.wantOutput != gotOutput {
				t.Fatalf("want output %q, got output %q", tc.wantOutput, gotOutput)
			}
		})
	}
}

func TestTracker_UpsertCorrectlyIncrementsStreakForHabitLessThan1DayOld(t *testing.T) {
	lastDone, err := time.Parse(time.RFC3339, "2024-02-05T13:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/test.store"
	store, err := habit.OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	store.Set("programming", &habit.Habit{
		Name:          "programming",
		CurrentStreak: 1,
		LastDone:      lastDone,
	})
	output := new(bytes.Buffer)
	tracker, err := habit.NewTracker(habit.WithStore(store), habit.WithOutput(output))
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = getTimeFunc(t, "2024-02-06T12:59:00Z")
	err = tracker.Upsert("programming")
	if err != nil {
		t.Fatal(err)
	}
	want := &habit.Habit{
		Name:          "programming",
		CurrentStreak: 2,
		LastDone:      habit.Now(),
	}
	got, ok := store.Get("programming")
	if !ok {
		t.Fatal("expected habit 'programming' to be present in store")
	}
	if !cmp.Equal(want, got) {
		t.Fatal(cmp.Diff(want, got))
	}
	wantOutput := "Nice work: you've done the habit 'programming' for 2 days in a row now.\n"
	gotOutput := output.String()
	if wantOutput != gotOutput {
		t.Errorf("want output %q, got output %q", wantOutput, gotOutput)
	}
}

func TestTracker_PrintSummaryPrintsExpectedMessageForHabitsWithExpiredStreaks(t *testing.T) {
	path := t.TempDir() + "/test.store"
	store, err := habit.OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = getTimeFunc(t, "2024-02-06T13:00:00Z")
	oneDayAgo, err := time.Parse(time.RFC3339, "2024-02-05T12:30:00Z")
	if err != nil {
		t.Fatal(err)
	}
	store.Set("programming", &habit.Habit{
		Name:          "programming",
		CurrentStreak: 1,
		LastDone:      oneDayAgo,
	})
	threeDaysAgo, err := time.Parse(time.RFC3339, "2024-02-03T13:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	store.Set("exercising", &habit.Habit{
		Name:          "exercising",
		CurrentStreak: 4,
		LastDone:      threeDaysAgo,
	})
	output := new(bytes.Buffer)
	tracker, err := habit.NewTracker(habit.WithOutput(output), habit.WithStore(store))
	if err != nil {
		t.Fatal(err)
	}
	tracker.PrintSummary()
	wantSubstrings := []string{
		"It's been 1 day since you did 'programming'. Stay positive and get back on it!\n",
		"It's been 3 days since you did 'exercising'. Stay positive and get back on it!\n",
	}
	got := output.String()
	for _, w := range wantSubstrings {
		if !strings.Contains(got, w) {
			t.Errorf("wanted output to contain %s, got output %s", w, got)
		}
	}
}

func TestMain(m *testing.M) {
	testscript.RunMain(m, map[string]func() int{
		"habit": habit.Main,
	})
}

func Test(t *testing.T) {
	habit.Now = getTimeFunc(t, "2024-01-02T00:00:30Z")
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
	})
}

func getTimeFunc(t *testing.T, timestamp string) func() time.Time {
	t.Helper()
	return func() time.Time {
		testTime, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			t.Fatal(err)
		}
		return testTime
	}
}
