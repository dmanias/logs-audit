package main

import (
	"encoding/json"
	"fmt"
	"github.com/dmanias/logs-audit/config"
	"github.com/dmanias/logs-audit/mongo"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
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

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/events", queryDBHandler).Methods("GET")
	router.HandleFunc("/events", storeEventsHandler).Methods("POST")
	router.HandleFunc("/auth", authenticationHandler).Methods("GET")
	router.HandleFunc("/auth", registrationsHandler).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func registrationsHandler(w http.ResponseWriter, r *http.Request) {

	var credentials Credentials
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if credentials.Username == "" || credentials.Password == "" {

		fmt.Fprintf(w, "Please enter a valid username and password.\r\n")

	} else {

		response, err := registerUser(credentials.Username, credentials.Password)

		if err != nil {
			fmt.Fprintf(w, err.Error())
		} else {
			fmt.Fprintf(w, response)
		}
	}
}

func writeToFile(json bson.M) {
	//file, err := os.Open("mongo/temp.json")
	file, err := os.Create("mongo/temp.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var jsonStr []byte
	jsonStr, err = bson.Marshal(json)
	if err != nil {
		log.Fatal(err)
	}
	file.WriteString(string(jsonStr))
}

func checkToken(r *http.Request) bool {
	authToken := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]
	validToken, err := validateToken(authToken)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return validToken
}

func storeEventsHandler(w http.ResponseWriter, r *http.Request) {

	if !checkToken(r) {
		panic("Token is not valid")
	}

	cfg := config.New()
	mongoClient, ctx, cancel, err := mongo.Connect(cfg.Database.Connector)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	inputEvent := createEventFromInput(r)

	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	defer mongo.Close(mongoClient, ctx, cancel)

	db := mongoClient.Database("db")
	eventCollection := db.Collection("events")

	bsonFromEvent := createEventBson(inputEvent)

	_, err = eventCollection.InsertOne(ctx, bsonFromEvent)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		log.Fatal(err)
		writeToFile(bsonFromEvent)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(bson.M{
			"message": "Error while inserting event. Event is stored in temporal storage",
		})
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bson.M{
		"message": "Event has been stored.",
	})
}

func createEventFromInput(r *http.Request) Event {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	event := Event{}
	if err := json.Unmarshal(body, &event); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &event.Data); err != nil {
		panic(err)
	}
	delete(event.Data, "timestamp")
	delete(event.Data, "eventType")
	delete(event.Data, "service")

	return event
}

func createEventBson(inputEvent Event) bson.M {

	bsonFromJson := bson.M{
		"timestamp": inputEvent.Timestamp,
		"service":   inputEvent.Service,
		"eventType": inputEvent.EventType,
		"data":      inputEvent.Data,
		"tags":      bson.A{"coding", "test"},
	}
	return bsonFromJson
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

func queryDBHandler(w http.ResponseWriter, r *http.Request) {

	if !checkToken(r) {
		panic("Token is not valid")
	}

	cfg := config.New()
	mongoClient, ctx, cancel, err := mongo.Connect(cfg.Database.Connector)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	defer mongo.Close(mongoClient, ctx, cancel)

	db := mongoClient.Database("db")
	eventsCollection := db.Collection("events")

	query := buildBsonObject(r)

	filterCursor, err := eventsCollection.Find(ctx, query)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(bson.M{"message": "Something went wrong"})
	}

	var eventsFiltered []bson.M
	if err = filterCursor.All(ctx, &eventsFiltered); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(bson.M{"message": "Something went wrong"})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bson.M{"events": eventsFiltered})
}

func authenticationHandler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()

	if ok {
		tokenDetails, err := generateToken(username, password)

		if err != nil {
			fmt.Fprintf(w, err.Error())
		} else {
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			enc.Encode(tokenDetails)
		}
	} else {
		fmt.Fprintf(w, "You require a username/password to get a token.\r\n")
	}

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
//TODO get with {id}

//TODO sos search mongo from data and metadata
//TODO SOS mongo secondary keys etc
//TODO SOS sort the service or db
//TODO SOS refactor (functions packages) and tests
//TODO SOS add tags to the query
//TODO SOS fix timestamp query
//TODO SOS logs and if
//TODO SOS users index username
