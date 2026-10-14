package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
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
	{name: "add", args: "<title>", short: "add a task", run: add},
	{name: "list", short: "list all tasks", run: list},
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

func list(db *sql.DB, args []string) error {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	fs.Parse(args)

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
		fmt.Printf("%*d  %s\n", width, t.id, t.title)
	}

	return nil
}
