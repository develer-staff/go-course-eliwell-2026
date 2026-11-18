package main

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	errWrongType  = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
	errNotInteger = errors.New("ERR value is not an integer or out of range")
)

type command struct {
	minArgs, maxArgs int // not counting the command's name; -1 means no maximum
	run              func(db *DB, args []string) string
}

var commands = map[string]command{
	"PING":    {0, 1, cmdPing},
	"SET":     {2, 2, cmdSet},
	"GET":     {1, 1, cmdGet},
	"DEL":     {1, -1, cmdDel},
	"INCR":    {1, 1, cmdIncr},
	"KEYS":    {1, 1, cmdKeys},
	"DBSIZE":  {0, 0, cmdDBSize},
	"RPUSH":   {2, -1, cmdRPush},
	"LPOP":    {1, 1, cmdLPop},
	"LLEN":    {1, 1, cmdLLen},
	"LRANGE":  {3, 3, cmdLRange},
	"HSET":    {3, 3, cmdHSet},
	"HGET":    {2, 2, cmdHGet},
	"HGETALL": {1, 1, cmdHGetAll},
	"EXPIRE":  {2, 2, cmdExpire},
	"TTL":     {1, 1, cmdTTL},
	"BLPOP":   {2, 2, nil}, // see exec
}

// exec runs a command and returns its reply.
func (db *DB) exec(ctx context.Context, args []string) string {
	name := strings.ToUpper(args[0])
	cmd, ok := commands[name]
	if !ok {
		return fmt.Sprintf("(error) ERR unknown command '%s'", args[0])
	}
	if n := len(args) - 1; n < cmd.minArgs || (cmd.maxArgs >= 0 && n > cmd.maxArgs) {
		return fmt.Sprintf("(error) ERR wrong number of arguments for '%s' command", strings.ToLower(name))
	}
	if name == "BLPOP" {
		return db.blpop(ctx, args[1], args[2]) // it may wait, so it locks by itself
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	return cmd.run(db, args[1:])
}

// The replies look like redis-cli's.

func str(s string) string       { return strconv.Quote(s) }
func integer(n int) string      { return fmt.Sprintf("(integer) %d", n) }
func errReply(err error) string { return "(error) " + err.Error() }

const nilReply = "(nil)"

func array(items []string) string {
	if len(items) == 0 {
		return "(empty array)"
	}
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = fmt.Sprintf("%d) %q", i+1, item)
	}
	return strings.Join(lines, "\n")
}

// ---- strings -------------------------------------------------------------------

func cmdPing(db *DB, args []string) string {
	if len(args) == 1 {
		return str(args[0])
	}
	return "PONG"
}

func cmdSet(db *DB, args []string) string {
	db.remove(args[0]) // forgets the old value, of any type, and its expiry
	db.data[args[0]] = args[1]
	return "OK"
}

func cmdGet(db *DB, args []string) string {
	v, ok, err := get[string](db, args[0])
	switch {
	case err != nil:
		return errReply(err)
	case !ok:
		return nilReply
	}
	return str(v)
}

func cmdDel(db *DB, args []string) string {
	n := 0
	for _, key := range args {
		db.expireIfDue(key)
		if _, ok := db.data[key]; ok {
			db.remove(key)
			n++
		}
	}
	return integer(n)
}

func cmdIncr(db *DB, args []string) string {
	v, ok, err := get[string](db, args[0])
	if err != nil {
		return errReply(err)
	}
	if !ok {
		v = "0"
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return errReply(errNotInteger)
	}
	db.data[args[0]] = strconv.Itoa(n + 1)
	return integer(n + 1)
}

func cmdKeys(db *DB, args []string) string {
	var found []string
	for key := range db.data {
		db.expireIfDue(key)
		if ok, _ := path.Match(args[0], key); ok && db.data[key] != nil {
			found = append(found, key)
		}
	}
	slices.Sort(found)
	return array(found)
}

func cmdDBSize(db *DB, args []string) string { return integer(len(db.data)) }

// ---- lists -----------------------------------------------------------------------

func cmdRPush(db *DB, args []string) string {
	list, _, err := get[[]string](db, args[0])
	if err != nil {
		return errReply(err)
	}
	list = append(list, args[1:]...)
	db.data[args[0]] = list
	close(db.pushed) // wakes up every waiting BLPOP
	db.pushed = make(chan struct{})
	return integer(len(list))
}

