package main

import (
	"database/sql"
	"errors"
	"path/filepath"
	"slices"
	"testing"
)

func TestInsertAndSelect(t *testing.T) {
	db := newTestDB(t)

	if err := insertTask(db, "buy milk"); err != nil {
		t.Fatal(err)
	}

	got, err := selectTasks(db)
	if err != nil {
		t.Fatal(err)
	}
	want := []task{{id: 1, title: "buy milk"}}
	if !slices.Equal(got, want) {
		t.Errorf("selectTasks():\n got: %#v\nwant: %#v", got, want)
	}
}

func TestDeleteTask(t *testing.T) {
	db := newTestDB(t)

	if err := insertTask(db, "buy milk"); err != nil {
		t.Fatal(err)
	}

	got, err := deleteTask(db, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := task{id: 1, title: "buy milk"}
	if got != want {
		t.Errorf("deleteTask(1):\n got: %#v\nwant: %#v", got, want)
	}

	tasks, err := selectTasks(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Errorf("selectTasks() after delete = %#v, want no tasks", tasks)
	}
}

func TestDeleteMissingTask(t *testing.T) {
	db := newTestDB(t)

	_, err := deleteTask(db, 1)
	if !errors.Is(err, errNoTask) {
		t.Errorf("deleteTask(1) error = %v, want %v", err, errNoTask)
	}
}

func TestUpdateTask(t *testing.T) {
	db := newTestDB(t)

	if err := insertTask(db, "buy milk"); err != nil {
		t.Fatal(err)
	}

	want := task{1, "call mom", true}
	if err := updateTask(db, want); err != nil {
		t.Fatal(err)
	}

	got, err := selectTask(db, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("selectTask(1) after update:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestUpdateMissingTask(t *testing.T) {
	db := newTestDB(t)

	if err := updateTask(db, task{id: 1}); !errors.Is(err, errNoTask) {
		t.Errorf("updateTask(id 1) error = %v, want %v", err, errNoTask)
	}
}

func TestSelectMissingTask(t *testing.T) {
	db := newTestDB(t)

	_, err := selectTask(db, 1)
	if !errors.Is(err, errNoTask) {
		t.Errorf("selectTask(1) error = %v, want %v", err, errNoTask)
	}
}

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := openDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	return db
}
