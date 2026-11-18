// Command miniredis is a small Redis: a server that keeps data in memory and
// shares it between its clients. Connect to it with: nc localhost 6380
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
	"strings"
	"sync"
	"time"
)

func main() {
	addr := flag.String("addr", "localhost:6380", "address to listen on")
	flag.Parse()

	// ctx is cancelled when Ctrl+C is pressed, or when stop is called.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listening on %s", ln.Addr())
	context.AfterFunc(ctx, func() { ln.Close() }) // makes Accept return

	db := newDB()
	var wg sync.WaitGroup
	wg.Go(func() { db.expireLoop(ctx) })
	for {
		conn, err := ln.Accept()
		if err != nil {
			break
		}
		wg.Go(func() { serve(ctx, db, conn) })
	}
	stop()
	wg.Wait()
	log.Print("bye")
}

// serve runs the commands of one client, one line at a time.
func serve(ctx context.Context, db *DB, conn net.Conn) {
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { conn.Close() }) // ends Scan on Ctrl+C
	defer stop()

	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		args := strings.Fields(sc.Text())
		switch {
		case len(args) == 0:
		case strings.EqualFold(args[0], "QUIT"):
			fmt.Fprintln(conn, "OK")
			return
		default:
			reply := db.exec(ctx, args)
			if ctx.Err() != nil {
				return // shutting down: BLPOP stopped waiting, don't answer
			}
			fmt.Fprintln(conn, reply)
		}
	}
}

// DB holds the data. Every command runs with mu locked, so commands never
// see each other half done. (Real Redis does the same: it runs one command
// at a time.)
type DB struct {
	mu      sync.Mutex
	data    map[string]any       // a string, a []string or a map[string]string
	expires map[string]time.Time // when keys expire; keys without expiry are absent
	pushed  chan struct{}        // closed and replaced by every RPUSH, to wake up BLPOP
}

func newDB() *DB {
	return &DB{data: map[string]any{}, expires: map[string]time.Time{}, pushed: make(chan struct{})}
}

// get returns the value of key, if the key holds a T. ok is false if the
// key doesn't exist. db.mu must be locked.
func get[T any](db *DB, key string) (v T, ok bool, err error) {
	db.expireIfDue(key)
	x, ok := db.data[key]
	if !ok {
		return v, false, nil
	}
	v, ok = x.(T)
	if !ok {
		return v, false, errWrongType
	}
	return v, true, nil
}

func (db *DB) remove(key string) {
	delete(db.data, key)
	delete(db.expires, key)
}

func (db *DB) expireIfDue(key string) {
	if t, ok := db.expires[key]; ok && !time.Now().Before(t) {
		db.remove(key)
	}
}

// expireLoop removes the expired keys that nobody asks for, until ctx is
// cancelled.
func (db *DB) expireLoop(ctx context.Context) {
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			db.mu.Lock()
			for key := range db.expires {
				db.expireIfDue(key)
			}
			db.mu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
