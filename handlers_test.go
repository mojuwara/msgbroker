package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateTopic(t *testing.T) {
	queueName := "TestQueueName"
	testMsg := Msg{Queue: queueName}

	// Creating a topic the first time adds it to queue
	queue := make(QueueType)
	CreateTopic(&testMsg, queue)
	_, ok := queue[queueName]
	assert.True(t, ok, "Should have inserted", queueName, "into the queue")

	// Attempting to create a queue twice shouldn't be an issue
	CreateTopic(&testMsg, queue)
	_, ok = queue[queueName]
	assert.True(t, ok, "Creating the same queue twice should be idempotent")
}

func TestInsertSub(t *testing.T) {
	var (
		queueName  = "TestQueueName"
		testSender = "Eric"
		testMsg1   = Msg{Sender: testSender, Queue: queueName}
		testMsg2   = Msg{Sender: testSender, Queue: queueName}
		testMsg3   = Msg{Sender: testSender, Queue: queueName}
		queues     = make(QueueType)
	)

	CreateTopic(&testMsg1, queues)

	// Inserting a subscriber once should be fine
	AddSubscriber(&testMsg1, queues)
	subs := queues[queueName]
	_, ok := subs[testSender]
	assert.True(t, ok, "Sender should be in list of subscribers")

	// Inserting the same subscriber twice should be fine
	AddSubscriber(&testMsg1, queues)
	subs = queues[queueName]
	_, ok = subs[testSender]
	assert.True(t, ok, "Someone subscribing to a queue twice should be idempotent")

	// Inserting multiple subscribers
	msgs := []*Msg{&testMsg1, &testMsg2, &testMsg3}
	for _, msg := range msgs {
		AddSubscriber(msg, queues)
		subs = queues[msg.Queue]
		_, ok = subs[msg.Sender]
		assert.True(t, ok, "Inserting multiple users in a queue should be fine")
	}
}

func TestRemoveSub(t *testing.T) {
	var (
		queueName  = "TestQueueName"
		testSender = "Eric"
		testMsg1   = Msg{Sender: testSender, Queue: queueName}
		testMsg2   = Msg{Sender: testSender, Queue: queueName}
		testMsg3   = Msg{Sender: testSender, Queue: queueName}
		queues     = make(QueueType)
	)

	CreateTopic(&testMsg1, queues)

	// Inserting multiple subscribers
	msgs := []*Msg{&testMsg1, &testMsg2, &testMsg3}
	for _, msg := range msgs {
		AddSubscriber(msg, queues)
	}

	// Remove and check the subscribers have been removed
	for _, msg := range msgs {
		RemoveSubscriber(msg, queues)
		subs := queues[msg.Queue]
		_, ok := subs[msg.Sender]
		assert.False(t, ok, "User should've been removed from the list of subscribers")
	}
}

func TestGetRandomSub(t *testing.T) {
	subs := map[string]bool{"Alice": true, "Bob": true, "Carl": true, "Damian": true, "Eric": true}

	// Try getting a random subscriber a few times
	for ndx := 0; ndx < 10; ndx++ {
		randomSub := GetRandomSub(subs)
		_, ok := subs[randomSub]
		assert.True(t, ok, "Random subscriber received should be in list of subscribers")
	}
}

