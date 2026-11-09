# chatd - a chat server

Write `chatd`, a chat server. Start it with `go run .`, then connect clients from other terminals with `nc localhost 4000`.

Each milestone is specified in `acceptance/mN.txt`, as conversations between clients.
Check it with `go test ./acceptance -run MN`. Earlier milestones must keep passing.

## 1. Connections

Accept many clients at the same time, one goroutine each. Welcome each with a name, `guest1`, `guest2`... in connection order, and answer `/ping` and `/quit`.

```
* Welcome guest1!
/ping
pong
```

## 2. The hub

Messages go to everybody else, and arrivals and departures are announced. One goroutine, the hub, owns the list of clients; the others talk to it through channels.

```
reader goroutines ──join/leave/lines──▶ hub ──out channel──▶ writer goroutines
```

## 3. /who and /nick

`/who` lists the connected names, sorted. `/nick NAME` changes your name, if it is valid and free, and everybody is told.

## 4. Private messages

`/msg NAME TEXT` sends TEXT to one client only.

## 5. Ctrl+C

On Ctrl+C, the server tells every client, closes all connections, waits for its goroutines and exits with code 0.

## 6. Slow clients

A client that doesn't read must not block the chat: disconnect it when too many lines are waiting for it.
Try it by hand: freeze an `nc` with Ctrl+Z, and flood from another terminal with `yes hi | nc localhost 4000`.
