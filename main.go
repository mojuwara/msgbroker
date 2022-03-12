package main

import (
	"log"
	"net/http"
)

var logger = log.Default()

// A keyrange and the number of partitions is decided
// The number of partitions is the number of goroutines and channels created
// Each goroutine handles queues that fall into that partition
// Queues "fall into" a partition if the sha256 of the queue name falls in that range
// When the correct range + goroutine is determined, the message is placed into that goroutines channel
// If a message fails to be delivered, it will create a goroutine that retries to deliver the message, with exponential backoff
// Goroutines are unaware of which key range they have, they process whatever msg they receive
func main() {
	// Initialize goroutines and channels
	KEYRANGE = 1000 // 0 - 1000
	PARTITIONS = 10 // Number of workers

	CHANNELS = make([]chan Msg, int(PARTITIONS))
	RANGES = MakePartitions(KEYRANGE, PARTITIONS)

	// Create the channels and pass them to goroutines to begin listening
	for i := 0; i < PARTITIONS; i++ {
		channel := make(chan Msg, MaxChannelLength)
		CHANNELS[i] = channel
		go worker(channel)
	}

	// Start server
	http.HandleFunc("/msg", handlePost)

	logger.Println("Started server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type Status int

// Possible statuses for a message
const (
	Failed     Status = 0
	Success    Status = 1
	Processing Status = 2
	Invalid    Status = 3
)

// Valid actions for an incoming message
const (
	Sub    = "Sub"
	Pub    = "Pub"
	Unsub  = "Unsub"
	Create = "Create"
)

type Msg struct {
	Act    string `json:"act"` // Sub, Unsub, Publish
	Body   string `json:"body"`
	Queue  string `json:"queue"`  // Queue to deliver
	Sender string `json:"sender"` // Who sent the message

	// Internal fields
	id      int    // Used to keep track of messages
	status  Status // Used internally(0 is failed, 1 is delivered, 2 is processing)
	retries int    // Number of times we retried this message
}
