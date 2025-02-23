package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// ANSI colors and styles
const (
	ColorGreen   = "\033[92m"
	ColorYellow  = "\033[93m"
	ColorRed     = "\033[91m"
	ColorBlue    = "\033[94m"
	ColorMagenta = "\033[95m"
	ColorCyan    = "\033[96m"
	ColorReset   = "\033[0m"
	Bold         = "\033[1m"
)

type RedditMetrics struct {
	activeUsers    int
	eventsHandled  int
	dbOperations   struct {
		writes   int
		reads    int
		updates  int
	}
	startTime      time.Time
	processingTime time.Duration
	mutex          sync.Mutex
}

func initDB() (*sql.DB, error) {
	connStr := "postgres://soulbliss:localpass@localhost:5432/webtraffic_db?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		DROP TABLE IF EXISTS events;
		CREATE TABLE events (
			id SERIAL PRIMARY KEY,
			type VARCHAR(20),
			data JSONB,
			processed BOOLEAN DEFAULT false,
			created_at TIMESTAMP DEFAULT NOW()
		);
		CREATE INDEX idx_events_processed ON events(processed) WHERE NOT processed;
	`)
	return db, err
}

// Simulates user activity - runs in its own goroutine
func generateEvents(eventChan chan<- map[string]interface{}, metrics *RedditMetrics, quit <-chan bool) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-quit:
			return
		case <-ticker.C:
			event := map[string]interface{}{
				"type": []string{"post", "comment", "upvote", "downvote"}[rand.Intn(4)],
				"user": fmt.Sprintf("user_%d", rand.Intn(1000)),
				"data": fmt.Sprintf("content_%d", rand.Intn(1000)),
				"timestamp": time.Now(),
			}
			eventChan <- event

			metrics.mutex.Lock()
			metrics.eventsHandled++
			metrics.mutex.Unlock()
		}
	}
}

// Stores events in database - runs in its own goroutine
func storeEvents(db *sql.DB, eventChan <-chan map[string]interface{}, metrics *RedditMetrics, quit <-chan bool) {
	for {
		select {
		case <-quit:
			return
		case event := <-eventChan:
			start := time.Now()
			
			// Convert map to JSON string for PostgreSQL JSONB
			jsonData, err := json.Marshal(event)
			if err != nil {
				fmt.Printf("Error marshaling event: %v\n", err)
				continue
			}

			_, err = db.Exec(`
				INSERT INTO events (type, data)
				VALUES ($1, $2)
			`, event["type"], jsonData)

			if err != nil {
				fmt.Printf("Error storing event: %v\n", err)
				continue
			}

			metrics.mutex.Lock()
			metrics.dbOperations.writes++
			metrics.processingTime += time.Since(start)
			metrics.mutex.Unlock()
		}
	}
}

// Processes events - runs in its own goroutine
func processEvents(db *sql.DB, metrics *RedditMetrics, quit <-chan bool) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-quit:
			return
		case <-ticker.C:
			// First read unprocessed events
			rows, err := db.Query(`
				SELECT id FROM events 
				WHERE processed = false 
				ORDER BY created_at 
				LIMIT 10
				FOR UPDATE SKIP LOCKED
			`)
			if err != nil {
				fmt.Printf("Error reading events: %v\n", err)
				continue
			}

			metrics.mutex.Lock()
			metrics.dbOperations.reads++
			metrics.mutex.Unlock()

			// Collect IDs to update
			var ids []int
			for rows.Next() {
				var id int
				if err := rows.Scan(&id); err != nil {
					fmt.Printf("Error scanning row: %v\n", err)
					continue
				}
				ids = append(ids, id)
			}
			rows.Close()

			if len(ids) > 0 {
				// Update events in batch
				query := fmt.Sprintf(`
					UPDATE events 
					SET processed = true 
					WHERE id = ANY($1)
				`)
				_, err = db.Exec(query, pq.Array(ids))
				if err != nil {
					fmt.Printf("Error updating events: %v\n", err)
					continue
				}

				metrics.mutex.Lock()
				metrics.dbOperations.updates++
				metrics.mutex.Unlock()
			}
		}
	}
}

func visualizeMetrics(metrics *RedditMetrics, quit <-chan bool) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-quit:
			return
		case <-ticker.C:
			metrics.mutex.Lock()
			runningTime := time.Since(metrics.startTime).Seconds()
			eventsPerSec := 0.0
			writesPerSec := 0.0
			readsPerSec := 0.0
			updatesPerSec := 0.0
			totalOps := metrics.dbOperations.writes + metrics.dbOperations.reads + metrics.dbOperations.updates
			avgProcessingTime := int64(0)

			if runningTime > 0 {
				eventsPerSec = float64(metrics.eventsHandled) / runningTime
				writesPerSec = float64(metrics.dbOperations.writes) / runningTime
				readsPerSec = float64(metrics.dbOperations.reads) / runningTime
				updatesPerSec = float64(metrics.dbOperations.updates) / runningTime
			}
			
			if totalOps > 0 {
				avgProcessingTime = metrics.processingTime.Milliseconds() / int64(totalOps)
			}
			metrics.mutex.Unlock()

			// Clear screen
			fmt.Print("\033[H\033[2J")
			
			// Header
			fmt.Printf("%s%s🚀 Go Concurrency Demo - Real-time Event Processing%s\n", Bold, ColorCyan, ColorReset)
			fmt.Println(strings.Repeat("=", 70))

			// System Status
			fmt.Printf("\n%s💻 System Status:%s\n", Bold, ColorReset)
			fmt.Printf("• Event Generator    : %sGenerating %d events/second%s\n", 
				ColorGreen, int(eventsPerSec), ColorReset)
			fmt.Printf("• Database Writer    : %sWriting %d records/second%s\n", 
				ColorBlue, int(writesPerSec), ColorReset)
			fmt.Printf("• Event Processor    : %sProcessing %d records/second%s\n", 
				ColorMagenta, int(updatesPerSec), ColorReset)

			// Real-time Performance
			fmt.Printf("\n%s📊 Real-time Performance:%s\n", Bold, ColorReset)
			showActivityBar("Writes/sec", writesPerSec, 50, ColorBlue, "records")
			showActivityBar("Reads/sec", readsPerSec, 50, ColorGreen, "records")
			showActivityBar("Updates/sec", updatesPerSec, 50, ColorMagenta, "records")

			// Overall Statistics
			fmt.Printf("\n%s📈 Overall Statistics:%s\n", Bold, ColorReset)
			fmt.Printf("Total Events      : %s%d events generated%s\n", ColorGreen, metrics.eventsHandled, ColorReset)
			fmt.Printf("Database Writes   : %s%d records written%s\n", ColorBlue, metrics.dbOperations.writes, ColorReset)
			fmt.Printf("Database Reads    : %s%d records read%s\n", ColorGreen, metrics.dbOperations.reads, ColorReset)
			fmt.Printf("Records Processed : %s%d records updated%s\n", ColorMagenta, metrics.dbOperations.updates, ColorReset)
			fmt.Printf("Average Latency   : %s%d milliseconds%s per operation\n", ColorYellow, avgProcessingTime, ColorReset)
			fmt.Printf("Uptime           : %s%.1f seconds%s\n", ColorCyan, runningTime, ColorReset)

			// Explanation
			fmt.Printf("\n%s💡 How It Works:%s\n", Bold, ColorReset)
			fmt.Println("1. Generator creates new events every 100 milliseconds")
			fmt.Println("2. Writer instantly saves each event to PostgreSQL")
			fmt.Println("3. Processor handles events in batches every 200 milliseconds")
			fmt.Printf("%sAll operations run simultaneously with zero blocking!%s\n", ColorYellow, ColorReset)
		}
	}
}

func showActivityBar(label string, value float64, max float64, color string, unit string) {
	width := 40
	filled := int((value / max) * float64(width))
	if filled > width {
		filled = width
	}

	fmt.Printf("%-14s [%s%s%s%s] %d %s/second\n",
		label,
		color,
		strings.Repeat("█", filled),
		strings.Repeat(" ", width-filled),
		ColorReset,
		int(value),
		unit,
	)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Step 1: Initialize
	fmt.Println("🚀 Starting Go Concurrency Demo")
	fmt.Println("Watch how Go handles multiple operations in parallel...")
	time.Sleep(2 * time.Second)

	// Step 2: Setup Database
	fmt.Println("\n1️⃣  Connecting to PostgreSQL...")
	db, err := initDB()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer db.Close()
	time.Sleep(1 * time.Second)

	// Step 3: Initialize channels and metrics
	fmt.Println("2️⃣  Initializing communication channels...")
	eventChan := make(chan map[string]interface{}, 100)
	quit := make(chan bool)
	metrics := &RedditMetrics{startTime: time.Now()}
	time.Sleep(1 * time.Second)

	// Step 4: Launch goroutines
	fmt.Println("3️⃣  Launching goroutines...")
	fmt.Println("     • Event Generator")
	go generateEvents(eventChan, metrics, quit)
	time.Sleep(500 * time.Millisecond)

	fmt.Println("     • Database Writer")
	go storeEvents(db, eventChan, metrics, quit)
	time.Sleep(500 * time.Millisecond)

	fmt.Println("     • Event Processor")
	go processEvents(db, metrics, quit)
	time.Sleep(500 * time.Millisecond)

	fmt.Println("     • Metrics Visualizer")
	go visualizeMetrics(metrics, quit)
	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n4️⃣  All systems running! Watch the magic happen...")
	time.Sleep(1 * time.Second)

	// Run for 30 seconds
	time.Sleep(60 * time.Second)

	// Cleanup
	close(quit)
	fmt.Println("\n✨ Demo complete! This showed how Go makes concurrent programming simple and efficient.")
}
