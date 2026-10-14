package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var errNoTask = errors.New("no such task")

type scanner interface {
	Scan(dest ...any) error
}

// The tasks table. When adding or changing a column, update:
// schema, taskColumns, task, scanTask, and updateTask together.
const schema = `
CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY,
	title TEXT NOT NULL,
	done INTEGER NOT NULL DEFAULT 0
)`

const taskColumns = "id, title, done"

type task struct {
	id    int
	title string
	done  bool
}

func scanTask(s scanner) (task, error) {
	var t task
	err := s.Scan(&t.id, &t.title, &t.done)
	return t, err
}

// updateTask overwrites every field of t.id's row with t's values,
// so t should be a task the caller just read. If two updates to
// the same task run at once, one can silently undo the other.
func updateTask(db *sql.DB, t task) error {
	result, err := db.Exec(
		"UPDATE tasks SET title = ?, done = ? WHERE id = ?",
		t.title, t.done, t.id,
	)
	if err != nil {
		return fmt.Errorf("update task %d: %w", t.id, err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update task %d: %w", t.id, err)
	}
	if n == 0 {
		return fmt.Errorf("update task %d: %w", t.id, errNoTask)
	}

	return nil
}

func insertTask(db *sql.DB, title string) error {
	_, err := db.Exec("INSERT INTO tasks (title) VALUES (?)", title)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}
	return nil
}

func selectTasks(db *sql.DB) ([]task, error) {
	rows, err := db.Query("SELECT " + taskColumns + " FROM tasks ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}
	defer rows.Close()

	var tasks []task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("select tasks: %w", err)
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}

	return tasks, nil
}

func selectTask(db *sql.DB, id int) (task, error) {
	row := db.QueryRow("SELECT "+taskColumns+" FROM tasks WHERE id = ?", id)
	t, err := scanTask(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return task{}, fmt.Errorf("select task %d: %w", id, errNoTask)
		}
		return task{}, fmt.Errorf("select task %d: %w", id, err)
	}
	return t, nil
}

func deleteTask(db *sql.DB, id int) (task, error) {
	row := db.QueryRow("DELETE FROM tasks WHERE id = ? RETURNING "+taskColumns, id)
	t, err := scanTask(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return task{}, fmt.Errorf("delete task %d: %w", id, errNoTask)
		}
		return task{}, fmt.Errorf("delete task %d: %w", id, err)
	}
	return t, nil
}

func openDB(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("open database %s: %w", path, err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create tasks table: %w", err)
	}

	return db, nil
}

func dbPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".local", "share", "dibs", "dibs.db"), nil
}
