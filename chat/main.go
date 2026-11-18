// Command chatd is a chat server. Connect to it with: nc localhost 4000
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
)

func main() {
	addr := flag.String("addr", "localhost:4000", "address to listen on")
	flag.Parse()

	// ctx is cancelled when Ctrl+C is pressed.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listening on %s", ln.Addr())
	serve(ctx, ln)
	log.Print("bye")
}

// serve accepts connections until ctx is cancelled, then waits for every
// goroutine it started to finish.
func serve(ctx context.Context, ln net.Listener) {
	ctx, cancel := context.WithCancel(ctx)
	context.AfterFunc(ctx, func() { ln.Close() }) // makes Accept return

	var wg sync.WaitGroup
	h := newHub()
	wg.Go(func() { h.run(ctx) })

	for n := 1; ; n++ {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() == nil {
				log.Print(err)
			}
			break
		}
		c := &client{
			conn: conn,
			name: fmt.Sprintf("guest%d", n),
			out:  make(chan string, queueSize),
		}
		wg.Go(c.write)
		wg.Go(func() { c.read(ctx, h) })
	}
	cancel()
	wg.Wait()
}

// client is one connection. Two goroutines serve it: read and write.
type client struct {
	conn net.Conn
	name string      // read and changed by the hub only
	out  chan string // lines to send to the client; closed by the hub only
}

// read passes everything the client types to the hub.
func (c *client) read(ctx context.Context, h *hub) {
	if !send(ctx, h.join, c) {
		close(c.out) // the hub has stopped and will never close it
		return
	}
	sc := bufio.NewScanner(c.conn)
	for sc.Scan() {
		if !send(ctx, h.lines, line{c, sc.Text()}) {
			return
		}
	}
	send(ctx, h.leave, c)
}

// write sends the client's lines to its connection, until the hub closes
// c.out. Then it closes the connection.
func (c *client) write() {
	defer c.conn.Close()
	for msg := range c.out {
		if _, err := fmt.Fprintln(c.conn, msg); err != nil {
			return
		}
	}
}

// send sends v on ch, unless ctx is cancelled first. It reports whether v was sent.
func send[T any](ctx context.Context, ch chan<- T, v T) bool {
	select {
	case ch <- v:
		return true
	case <-ctx.Done():
		return false
	}
}
