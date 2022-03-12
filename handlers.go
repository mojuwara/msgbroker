package main

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// Each queue will have a map of subscribers
// The key is the subscriber and the value is always boolean true
type SubsType = map[string]bool

// Queue type, each queue topic maps to subscribers
type QueueType = map[string]SubsType

// Size of channel for each worker
const MaxChannelLength = 100

// The maximum value of the key range, [0, KEYRANGE]
var KEYRANGE int

// Stores the key range for each partition
var RANGES []int

// Stores the channels that workers are listening on
var CHANNELS []chan Msg

// The number of partitions(and worker goroutines) created to process messages
var PARTITIONS int

// After a goroutine handles this many messages, will save to DB
const PersistThreshold = 50

// TODO: Send response to sender that this message was received
func handlePost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Read body of message
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Println(err)
		return
	}

	// Unmarshal body of message into Msg object
	var msg Msg
	err = json.Unmarshal(body, &msg)
	if err != nil {
		logger.Println("Error while unmarshalling message", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Could not unmarshal msg" + err.Error()))
		return
	}

	logger.Printf("Message for queue: %s, from: %s, id: %d, act: %s received\n", msg.Queue, msg.Sender, msg.id, msg.Act)
	channel := CHANNELS[FindRange(RANGES, Hash(msg.Queue, KEYRANGE))]
	channel <- msg

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Message received"))
}

// Save to a DB every 100 messages?
func worker(channel chan Msg) {
	// Create the map of queue topics to subscribers
	queue := make(QueueType)

	for msg := range channel {
		processMsg(&msg, queue)
		if msg.status == Failed {
			msg.retries += 1
			go retryFailedMsg(&msg, channel)
		}
	}
}

// Place message back in the queue after sleeping
// Retries in 2 secs, then 4 secs, 8, 16, 32, ...
func retryFailedMsg(msg *Msg, channel chan Msg) {
	// Sleep time doubles each time the messsage fails to deliver
	secondsToSleep := time.Duration(math.Exp2(float64(msg.retries)))
	logger.Printf("Message for queue: %s, from: %s, id: %d, act: %s will be retried in %d seconds", msg.Queue, msg.Sender, msg.id, msg.Act, secondsToSleep)

	time.Sleep(time.Second * secondsToSleep)
	channel <- *msg
}

// Calls the correct function to handle this message
func processMsg(msg *Msg, queue QueueType) {
	switch msg.Act {
	case Sub:
		AddSubscriber(msg, queue)
	case Unsub:
		RemoveSubscriber(msg, queue)
	case Pub:
		publishMsg(msg, queue)
	case Create:
		CreateTopic(msg, queue)
	default:
		logger.Println("Invalid action specified in message")
	}
}

// Create a new queue topic
func CreateTopic(msg *Msg, queue QueueType) {
	_, ok := queue[msg.Queue]
	if !ok {
		queue[msg.Queue] = make(SubsType)
	}
	msg.status = Success
	logger.Println("Created queue topic:", msg.Queue)
}

// Add  subscriber to queue list
func AddSubscriber(msg *Msg, queue QueueType) {
	subs, ok := queue[msg.Queue]
	if !ok {
		logger.Printf("Message for queue: %s, from: %s, id: %d, act: %s is invalid. Cannot subscribe to queue topic that does not exist.\n", msg.Queue, msg.Sender, msg.id, msg.Act)
		msg.status = Invalid
		return
	}

	subs[msg.Sender] = true
}

// Remove subscriber from queue list
func RemoveSubscriber(msg *Msg, queue QueueType) {
	subs, ok := queue[msg.Queue]
	if !ok {
		logger.Printf("Message for queue: %s, from: %s, id: %d, act: %s is invalid. Cannot unsubscribe to queue topic that does not exist.\n", msg.Queue, msg.Sender, msg.id, msg.Act)
		msg.status = Invalid
		return
	}

	delete(subs, msg.Sender)
	msg.status = Success
}

// Publish message to first person we find subscribed to this queue
func publishMsg(msg *Msg, queue QueueType) {
	logger.Println("in publishMsg")
	subs, ok := queue[msg.Queue]
	if !ok {
		logger.Printf("Message for queue: %s, from: %s, id: %d, act: %s is invalid. Cannot publish to queue topic that does not exist.\n", msg.Queue, msg.Sender, msg.id, msg.Act)
		msg.status = Invalid
		return
	}

	// Marshall message to send to recepient
	body, err := json.Marshal(*msg)
	if err != nil {
		// TODO: Tell sender the message failed to be encoded and to check format
		logger.Printf("Message for queue: %s, from: %s, id: %d, act: %s failed to marshal.\n", msg.Queue, msg.Sender, msg.id, msg.Act)
		return
	}

	// Send message to recepient
	subscriber := GetRandomSub(subs)
	bodyBuffer := bytes.NewBuffer(body)
	resp, err := http.Post(subscriber, "application/json", bodyBuffer)
	if err == nil && resp.StatusCode == 200 {
		msg.status = Success
		logger.Printf("Message for queue: %s, from: %s, id: %d, act: %s sent successfully to %s.\n", msg.Queue, msg.Sender, msg.id, msg.Act, subscriber)
	} else {
		msg.status = Failed
		logger.Printf("Message for queue: %s, from: %s, id: %d, act: %s failed to send to %s.\n", msg.Queue, msg.Sender, msg.id, msg.Act, subscriber)
	}
}

// Pick a random person from given map of subscribers
func GetRandomSub(subs SubsType) string {
	length := len(subs)
	if length == 0 {
		return ""
	}

	keys := []string{}
	for k := range subs {
		keys = append(keys, k)
	}
	return keys[rand.Intn(length)]
}
