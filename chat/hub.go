package main

import (
	"context"
	"regexp"
	"slices"
	"strings"
)

// queueSize is how many lines can wait for a client. A client with that many
// lines waiting is not reading, and is disconnected.
const queueSize = 256

type line struct {
	from *client
	text string
}

// The hub owns the chat: only its goroutine reads or changes clients and
// their names, so nothing needs a mutex. Other goroutines talk to it through
// the channels.
type hub struct {
	join    chan *client
	leave   chan *client
	lines   chan line
	clients map[*client]bool
}

func newHub() *hub {
	return &hub{
		join:    make(chan *client),
		leave:   make(chan *client),
		lines:   make(chan line),
		clients: map[*client]bool{},
	}
}

func (h *hub) run(ctx context.Context) {
	for {
		select {
		case c := <-h.join:
			h.clients[c] = true
			h.send(c, "* Welcome "+c.name+"!")
			h.broadcast(c, "* "+c.name+" joined")
		case c := <-h.leave:
			h.remove(c)
		case l := <-h.lines:
			h.handle(l.from, strings.TrimSpace(l.text))
		case <-ctx.Done():
			for c := range h.clients {
				select {
				case c.out <- "* server is shutting down":
				default:
				}
				close(c.out)
			}
			return
		}
	}
}

func (h *hub) handle(c *client, text string) {
	if !h.clients[c] {
		return // c has already left, for example after /quit
	}
	cmd, arg, _ := strings.Cut(text, " ")
	switch {
	case text == "":
	case cmd == "/ping":
		h.send(c, "pong")
	case cmd == "/quit":
		h.send(c, "* bye")
		h.remove(c)
	case cmd == "/who":
		h.send(c, "* online: "+strings.Join(h.names(), ", "))
	case cmd == "/nick":
		h.nick(c, strings.TrimSpace(arg))
	case cmd == "/msg":
		h.private(c, arg)
	case strings.HasPrefix(cmd, "/"):
		h.send(c, "! unknown command: "+cmd)
	default:
		h.broadcast(c, c.name+": "+text)
	}
}

// send queues a line for c without ever waiting: a client that is not
// reading must not stop the whole chat, so it is disconnected instead.
func (h *hub) send(c *client, msg string) {
	select {
	case c.out <- msg:
	default:
		h.remove(c)
		c.conn.Close() // its writer may be stuck writing to it
	}
}

// broadcast sends msg to every client except one (none if except is nil).
func (h *hub) broadcast(except *client, msg string) {
	for c := range h.clients {
		if c != except {
			h.send(c, msg)
		}
	}
}

// remove forgets c. Closing c.out makes its writer close the connection,
// once the lines already queued are sent.
func (h *hub) remove(c *client) {
	if !h.clients[c] {
		return
	}
	delete(h.clients, c)
	close(c.out)
	h.broadcast(nil, "* "+c.name+" left")
}

func (h *hub) names() []string {
	var names []string
	for c := range h.clients {
		names = append(names, c.name)
	}
	slices.Sort(names)
	return names
}

func (h *hub) find(name string) *client {
	for c := range h.clients {
		if c.name == name {
			return c
		}
	}
	return nil
}

var validName = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,16}$`)

func (h *hub) nick(c *client, name string) {
	switch {
	case !validName.MatchString(name):
		h.send(c, "! invalid nickname: "+name)
	case h.find(name) != nil:
		h.send(c, "! nickname taken: "+name)
	default:
		h.broadcast(nil, "* "+c.name+" is now known as "+name)
		c.name = name
	}
}

func (h *hub) private(from *client, arg string) {
	name, text, _ := strings.Cut(strings.TrimSpace(arg), " ")
	to := h.find(name)
	switch {
	case name == "" || strings.TrimSpace(text) == "":
		h.send(from, "! usage: /msg NAME TEXT")
	case to == nil:
		h.send(from, "! no such user: "+name)
	default:
		h.send(to, "[private] "+from.name+": "+strings.TrimSpace(text))
	}
}
