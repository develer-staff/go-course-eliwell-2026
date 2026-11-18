package main

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

// Level is the severity of a log entry. The levels are ordered, so they can
// be compared with < and >=. The zero value is Unknown.
type Level int

const (
	Unknown Level = iota
	Debug
	Info
	Warn
	Error
)

var levelNames = []string{"UNKNOWN", "DEBUG", "INFO", "WARN", "ERROR"}

// String makes fmt print a Level by name.
func (l Level) String() string { return levelNames[l] }

// MarshalText makes encoding/json write a Level by name.
func (l Level) MarshalText() ([]byte, error) { return []byte(l.String()), nil }

// ParseLevel turns "error", "ERROR" or "Error" into Error.
func ParseLevel(s string) (Level, error) {
	i := slices.Index(levelNames, strings.ToUpper(s))
	if i <= 0 {
		return Unknown, fmt.Errorf("unknown level %q", s)
	}
	return Level(i), nil
}

// Entry is one line of a log file.
type Entry struct {
	File    string    `json:"file"`
	Line    int       `json:"line"`
	Time    time.Time `json:"time,omitzero"`
	Level   Level     `json:"level"`
	Service string    `json:"service,omitempty"`
	Message string    `json:"message"`
	Raw     string    `json:"-"` // the whole line, as read
}

// parse splits a line such as
//
//	2024-03-01T08:00:04Z ERROR payments timeout calling bank gateway
//
// into its time, level, service and message. A line in another format gives
// an entry with level Unknown and the whole line as message.
func parse(line string) Entry {
	e := Entry{Raw: line, Message: line}
	parts := strings.SplitN(line, " ", 4)
	if len(parts) < 4 {
		return e
	}
	t, err := time.Parse(time.RFC3339, parts[0])
	if err != nil {
		return e
	}
	level, err := ParseLevel(parts[1])
	if err != nil {
		return e
	}
	e.Time, e.Level, e.Service, e.Message = t, level, parts[2], parts[3]
	return e
}
