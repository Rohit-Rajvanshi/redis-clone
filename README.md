# redis-clone

[![progress-banner](https://backend.codecrafters.io/progress/redis/56a56390-c651-4399-932f-4b47d01f150a)](https://app.codecrafters.io/users/Rohit-Rajvanshi?r=2qF)

A Redis server written from scratch in Go. It speaks the real Redis wire
protocol (RESP), so you can talk to it with the official `redis-cli`.

Built as part of the [CodeCrafters "Build Your Own Redis"](https://codecrafters.io/challenges/redis) challenge.

## Features

- TCP server on port `6379`
- Multiple concurrent clients, with one goroutine per connection
- A shared in-memory key-value store protected by a `sync.RWMutex`
- Commands:

| Command | Notes |
|---------|-------|
| `PING` | Replies `PONG` |
| `ECHO <msg>` | Replies with the message |
| `SET <key> <value> [PX ms]` | Optional expiry in milliseconds |
| `GET <key>` | Returns the value, or nil if it's missing or expired |

### In progress

- [ ] Lists (`RPUSH`, ...)

## Running it

Requires Go 1.27+.

```sh
./your_program.sh
```

Then, from another terminal:

```sh
redis-cli PING
redis-cli SET fruit mango PX 5000
redis-cli GET fruit
```

## Project layout

```
app/main.go        server, connection handling, command dispatch
your_program.sh    builds and runs the server locally
```

## What I learned

### Lazy expiry

**Lazy expiry.** An expired key is only removed when someone tries to read it: GET checks the expiry time and deletes the key if it has passed. That's cheap, because no background work is needed. The weakness is that an expired key nobody ever reads again stays in memory forever. Real Redis also runs **active expiry**: about 10 times per second, it samples random keys that have an expiry and deletes the expired ones. My server only does lazy expiry for now.

### Locking the shared store

**Check-then-act race.** In GET, I read the key under a read lock, released the lock, and then deleted it under a write lock if it looked expired. In the gap between those two locks, another client could SET the same key to a fresh value, and my delete would wipe it out. The fix: once I hold the write lock, re-read the key *from the map* and check expiry again before deleting. Re-checking the copy I read earlier isn't enough, because that copy is stale.

### Go slices and `append`

**A slice is a small header, not the data.** A slice value is a struct holding `{pointer to an array, len, cap}`. Copying a slice (for example, with `v := entry.val.([]string)`) copies that header. Both copies point at the same array, but each one has its *own* `len`. My first RPUSH did `v = append(v, x)` on the local copy, so the slice stored in the map never grew. Even when `append` writes into the same shared array, the map's header still has the old `len`, so it never "sees" the new element. If the array runs out of room, `append` allocates a new one, and then the two slices don't even share memory. The rule: **always store the result of `append` back where the slice lives.** The fix was to append into `v`, store `v` back into the map, and use `len(v)` for the reply, since `len` doesn't work on an `any`.
