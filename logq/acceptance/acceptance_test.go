// Package acceptance checks logq against the scenarios in the m*.txt files.
//
//	go test ./acceptance              every milestone
//	go test ./acceptance -run M2      milestone 2 only
//
// You don't need to read this file: the .txt files are the specification.
package acceptance

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

var logq string // the binary built from the parent directory

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "logq-acceptance")
	if err != nil {
		panic(err)
	}
	logq = filepath.Join(dir, "logq")
	if runtime.GOOS == "windows" {
		logq += ".exe"
	}
	build := exec.Command("go", "build", "-o", logq, ".")
	build.Dir = ".."
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Printf("logq does not build:\n%s", out)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func TestM1(t *testing.T) { runScenarios(t, "m1.txt") }
func TestM2(t *testing.T) { runScenarios(t, "m2.txt") }
func TestM3(t *testing.T) { runScenarios(t, "m3.txt") }
func TestM4(t *testing.T) { runScenarios(t, "m4.txt") }
func TestM5(t *testing.T) { runScenarios(t, "m5.txt") }
func TestM6(t *testing.T) { runScenarios(t, "m6.txt") }

// A check is a "$ logq ..." line of a scenario file, followed by the
// expected output and optionally "[exit N]", up to the next blank line.
type check struct {
	where   string // file:line
	command string
	want    string
	exit    int
}

func runScenarios(t *testing.T, file string) {
	checks, err := readChecks(file)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range checks {
		c.run(t)
	}
}

func readChecks(file string) ([]check, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var checks []check
	var cur *check
	for i, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		switch {
		case strings.HasPrefix(line, "$ "):
			checks = append(checks, check{where: fmt.Sprintf("%s:%d", file, i+1), command: line[2:]})
			cur = &checks[len(checks)-1]
		case cur == nil || strings.HasPrefix(line, "#"):
		case line == "":
			cur = nil
		case strings.HasPrefix(line, "[exit "):
			cur.exit, err = strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(line, "[exit "), "]"))
			if err != nil {
				return nil, fmt.Errorf("%s:%d: bad exit line %q", file, i+1, line)
			}
		default:
			cur.want += line + "\n"
		}
	}
	return checks, nil
}

func (c check) run(t *testing.T) {
	t.Helper()
	args, stdin := splitCommand(c.command)
	if len(args) == 0 || args[0] != "logq" {
		t.Fatalf("%s: commands must start with logq", c.where)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, logq, args[1:]...)
	cmd.Dir = ".."
	if stdin != "" {
		f, err := os.Open(filepath.Join("..", stdin))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		cmd.Stdin = f
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr

	err := cmd.Run()
	exit := 0
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		exit = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("%s: %v", c.where, err)
	}

	got := strings.ReplaceAll(stdout.String(), "\r\n", "\n")
	if got == c.want && exit == c.exit {
		return
	}
	msg := fmt.Sprintf("%s\n$ %s\n", c.where, c.command)
	if got != c.want {
		msg += "--- expected output:\n" + indent(c.want) + "--- got:\n" + indent(got)
	}
	if exit != c.exit {
		msg += fmt.Sprintf("--- expected exit code %d, got %d\n", c.exit, exit)
	}
	if stderr.Len() > 0 {
		msg += "--- stderr:\n" + indent(stderr.String())
	}
	if ctx.Err() != nil {
		msg += "--- logq was still running after 10s and was stopped\n"
	}
	t.Error(msg)
}

// splitCommand splits a command line into words. Single quotes group words
// (” is an empty word), and "< FILE" sets the standard input.
func splitCommand(line string) (args []string, stdin string) {
	var word strings.Builder
	inWord, quoted := false, false
	for _, r := range line + " " {
		switch {
		case r == '\'':
			quoted, inWord = !quoted, true
		case r == ' ' && !quoted:
			if inWord {
				args = append(args, word.String())
				word.Reset()
				inWord = false
			}
		default:
			word.WriteRune(r)
			inWord = true
		}
	}
	if i := len(args) - 2; i >= 0 && args[i] == "<" {
		return args[:i], args[i+1]
	}
	return args, ""
}

func indent(s string) string {
	if s == "" {
		return "    (nothing)\n"
	}
	return "    " + strings.ReplaceAll(strings.TrimSuffix(s, "\n"), "\n", "\n    ") + "\n"
}
