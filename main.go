package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type command struct {
	name  string
	args  string
	short string
	run   func(db *sql.DB, args []string) error
}

func (c command) synopsis() string {
	return strings.TrimSpace(c.name + " " + c.args)
}

var commands = []command{
	{name: "add", args: "<title>", short: "add a task", run: add},
}

const schema = `
CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY,
	title TEXT NOT NULL
)`

var errUsage = errors.New("usage")

const helpHint = "run 'dibs help' for usage"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "dibs:", err)
		if errors.Is(err, errUsage) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("%w: missing command\n%s", errUsage, helpHint)
	}

	if args[0] == "help" {
		fmt.Print(usage())
		return nil
	}

	for _, c := range commands {
		if c.name == args[0] {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			if err := c.run(db, args[1:]); err != nil {
				if errors.Is(err, errUsage) {
					return fmt.Errorf("%w\nusage: dibs %s", err, c.synopsis())
				}
				return err
			}
			return nil
		}
	}

	return fmt.Errorf("%w: unknown command %q\n%s", errUsage, args[0], helpHint)
}

func usage() string {
	width := len("help")
	for _, c := range commands {
		width = max(width, len(c.synopsis()))
	}
	width += 4

	var b strings.Builder
	b.WriteString("usage: dibs <command> [arguments]\n\ncommands:\n")
	for _, c := range commands {
		fmt.Fprintf(&b, "  %-*s%s\n", width, c.synopsis(), c.short)
	}
	fmt.Fprintf(&b, "  %-*s%s\n", width, "help", "show this help")
	return b.String()
}

func add(db *sql.DB, args []string) error {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() != 1 {
		return fmt.Errorf("%w: add requires exactly one title", errUsage)
	}

	if err := insertTask(db, fs.Arg(0)); err != nil {
		return err
	}
	fmt.Println("added:", fs.Arg(0))
	return nil
}

func insertTask(db *sql.DB, title string) error {
	_, err := db.Exec("INSERT INTO tasks (title) VALUES (?)", title)
	if err != nil {
		return fmt.Errorf("save task: %w", err)
	}
	return nil
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
