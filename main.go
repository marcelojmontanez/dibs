package main

import (
	"errors"
	"fmt"
	"os"
)

const usage = `usage: dibs <command> [arguments]

commands:
  add <title>    add a task
`

var errUsage = errors.New("usage")

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "dibs:", err)
		if errors.Is(err, errUsage) {
			fmt.Fprint(os.Stderr, usage)
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("%w: missing command", errUsage)
	}

	switch args[0] {
	case "add":
	default:
		return fmt.Errorf("%w: unknown command %q", errUsage, args[0])
	}

	return nil
}
