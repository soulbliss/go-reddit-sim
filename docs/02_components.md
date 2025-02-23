# System Components Deep Dive

## 1. Event Generator (`generateEvents`)

The Event Generator simulates user activity on a platform like Reddit:

```go
func generateEvents(eventChan chan<- map[string]interface{}, metrics *RedditMetrics, quit <-chan bool)
```

### How it works:
- Generates events every 100ms (10 events per second)
- Simulates different types of actions: posts, comments, upvotes, downvotes
- Uses channels for non-blocking communication
- Updates metrics in a thread-safe way using mutexes

### Aha Moment! 🎉
The generator never waits for the database or processor - it keeps generating events regardless of what happens downstream, just like real users don't wait for the database to save their actions!

## 2. Database Writer (`storeEvents`)

The Database Writer persists events to PostgreSQL:

```go
func storeEvents(db *sql.DB, eventChan <-chan map[string]interface{}, metrics *RedditMetrics, quit <-chan bool)
```

### How it works:
- Listens continuously for new events
- Converts events to JSON for storage
- Uses parameterized SQL queries for safety
- Tracks performance metrics for each operation

### Aha Moment! 🎉
The writer uses Go's built-in JSON marshaling to store complex data structures in PostgreSQL's JSONB format - this means we can store any type of event without changing our database schema!

## 3. Event Processor (`processEvents`)

The Event Processor handles batched updates:

```go
func processEvents(db *sql.DB, metrics *RedditMetrics, quit <-chan bool)
```

### How it works:
- Processes events in batches every 200ms
- Uses `SELECT ... FOR UPDATE SKIP LOCKED` for concurrent safety
- Updates multiple records in a single transaction
- Prevents duplicate processing through database locks

### Aha Moment! 🎉
The `SKIP LOCKED` feature allows multiple processors to work simultaneously without conflicts - it's like multiple checkout lines in a supermarket, each processor can grab its own batch of events!

## 4. Metrics Visualizer (`visualizeMetrics`)

The Metrics Visualizer creates a real-time dashboard:

```go
func visualizeMetrics(metrics *RedditMetrics, quit <-chan bool)
```

### How it works:
- Updates every 500ms
- Uses ANSI colors for beautiful visualization
- Calculates real-time throughput and latency
- Shows activity bars for visual performance tracking

### Aha Moment! 🎉
The visualizer demonstrates how a system can be both high-performance AND user-friendly - it processes thousands of events while providing real-time insights!

## Data Flow

1. Generator creates events → sends to channel
2. Writer receives from channel → saves to PostgreSQL
3. Processor reads from PostgreSQL → marks as processed
4. Visualizer reads metrics → updates dashboard

Each component runs independently in its own goroutine, demonstrating the power of Go's concurrency model! 