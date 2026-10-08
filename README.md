# jukebox-queue

A bar's song queue as a small REST service — **the same spec implemented twice,
in Go and in Kotlin**, with the same tests, to compare both languages side by
side.

Tables request songs, vote, and the venue advances the queue. Rules: no
duplicate songs, max 2 pending requests per table, one vote per table per item,
most-voted first (ties keep arrival order), one queue per venue. Everything is
in memory and concurrency-safe. Full spec: [`SPEC.md`](SPEC.md).

| | Go | Kotlin |
| --- | --- | --- |
| HTTP | `net/http` (Go 1.22+ pattern routing) | Ktor 3 + kotlinx.serialization |
| Concurrency | `sync.Mutex` | `@Synchronized` |
| Errors | sentinel errors + `errors.Is` | sealed class hierarchy |
| Tests | `go test -race` | JUnit 5 + Ktor test host |

## Run

```bash
cd go && go test -race ./... && go run ./cmd/server        # :8080
cd kotlin && ./gradlew test && ./gradlew run               # :8081
```

```bash
curl -X POST localhost:8080/venues/bar/queue -d '{"songId":"s1","title":"Song","tableId":"t1"}'
curl localhost:8080/venues/bar/queue
```
