package main

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// TODO: Remove overhead of creating a new connection each time
// TODO: Add authserver
func connectToMongo(ctx context.Context) (*mongo.Client, error) {
	host := os.Getenv("DBHOST")
	port := os.Getenv("DBPORT")
	user := os.Getenv("DBUSER")
	pwrd := os.Getenv("DBPWRD")
	dbURI := fmt.Sprintf("mongodb://%s:%s@%s:%s", user, pwrd, host, port)

	// Create a new client and connect to the server
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURI))
	if err != nil {
		return client, err
	}

	// Ping the primary
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return client, err
	}

	fmt.Println("Successfully connected to MongoDB.")
	return client, nil
}

// Find a way to retry messages if it fails to persist
func persistWorker() {
	msgs := []Msg{}
	for msg := range PersistChannel {
		msgs = append(msgs, msg)

		// Clear and then reset the queue once saved
		if len(msgs) == PersistThreshold {
			persistInMongo(msgs)
			msgs = []Msg{}
		}
	}
}

// Create connection and persist
func persistInMongo(msgs []Msg) bool {
	db := os.Getenv("DBNAME")
	collection := os.Getenv("DBCOLL")
	ctx := context.TODO()
	client, err := connectToMongo(ctx)
	if err != nil {
		logger.Println("Could not establish connection to MongoDB:", err)
		return false
	}

	// Close connection after we're done
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			logger.Println("Error disconnecting from DB:", err)
		}
	}()

	// Inserts the docs into MongoDB
	coll := client.Database(db).Collection(collection)
	_, err = coll.InsertMany(ctx, marshalToBson(msgs))
	if err == nil {
		logger.Println("Wrote", len(msgs), "records into MongoDB")
		return true
	} else {
		logger.Println("Unable to persist messages to MongoDB")
		return false
	}
}

func marshalToBson(msgs []Msg) []interface{} {
	// Marshal the message
	docs := make([]interface{}, 0)
	for _, msg := range msgs {
		bsonVal := bson.D{
			{Key: "_id", Value: msg.id},
			{Key: "Act", Value: msg.Act},
			{Key: "Time", Value: msg.time},
			{Key: "Body", Value: msg.Body},
			{Key: "Queue", Value: msg.Queue},
			{Key: "Sender", Value: msg.Sender},
		}
		docs = append(docs, bsonVal)
	}
	return docs
}
