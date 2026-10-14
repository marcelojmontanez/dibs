package main

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestInsertAndSelect(t *testing.T) {
	db, err := openDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := insertTask(db, "buy milk"); err != nil {
		t.Fatal(err)
	}

	tasks, err := selectTasks(db)
	if err != nil {
		t.Fatal(err)
	}

	want := []task{{id: 1, title: "buy milk"}}
	if !slices.Equal(tasks, want) {
		t.Errorf("selectTasks():\n got: %#v\nwant: %#v", tasks, want)
	}
}
