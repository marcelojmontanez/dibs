package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY,
	title TEXT NOT NULL
)`

type task struct {
	id    int
	title string
}

func insertTask(db *sql.DB, title string) error {
	_, err := db.Exec("INSERT INTO tasks (title) VALUES (?)", title)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}
	return nil
}

func selectTasks(db *sql.DB) ([]task, error) {
	rows, err := db.Query("SELECT id, title FROM tasks ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}
	defer rows.Close()

	var tasks []task
	for rows.Next() {
		var t task
		if err := rows.Scan(&t.id, &t.title); err != nil {
			return nil, fmt.Errorf("select tasks: %w", err)
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("select tasks: %w", err)
	}

	return tasks, nil
}

func openDB() (*sql.DB, error) {
	path, err := dbPath()
	if err != nil {
		return nil, err
	}

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
