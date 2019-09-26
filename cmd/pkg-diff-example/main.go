// Command pkg-diff-example implements a subset of the diff command using
// github.com/pkg/diff.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/diff"
)

var (
	color   = flag.Bool("color", false, "colorize the output")
	timeout = flag.Duration("timeout", 0, "timeout")
	unified = flag.Int("unified", 3, "lines of unified context")
)

// check logs a fatal error and exits if err is not nil.
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// fileLines returns the lines of the file called name.
func fileLines(name string) ([]string, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	return lines, s.Err()
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 2 {
		fmt.Printf("syntax: %s name1 name2\n", filepath.Base(os.Args[0]))
		os.Exit(2)
	}

	aName := flag.Arg(0)
	aLines, err := fileLines(aName)
	check(err)

	bName := flag.Arg(1)
	bLines, err := fileLines(bName)
	check(err)

	ab := diff.Strings(aLines, bLines)
	ctx := context.Background()
	if *timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}
	e := diff.Myers(ctx, ab)
	opts := []diff.WriteOpt{
		diff.Names(aName, bName),
	}
	if *color {
		opts = append(opts, diff.TerminalColor())
	}
	_, err = e.WithContextSize(*unified).WriteUnified(os.Stdout, ab, opts...)
	check(err)
}
