# Message Broker

## Learn:
- Encryption-at-rest
- Connection pools

## Features:
- Can register multiple queues, by name
- Can subscribe/unsubscribe to different message types(RESTfully)
- Messages remain in the queue until delivered
- Logging format: queue_name, sender/recipient, msg
	- Log when initialized, message sent/received, queue created(and destroyed?),
- Register callbacks(might need some client or glue code)
- Guaranteed message delivery
	- If all consumers offline, will wait until one is up
	- Will resend message if ack not received, possibly to a different broker
- Support greedy delivery pattern & fan-out pattern
- Discovery: Can be registered with a Service Registry

## Interface:
- Host the message broker somewhere(VM, cloud, private network)
- Open connection with the message broker
- Create a queue if it doesn't already exist(idempotent)
- Publish messages into the queue
- Close connection to the message broker
- Messages currently accept string bodies

## Implementation:
- Main thread listens for incoming requests and forwards it to correct goroutine, via channel
	- Odds of creating 2 channels at the same time?
- A single goroutine handles each queue to avoid race conditions or slowing down the program with locking
	- The goroutine maintains list of subscribers
- All messages received are POST requests

## Future:
- Work as a cluster of message brokers
	- Need a message broker between them
- Register callbacks when message is delivered
- Auto acknowledge?
- Metrics
- Store messages in DB
- Returning error messages
- Field to determine if broadcast or send to 1 person
- Give ID to messages
- Check how much memory is used as number of goroutines increases
- Have some restriction on who can send a message to the msgbroker
