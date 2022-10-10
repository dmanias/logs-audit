package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dmanias/logs-audit/config"
	"github.com/dmanias/logs-audit/mongo"
	"github.com/gorilla/mux"
	"github.com/hellofresh/health-go/v4"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"os"
	"strings"
	"time"
)

type Event struct {
	Timestamp time.Time              `json:"timestamp"`
	Service   string                 `json:"service"`
	EventType string                 `json:"eventType"`
	Data      map[string]interface{} `json:"-"` // Rest of the fields should go here.
}

type Query struct {
	Timestamp string `bson:"timestamp"`
	Service   string `bson:"service"`
	EventType string `bson:"eventType"`
	Data      string `bson:"-"` // Rest of the fields should go here.
}

func main() {

	jsons, err := loadJsons()
	if err != nil {
		log.Fatal(err)
	}
	addToDB(jsons)
	router := mux.NewRouter()
	router.HandleFunc("/events", queryDB).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}
func loadJsons() ([]string, error) {

	filename := "jsons.txt"
	content, err := os.ReadFile(filename)

	if err != nil {
		log.Fatal(err)
	}

	fileBody := string(content)
	split := strings.Split(fileBody, "\n\n")
	return split, nil
}

func addToDB(jsons []string) {
	cfg := config.New()
	mongoClient, ctx, cancel, err := mongo.Connect(cfg.Database.Connector)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	defer mongo.Close(mongoClient, ctx, cancel)
	h, _ := health.New()
	err = h.Register(health.Config{
		Name:      "mongo-check",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: func(ctx context.Context) error {
			mongo.Ping(mongoClient, ctx)
			return nil
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	db := mongoClient.Database("db")
	event := db.Collection("event")

	for _, jsonStr := range jsons {
		json := jsonStruct(jsonStr)
		_, err := event.InsertOne(ctx, bson.D{
			{"timestamp", json.Timestamp},
			{"eventType", json.EventType},
			{"data", json.Data},
			{"service", json.Service},
			{"tags", bson.A{"coding", "test"}},
		})

		if err != nil {
			log.Fatal(err)
		}
	}
}

func jsonStruct(jsonStr string) Event {
	event := Event{}
	if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
		panic(err)
	}
	if err := json.Unmarshal([]byte(jsonStr), &event.Data); err != nil {
		panic(err)
	}
	delete(event.Data, "timestamp")
	delete(event.Data, "eventType")
	delete(event.Data, "service")

	return event
}

func buildBsonObject(r *http.Request) bson.M {

	hasTimestamp := r.URL.Query().Has("timeStamp")
	hasService := r.URL.Query().Has("service")
	hasEventType := r.URL.Query().Has("eventType")
	hasData := r.URL.Query().Has("data")

	query := bson.M{}
	if hasTimestamp {
		query["timestamp"] = r.URL.Query().Get("timeStamp")
	}
	if hasEventType {
		query["eventType"] = r.URL.Query().Get("eventType")
	}
	if hasData {
		query["data"] = r.URL.Query().Get("data")
	}

	if hasService {
		query["service"] = r.URL.Query().Get("service")
	}
	return query
}

func queryDB(w http.ResponseWriter, r *http.Request) {
	// https://www.mongodb.com/blog/post/quick-start-golang--mongodb--how-to-read-documents

	cfg := config.New()
	mongoClient, ctx, cancel, err := mongo.Connect(cfg.Database.Connector)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	defer mongo.Close(mongoClient, ctx, cancel)

	db := mongoClient.Database("db")
	eventsCollection := db.Collection("event")

	query := buildBsonObject(r)
	fmt.Println(query)

	filterCursor, err := eventsCollection.Find(ctx, query)
	if err != nil {
		log.Fatal(err)
	}

	var eventsFiltered []bson.M
	if err = filterCursor.All(ctx, &eventsFiltered); err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eventsFiltered)
}

//TODO JWT https://blog.logrocket.com/jwt-authentication-go/
//TODO refactor connect to DB. add health check to the first function
//TODO create function for errors
//TODO template to show events
//TODO probably replace MongoDB with elasticsearch
//TODO add regex for better indexing
//TODO add enviroment for tags/labels
//TODO create admin enviroment
//TODO show results in html
//TODO bearer token to JWT
//TODO different http for GET, POST
//TODO timestamp higher, between etc

//TODO sos search mongo from data and metadata
//TODO SOS mongo secondary keys etc
//TODO SOS sort the service or db
//TODO SOS refactor (functions packages) and tests
//TODO SOS add tags to the query
//TODO SOS logs and if