// pop removes the first element of the list at key. db.mu must be locked.
func (db *DB) pop(key string) (string, bool, error) {
	list, ok, err := get[[]string](db, key)
	if !ok || err != nil {
		return "", false, err
	}
	if len(list) == 1 {
		db.remove(key) // as in Redis, an empty list doesn't exist
	} else {
		db.data[key] = list[1:]
	}
	return list[0], true, nil
}

func cmdLPop(db *DB, args []string) string {
	v, ok, err := db.pop(args[0])
	switch {
	case err != nil:
		return errReply(err)
	case !ok:
		return nilReply
	}
	return str(v)
}

func cmdLLen(db *DB, args []string) string {
	list, _, err := get[[]string](db, args[0])
	if err != nil {
		return errReply(err)
	}
	return integer(len(list))
}

func cmdLRange(db *DB, args []string) string {
	list, _, err := get[[]string](db, args[0])
	if err != nil {
		return errReply(err)
	}
	start, err1 := strconv.Atoi(args[1])
	stop, err2 := strconv.Atoi(args[2])
	if err1 != nil || err2 != nil {
		return errReply(errNotInteger)
	}
	if start < 0 {
		start += len(list)
	}
	if stop < 0 {
		stop += len(list)
	}
	start, stop = max(start, 0), min(stop, len(list)-1)
	if start > stop {
		return array(nil)
	}
	return array(list[start : stop+1])
}

// blpop is LPOP, but if the list is empty it waits until something is
// pushed, or until the timeout (in seconds, 0 = forever).
func (db *DB) blpop(ctx context.Context, key, timeout string) string {
	secs, err := strconv.Atoi(timeout)
	if err != nil || secs < 0 {
		return errReply(errNotInteger)
	}
	if secs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(secs)*time.Second)
		defer cancel()
	}
	for {
		db.mu.Lock()
		v, ok, err := db.pop(key)
		pushed := db.pushed // taken with the lock held: no RPUSH can slip in between
		db.mu.Unlock()
		switch {
		case err != nil:
			return errReply(err)
		case ok:
			return array([]string{key, v})
		}
		select {
		case <-pushed: // something was pushed somewhere: try again
		case <-ctx.Done():
			return nilReply
		}
	}
}

// ---- hashes ----------------------------------------------------------------------

func cmdHSet(db *DB, args []string) string {
	h, ok, err := get[map[string]string](db, args[0])
	if err != nil {
		return errReply(err)
	}
	if !ok {
		h = map[string]string{}
		db.data[args[0]] = h
	}
	_, existed := h[args[1]]
	h[args[1]] = args[2]
	if existed {
		return integer(0)
	}
	return integer(1)
}

func cmdHGet(db *DB, args []string) string {
	h, _, err := get[map[string]string](db, args[0])
	if err != nil {
		return errReply(err)
	}
	v, ok := h[args[1]] // reading a nil map is fine: it's empty
	if !ok {
		return nilReply
	}
	return str(v)
}

func cmdHGetAll(db *DB, args []string) string {
	h, _, err := get[map[string]string](db, args[0])
	if err != nil {
		return errReply(err)
	}
	var items []string
	for _, field := range slices.Sorted(maps.Keys(h)) {
		items = append(items, field, h[field])
	}
	return array(items)
}

// ---- expiry ------------------------------------------------------------------------

func cmdExpire(db *DB, args []string) string {
	secs, err := strconv.Atoi(args[1])
	if err != nil {
		return errReply(errNotInteger)
	}
	db.expireIfDue(args[0])
	if _, ok := db.data[args[0]]; !ok {
		return integer(0)
	}
	db.expires[args[0]] = time.Now().Add(time.Duration(secs) * time.Second)
	return integer(1)
}

func cmdTTL(db *DB, args []string) string {
	db.expireIfDue(args[0])
	if _, ok := db.data[args[0]]; !ok {
		return integer(-2)
	}
	t, ok := db.expires[args[0]]
	if !ok {
		return integer(-1)
	}
	return integer(int(time.Until(t).Round(time.Second).Seconds()))
}
