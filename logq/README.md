# logq - search tool for logs

Write `logq`, a command line tool that finds lines in log files, like `grep` but for logs.

A log line has four parts: `TIME LEVEL SERVICE MESSAGE...`, for example:

```
2024-03-01T08:00:04Z ERROR payments timeout calling bank gateway
```

Each milestone is specified in `acceptance/mN.txt`, with commands and their exact output.
Check it with `go test ./acceptance -run MN`. Earlier milestones must keep passing.

## 1. Find lines

Print the lines of a file that contain a word. Exit with code 1 if the file can't be read.

```console
$ logq timeout testdata/app.log
```

## 2. Several files, or standard input

Search the files one after the other. With no file, read standard input.
A file that can't be read is reported, the others are still searched, and the exit code is 1.

```console
$ logq timeout testdata/app.log testdata/old.log
$ cat testdata/app.log | logq timeout
```

## 3. Filters

Parse each line, and add flags to keep only some entries. Lines that aren't log lines have no level and no service.

```console
$ logq -level warn '' testdata/app.log     # WARN and ERROR entries ('' matches every line)
$ logq -service auth -i timeout testdata/app.log
```

## 4. Count and JSON

`-count` prints the number of matching entries. `-json` prints each entry as a JSON object, one per line.

```console
$ logq -count -level error '' testdata/app.log
$ logq -json -service search '' testdata/app.log
```

## 5. Statistics

`-stats` prints the number of matching entries per level and per service, highest first.

```console
$ logq -stats '' testdata/app.log
```

## 6. Compressed files

Files whose name ends with `.gz` are decompressed while reading.

```console
$ logq timeout testdata/archive.log.gz
```

---

Big files to play with: `go run ./loggen -n 1000000 > big.log`
