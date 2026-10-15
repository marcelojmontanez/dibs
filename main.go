package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
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
	{name: "add", args: "<title>", short: "add a task", run: runAdd},
	{name: "list", short: "list all tasks", run: runList},
	{name: "delete", args: "<id>", short: "delete a task", run: runDelete},
	{name: "done", args: "<id>", short: "mark a task done", run: runDone},
	{name: "edit", args: "<id> <title>", short: "edit a task's title", run: runEdit},
}

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
			path, err := dbPath()
			if err != nil {
				return err
			}

			db, err := openDB(path)
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

func runAdd(db *sql.DB, args []string) error {
	fs := newFlagSet("add")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%w: %w", errUsage, err)
	}

	if fs.NArg() != 1 {
		return fmt.Errorf("%w: add requires exactly one title", errUsage)
	}

	if err := insertTask(db, fs.Arg(0)); err != nil {
		return err
	}
	fmt.Println("added:", fs.Arg(0))
	return nil
}

func runList(db *sql.DB, args []string) error {
	fs := newFlagSet("list")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%w: %w", errUsage, err)
	}

	if fs.NArg() != 0 {
		return fmt.Errorf("%w: list takes no arguments", errUsage)
	}

	tasks, err := selectTasks(db)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		return nil
	}

	width := len(strconv.Itoa(tasks[len(tasks)-1].id))
	for _, t := range tasks {
		mark := " "
		if t.done {
			mark = "x"
		}
		fmt.Printf("%*d  [%s]  %s\n", width, t.id, mark, t.title)
	}

	return nil
}

func runDelete(db *sql.DB, args []string) error {
	fs := newFlagSet("delete")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%w: %w", errUsage, err)
	}

	if fs.NArg() != 1 {
		return fmt.Errorf("%w: delete requires exactly one id", errUsage)
	}

	id, err := parseID(fs.Arg(0))
	if err != nil {
		return err
	}

	t, err := deleteTask(db, id)
	if err != nil {
		return err
	}
	fmt.Println("deleted:", t.title)
	return nil
}

func runDone(db *sql.DB, args []string) error {
	fs := newFlagSet("done")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%w: %w", errUsage, err)
	}

	if fs.NArg() != 1 {
		return fmt.Errorf("%w: done requires exactly one id", errUsage)
	}

	id, err := parseID(fs.Arg(0))
	if err != nil {
		return err
	}

	t, err := selectTask(db, id)
	if err != nil {
		return err
	}
	if t.done {
		fmt.Println("already done:", t.title)
		return nil
	}

	t.done = true
	if err := updateTask(db, t); err != nil {
		return err
	}
	fmt.Println("completed:", t.title)
	return nil
}

func runEdit(db *sql.DB, args []string) error {
	fs := newFlagSet("edit")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("%w: %w", errUsage, err)
	}

	if fs.NArg() != 2 {
		return fmt.Errorf("%w: edit requires an id and a new title", errUsage)
	}

	id, err := parseID(fs.Arg(0))
	if err != nil {
		return err
	}

	t, err := selectTask(db, id)
	if err != nil {
		return err
	}

	oldTitle := t.title
	t.title = fs.Arg(1)
	if err := updateTask(db, t); err != nil {
		return err
	}
	fmt.Printf("edited: %s -> %s\n", oldTitle, t.title)
	return nil
}

func parseID(s string) (int, error) {
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("%w: task id must be an integer, got %q", errUsage, s)
	}
	return id, nil
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}
