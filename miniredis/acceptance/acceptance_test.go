// Package acceptance checks miniredis against the scenarios in the m*.txt files.
//
//	go test ./acceptance              every milestone
//	go test ./acceptance -run M2      milestone 2 only
//
// You don't need to read this file: the .txt files are the specification.
package acceptance

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

const program = "miniredis"

var bin string // the program, built from the parent directory

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", program)
	if err != nil {
		panic(err)
	}
	bin = filepath.Join(dir, program)
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	// -race makes the program report data races. It needs cgo, which some
	// machines don't have: then we build without it.
	if _, err := build("-race"); err != nil {
		out, err := build()
		if err != nil {
			fmt.Printf("%s does not build:\n%s", program, out)
			os.Exit(1)
		}
		fmt.Println("note: the race detector is not available on this machine; data races won't be reported")
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func build(flags ...string) ([]byte, error) {
	cmd := exec.Command("go", append(append([]string{"build"}, flags...), "-o", bin, ".")...)
	cmd.Dir = ".."
	return cmd.CombinedOutput()
}

func TestM1(t *testing.T) { runScenarios(t, "m1.txt") }
func TestM2(t *testing.T) { runScenarios(t, "m2.txt") }
func TestM3(t *testing.T) { runScenarios(t, "m3.txt") }
func TestM4(t *testing.T) { runScenarios(t, "m4.txt") }
func TestM5(t *testing.T) { runScenarios(t, "m5.txt") }
func TestM6(t *testing.T) { runScenarios(t, "m6.txt") }
func TestM7(t *testing.T) { runScenarios(t, "m7.txt") }

// ---- the server --------------------------------------------------------------

type server struct {
	t      *testing.T
	addr   string
	cmd    *exec.Cmd
	stderr bytes.Buffer // read only once the program has exited
	exited chan struct{}
	err    error // what the program exited with
}

func startServer(t *testing.T) *server {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &server{t: t, addr: ln.Addr().String(), exited: make(chan struct{})}
	ln.Close() // the port is free again, for the program to use

	s.cmd = exec.Command(bin, "-addr", s.addr)
	s.cmd.Stderr = &s.stderr
	if err := s.cmd.Start(); err != nil {
		t.Fatal(err)
	}
	go func() {
		s.err = s.cmd.Wait()
		close(s.exited)
	}()
	t.Cleanup(s.stop)
	return s
}

func (s *server) stop() {
	select {
	case <-s.exited:
	default:
		s.cmd.Process.Kill()
		<-s.exited
	}
	if strings.Contains(s.stderr.String(), "DATA RACE") {
		s.t.Errorf("the race detector found a data race in %s:\n%s", program, s.stderr.String())
	}
}

func (s *server) hasExited() bool {
	select {
	case <-s.exited:
		return true
	default:
		return false
	}
}

// connect dials the server, retrying while it starts up.
func (s *server) connect() (*client, error) {
	deadline := time.Now().Add(5 * time.Second)
	for {
		conn, err := net.Dial("tcp", s.addr)
		if err == nil {
			return &client{conn: conn, r: bufio.NewReader(conn)}, nil
		}
		if s.hasExited() {
			how := "normally"
			if s.err != nil {
				how = s.err.Error()
			}
			return nil, fmt.Errorf("%s is not running: it exited (%s). What it printed:\n%s", program, how, s.stderr.String())
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("can't connect to %s: %v", program, err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

type client struct {
	conn net.Conn
	r    *bufio.Reader
}

func (c *client) send(text string) error {
	c.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_, err := io.WriteString(c.conn, text+"\n")
	return err
}

// receive reads the next line, waiting at most d.
func (c *client) receive(d time.Duration) (string, error) {
	c.conn.SetReadDeadline(time.Now().Add(d))
	s, err := c.r.ReadString('\n')
	return strings.TrimRight(s, "\r\n"), err
}

func isTimeout(err error) bool {
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

// ---- scenarios ------------------------------------------------------------------

type step struct {
	where string // file:line
	text  string // the line of the scenario
	who   string // the client, or "" for the steps about the server
	do    string // ">", "<", "hears nothing", "hangs up", "is disconnected", "Ctrl+C", "exits", "wait"
	arg   string
}

type scenario struct {
	title string
	steps []step
}

func runScenarios(t *testing.T, file string) {
	scenarios, err := readScenarios(file)
	if err != nil {
		t.Fatal(err)
	}
	for _, sc := range scenarios {
		t.Run(sc.title, func(t *testing.T) { sc.run(t) })
	}
}

func readScenarios(file string) ([]scenario, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var scenarios []scenario
	for i, text := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		where := fmt.Sprintf("%s:%d", file, i+1)
		if title, ok := strings.CutPrefix(text, "## "); ok {
			scenarios = append(scenarios, scenario{title: title})
			continue
		}
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		if len(scenarios) == 0 {
			return nil, fmt.Errorf("%s: a scenario must start with a \"## title\" line", where)
		}
		st, err := parseStep(text)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", where, err)
		}
		st.where, st.text = where, text
		last := &scenarios[len(scenarios)-1]
		last.steps = append(last.steps, st)
	}
	return scenarios, nil
}

func parseStep(text string) (step, error) {
	switch {
	case text == "press Ctrl+C":
		return step{do: "Ctrl+C"}, nil
	case text == "server exits":
		return step{do: "exits"}, nil
	case strings.HasPrefix(text, "wait "):
		return step{do: "wait", arg: strings.TrimPrefix(text, "wait ")}, nil
	}
	who, rest, _ := strings.Cut(text, " ")
	switch {
	case rest == ">" || rest == "<":
		return step{who: who, do: rest}, nil
	case strings.HasPrefix(rest, "> "), strings.HasPrefix(rest, "< "):
		return step{who: who, do: rest[:1], arg: rest[2:]}, nil
	case rest == "hears nothing", rest == "hangs up", rest == "is disconnected":
		return step{who: who, do: rest}, nil
	}
	return step{}, fmt.Errorf("don't know what to do with %q", text)
}

func (sc scenario) run(t *testing.T) {
	s := startServer(t)
	clients := map[string]*client{}
	for _, st := range sc.steps {
		fail := func(format string, args ...any) {
			t.Fatalf("%s: %s\n    %s", st.where, st.text, fmt.Sprintf(format, args...))
		}
		c := clients[st.who]
		if c == nil && st.who != "" {
			var err error
			if c, err = s.connect(); err != nil {
				fail("%v", err)
			}
			defer c.conn.Close()
			clients[st.who] = c
		}

		switch st.do {
		case ">":
			if err := c.send(st.arg); err != nil {
				fail("could not send: %v", err)
			}
		case "<":
			got, err := c.receive(2 * time.Second)
			switch {
			case isTimeout(err):
				fail("%s received nothing for 2 seconds", st.who)
			case err != nil:
				fail("the connection was closed (%v)", err)
			case got != st.arg:
				fail("%s received %q instead", st.who, got)
			}
		case "hears nothing":
			if got, err := c.receive(300 * time.Millisecond); !isTimeout(err) {
				if err != nil {
					fail("the connection was closed (%v)", err)
				}
				fail("%s received %q", st.who, got)
			}
		case "hangs up":
			c.conn.Close()
		case "is disconnected":
			got, err := c.receive(2 * time.Second)
			switch {
			case err == nil:
				fail("%s received %q", st.who, got)
			case isTimeout(err):
				fail("the connection is still open 2 seconds later")
			}
		case "Ctrl+C":
			if runtime.GOOS == "windows" {
				t.Skip("this scenario presses Ctrl+C, which can't be simulated on Windows: try it by hand")
			}
			if len(clients) == 0 {
				// Make sure the program is up (and handles Ctrl+C) before pressing it.
				probe, err := s.connect()
				if err != nil {
					fail("%v", err)
				}
				probe.conn.Close()
				time.Sleep(100 * time.Millisecond)
			}
			s.cmd.Process.Signal(os.Interrupt)
		case "exits":
			select {
			case <-s.exited:
				if s.err != nil {
					fail("%s exited with an error: %v", program, s.err)
				}
			case <-time.After(3 * time.Second):
				fail("%s is still running 3 seconds after Ctrl+C", program)
			}
		case "wait":
			d, err := time.ParseDuration(st.arg)
			if err != nil {
				fail("%v", err)
			}
			time.Sleep(d)
		}
	}
}

// ---- checks that are easier to write in Go than as scenarios -------------------

// Twenty clients write and read their own keys at the same time.
func TestM1_ManyClientsAtOnce(t *testing.T) {
	s := startServer(t)
	var wg sync.WaitGroup
	for i := range 20 {
		wg.Go(func() {
			c, err := s.connect()
			if err != nil {
				t.Error(err)
				return
			}
			defer c.conn.Close()
			for j := range 50 {
				key, value := fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", j)
				c.send("SET " + key + " " + value)
				if got, err := c.receive(5 * time.Second); got != "OK" {
					t.Errorf("SET %s %s: got %q (%v)", key, value, got, err)
					return
				}
				c.send("GET " + key)
				if got, err := c.receive(5 * time.Second); got != `"`+value+`"` {
					t.Errorf("GET %s: got %q (%v), want %q", key, got, err, `"`+value+`"`)
					return
				}
			}
		})
	}
	wg.Wait()
}

// Fifty clients run INCR on the same key 200 times each, at the same time:
// not a single increment may be lost.
func TestM2_NoIncrementIsLost(t *testing.T) {
	s := startServer(t)
	const clients, each = 50, 200
	var wg sync.WaitGroup
	for range clients {
		wg.Go(func() {
			c, err := s.connect()
			if err != nil {
				t.Error(err)
				return
			}
			defer c.conn.Close()
			for range each {
				c.send("INCR counter")
				if got, err := c.receive(5 * time.Second); !strings.HasPrefix(got, "(integer) ") {
					t.Errorf("INCR counter: got %q (%v)", got, err)
					return
				}
			}
		})
	}
	wg.Wait()

	c, err := s.connect()
	if err != nil {
		t.Fatal(err)
	}
	defer c.conn.Close()
	c.send("GET counter")
	if got, _ := c.receive(5 * time.Second); got != fmt.Sprintf(`"%d"`, clients*each) {
		t.Errorf("after %d INCR, GET counter = %s: increments were lost", clients*each, got)
	}
}
