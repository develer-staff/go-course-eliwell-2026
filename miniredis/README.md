# miniredis - a small Redis

Write `miniredis`, an in-memory key-value server shared by all its clients. Start it with `go run .`, then connect with `nc localhost 6380`.

```
SET greeting hello
OK
GET greeting
"hello"
```

All the data lives in one map protected by one `sync.Mutex`, and each command runs with the mutex locked.

Each milestone is specified in `acceptance/mN.txt`, as client sessions.
Check it with `go test ./acceptance -run MN`. Earlier milestones must keep passing.

## 1. Strings

`PING`, `SET`, `GET`, `DEL` and `QUIT`, for many clients at once. Commands are case-insensitive, keys are not.

## 2. Counters and keys

`INCR` adds 1 to a number, `KEYS` lists the keys matching a pattern like `user:*`, `DBSIZE` counts the keys.
One check runs 10 000 `INCR` from 50 clients at once: none may be lost.

## 3. Lists

A key can hold a list: `RPUSH`, `LPOP`, `LLEN`, `LRANGE`. Using a key with the wrong kind of value is a `WRONGTYPE` error.

```
RPUSH todo milk eggs
(integer) 2
LRANGE todo 0 -1
1) "milk"
2) "eggs"
```

## 4. Hashes

A key can hold a hash, like a Go map: `HSET`, `HGET`, `HGETALL`.

## 5. Expiry

`EXPIRE key seconds` makes a key disappear after a while, and `TTL` tells how long is left. A background goroutine removes expired keys.

## 6. Ctrl+C

On Ctrl+C, the server closes all connections, stops its background goroutine, waits for its goroutines and exits with code 0.

## 7. BLPOP

`BLPOP key seconds` is `LPOP`, but it waits until another client pushes to the list, or until the timeout. Other clients must keep working meanwhile.