// Can't test receiving a message
func TestPostHandler(t *testing.T) {
	initMsgBroker()

	// Create a queue and some test messages
	var (
		queueName = "TestQueueName"
		testAddr  = "1.2.3.4"

		createMsg = Msg{Act: Create, Queue: queueName}
		subMsg    = Msg{Act: Sub, Sender: testAddr, Queue: queueName}
		pubMsg    = Msg{Act: Pub, Queue: queueName, Body: "Test1"}
		unsubMsg  = Msg{Act: Unsub, Sender: testAddr, Queue: queueName}
	)

	/////////////////////////////////////////////////////////////////////
	// Marshal and send create topic message
	result := postMsg(&createMsg)
	resultBody, _ := io.ReadAll(result.Body)

	assert.Equal(t, 200, result.StatusCode, "StatusCode for create topic message was not 200")
	assert.Equal(t, "Message received", string(resultBody), "Expected body of message to be '200'")
	/////////////////////////////////////////////////////////////////////

	/////////////////////////////////////////////////////////////////////
	// Marshal and send subscribe message
	result = postMsg(&subMsg)
	resultBody, _ = io.ReadAll(result.Body)

	assert.Equal(t, 200, result.StatusCode, "StatusCode for subscribe message was not 200")
	assert.Equal(t, "Message received", string(resultBody), "Expected body of message to be '200'")
	/////////////////////////////////////////////////////////////////////

	/////////////////////////////////////////////////////////////////////
	// Marshal and send publish message - Won't send to anyone since there is no server listening to this queue
	result = postMsg(&pubMsg)
	resultBody, _ = io.ReadAll(result.Body)

	assert.Equal(t, 200, result.StatusCode, "StatusCode for publish message was not 200")
	assert.Equal(t, "Message received", string(resultBody), "Expected body of message to be '200'")
	/////////////////////////////////////////////////////////////////////

	/////////////////////////////////////////////////////////////////////
	// Marshal and send unsubscribe message
	result = postMsg(&unsubMsg)
	resultBody, _ = io.ReadAll(result.Body)

	assert.Equal(t, 200, result.StatusCode, "StatusCode for unsubscribe message was not 200")
	assert.Equal(t, "Message received", string(resultBody), "Expected body of message to be '200'")
	/////////////////////////////////////////////////////////////////////

}

func TestPublishMsg(t *testing.T) {
	initMsgBroker()

	// Create test server that sets 'msgReceived' var to true if request is made
	msgReceived := false
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err1 := io.ReadAll(r.Body)

		var msg Msg
		err2 := json.Unmarshal(body, &msg)

		if err1 == nil && err2 == nil && msg.Body == "Test1" {
			msgReceived = true
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	// Create a queue and some test messages
	var (
		queueName = "TestQueueName"
		createMsg = Msg{Act: Create, Queue: queueName}
		subMsg    = Msg{Act: Sub, Sender: testServer.URL, Queue: queueName}
		pubMsg    = Msg{Act: Pub, Queue: queueName, Body: "Test1"}
	)

	/////////////////////////////////////////////////////////////////////
	// Marshal and send create topic message
	result := postMsg(&createMsg)
	resultBody, _ := io.ReadAll(result.Body)

	assert.Equal(t, 200, result.StatusCode, "StatusCode for create topic message was not 200")
	assert.Equal(t, "Message received", string(resultBody), "Expected body of message to be '200'")
	/////////////////////////////////////////////////////////////////////

	// Have the testServer subscribe to a queue
	postMsg(&subMsg)

	// Have someone publish to the queue, making testServer receive a message
	postMsg(&pubMsg)

	// Sleep for 1 second to allow testServer to receive the message
	time.Sleep(time.Second)
	assert.True(t, msgReceived, "Should have received a message after message was published to queue")
}

func postMsg(msg *Msg) *http.Response {
	body, _ := json.Marshal(msg)
	bodyBuffer := bytes.NewBuffer(body)

	request := httptest.NewRequest("POST", "/msg", bodyBuffer)
	response := httptest.NewRecorder()
	handlePost(response, request)

	// Ensure the subscribe message was successful
	return response.Result()
}

func initMsgBroker() {
	// Initialize the application variables
	KEYRANGE = 100 // 0 - 10
	PARTITIONS = 1 // Number of workers

	CHANNELS = make([]chan Msg, int(PARTITIONS))
	CHANNELS[0] = make(chan Msg)
	go worker(CHANNELS[0])
	RANGES = MakePartitions(KEYRANGE, PARTITIONS)
}
