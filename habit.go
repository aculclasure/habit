// Package habit implements a habit tracker and accompanying CLI.
package habit

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

// Now provides a seam to allow the time.Now() function to be overriden for
// testing.
var Now = time.Now

// A Habit represents a habit that can be tracked.
type Habit struct {
	// Name is the name of the habit.
	Name string
	// CurrentStreak is the number of days in a row this habit has
	// been performed.
	CurrentStreak int
	// LastDone is the timestamp when the habit was last done.
	LastDone time.Time
}

// A Tracker provides habit-tracking and summarization logic.
type Tracker struct {
	// output is the io.Writer to write the habit summary output to.
	output io.Writer
	// store is the data repository that stores Habits.
	store *store
}

// option provides a functional option that can be used in the NewTracker()
// function.
type option func(*Tracker) error

// WithOutput accepts an io.Writer and returns an option that wires the io.Writer
// to a Tracker.
func WithOutput(output io.Writer) option {
	return func(t *Tracker) error {
		if output == nil {
			return errors.New("output writer must be non-nil")
		}
		t.output = output
		return nil
	}
}

// WithStore accepts a store and returns an option that wires the store to a
// Tracker.
func WithStore(store *store) option {
	return func(t *Tracker) error {
		if store == nil {
			return errors.New("habit store must be non-nil")
		}
		t.store = store
		return nil
	}
}

// NewTracker accepts an optional list of options and returns a Tracker
// initialized with these options. If no options are provided, the Tracker
// stores its data to a local file "habit.store" and writes to stdout. An error
// is returned if there is a problem opening the data store or if any of the
// opts returns an error.
func NewTracker(opts ...option) (*Tracker, error) {
	s, err := OpenStore("habit.store")
	if err != nil {
		return nil, err
	}
	t := &Tracker{
		output: os.Stdout,
		store:  s,
	}
	for _, opt := range opts {
		err := opt(t)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

// Track adds a new Habit to the store or updates an already-existing Habit in
// the store. An error is returned if an update is attempted on a Habit with a
// timestamp in the future or if the store cannot be saved after adding/updating
// a Habit.
func (t *Tracker) Track(hbtName string) error {
	now := Now()
	hbt, ok := t.store.data[hbtName]
	if !ok {
		t.store.Set(hbtName, &Habit{
			Name:          hbtName,
			CurrentStreak: 1,
			LastDone:      now,
		})
		err := t.store.Save()
		if err != nil {
			return err
		}
		fmt.Fprintf(t.output, "Congratulations on starting your new habit '%s'! Don't forget to do it again.\n", hbtName)
		return nil
	}
	dayOutput := "days"
	daysSince := int(now.Sub(hbt.LastDone).Hours() / 24)
	if daysSince == 1 {
		dayOutput = "day"
	}
	switch {
	case now.Before(hbt.LastDone):
		return fmt.Errorf("current time %q cannot precede last time habit '%s' was updated on %q",
			now.Format(time.RFC3339),
			hbtName,
			hbt.LastDone.Format(time.RFC3339))
	case sameDate(now, hbt.LastDone):
		fmt.Fprintf(t.output, "Way to go practicing your habit '%s' more than once today!\n",
			hbtName)
	case daysSince > 0:
		hbt.CurrentStreak = 1
		fmt.Fprintf(t.output, "You last did the habit '%s' %d %s ago, so you're starting a new streak today. Good luck!\n",
			hbtName, daysSince, dayOutput)
	default:
		hbt.CurrentStreak++
		fmt.Fprintf(t.output, "Nice work: you've done the habit '%s' for %d %s in a row now.\n",
			hbtName, hbt.CurrentStreak, dayOutput)
	}
	hbt.LastDone = now
	err := t.store.Save()
	if err != nil {
		return err
	}
	return nil
}

// PrintSummary writes a summary of tracked Habits to the given Tracker's output.
func (t Tracker) PrintSummary() {
	if len(t.store.data) < 1 {
		fmt.Fprintln(t.output, "You're not currently tracking any habits.")
		return
	}
	now := Now()
	for _, hbt := range t.store.data {
		daysSince := int(now.Sub(hbt.LastDone).Hours() / 24)
		if daysSince > 0 {
			dayOutput := "days"
			if daysSince == 1 {
				dayOutput = "day"
			}
			fmt.Fprintf(t.output, "It's been %d %s since you did '%s'. Stay positive and get back on it!\n",
				daysSince, dayOutput, hbt.Name)
			continue
		}
		fmt.Fprintf(t.output, "You are currently on a %d-day streak for '%s'. Keep it going!\n",
			hbt.CurrentStreak, hbt.Name)
	}
}

// Main is the driver for the CLI. It reads command-line arguments and allows
// a new Habit to be added, an existing Habit to be updated, or a summary of all
// stored Habits to be printed. It returns an exit code where 0 means the
// command was successful and anything other than 0 means the command failed.
func Main() int {
	flag.Usage = func() {
		fmt.Println(`Usage: habit <habit-name>

habit is a tool that helps users track and establish a new habit, by reporting
their current streak. 
			
The default store file is 'habit.store'. This file will be
created automatically the first time a habbit is set using
'habit <habit-name>'.`)
	}
	flag.Parse()
	tracker, err := NewTracker()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if len(os.Args) > 1 {
		err = tracker.Track(os.Args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	}
	tracker.PrintSummary()
	return 0
}

// sameDate accepts 2 timestamps and returns true if they occur on the same
// calendar date.
func sameDate(t1, t2 time.Time) bool {
	t1Year, t1Month, t1Day := t1.Date()
	t2Year, t2Month, t2Day := t2.Date()
	return (t2Year == t1Year) && (t2Month == t1Month) && (t2Day == t1Day)
}
