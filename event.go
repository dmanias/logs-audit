package main

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"os"
	"time"
)

// The Event struct creates the event from the input and add it to DB
type Event struct {
	Timestamp time.Time              `json:"timestamp"`
	Service   string                 `json:"service"`
	EventType string                 `json:"eventType"`
	Data      map[string]interface{} `json:"data"` // Rest of the fields should go here.
	Tags      string                 `json:"tags"`
}

//@desc method createEventBson() creates a bson.M from an Event
//@parameter {Event} event. An event
func (event Event) eventToBson() bson.M {
	//TODO add input for tags
	bsonInput := bson.A{}
	for _, value := range event.Data {
		bsonInput = append(bsonInput, value)
	}

	bsonFromJson := bson.M{
		"timestamp": event.Timestamp,
		"service":   event.Service,
		"eventType": event.EventType,
		"data":      bsonInput,
		"tags":      bson.A{"coding", "test"},
	}
	return bsonFromJson
}

//@desc method createEventString() creates a string from an Event
//@parameter {Event} event. An event
func (event Event) eventToString() (string, error) {
	out, err := json.Marshal(event)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	return string(out), nil
}

func (event Event) store(client *mongo.Client, ctx context.Context) error {
	//Create bson.M from event
	bsonFromEvent := event.eventToBson()
	stringFromEvent, err := event.eventToString()

	if err != nil {
		log.Error("Input to String conversion failed")
		return err
	}

	//Add to DB
	db := client.Database("db")
	eventCollection := db.Collection("events")
	_, err = eventCollection.InsertOne(ctx, bsonFromEvent) //TODO change the Blank identifier
	if err != nil {
		writeToFile(stringFromEvent) //write to temp file if mongo is down
		return err
	}
	return nil
}

// @desc search in DB for the events
func search(client *mongo.Client, ctx context.Context, query bson.M) ([]bson.M, error) {
	eventsCollection := client.Database("db").Collection("events")
	var eventsFiltered []bson.M
	//Build filter object
	filterCursor, err := eventsCollection.Find(ctx, query)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if err = filterCursor.All(ctx, &eventsFiltered); err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return eventsFiltered, nil
}

//@desc writeToFile() writes the input to a temporary storage ("mongo/temp.json") when the DB is down
//@parameter {string} jsonInput. The input in json string
func writeToFile(jsonInput string) {

	f, err := os.OpenFile("mongo/temp.json", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Error(err.Error())
	}

	defer f.Close()

	if _, err = f.WriteString(jsonInput); err != nil {
		log.Error(err.Error())
	}
}
