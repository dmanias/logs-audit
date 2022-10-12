package main

import (
	"fmt"
	"github.com/dmanias/logs-audit/auth"
	"github.com/dmanias/logs-audit/config"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
This is a logs audit API. Events from logs are aggregated and the user can run queries on them.
Events are indexed by their invariant parts and the variant parts are stored all together under the name data.
The endpoints are protected with bearer token authentication.
*/

import (
	"encoding/json"
	_ "github.com/dmanias/logs-audit/docs"
	"github.com/dmanias/logs-audit/mongo"
)

// The Event struct creates the event from the input and add it to DB
type Event struct {
	Timestamp time.Time              `json:"timestamp"`
	Service   string                 `json:"service"`
	EventType string                 `json:"eventType"`
	Data      map[string]interface{} `json:"-"` // Rest of the fields should go here.
}

// The Credentials struct handles and stores the user credentials to the DB
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// @title Logs Audit API documentation
// @version 1.0.0
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.basic BasicAuth

//@desc main() exposes 4 endpoints for user registration, user authentication, logs aggregation and queries
func main() {
	muxRouter := mux.NewRouter()
	router := muxRouter.PathPrefix("/api/v1").Subrouter() //Create base path for all routes
	router.Use(prometheusMiddleware)
	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/events", searchDBHandler).Methods("GET")
	router.HandleFunc("/events", storeEventsHandler).Methods("POST")
	router.HandleFunc("/auth", authenticationHandler).Methods("GET")
	router.HandleFunc("/auth", registrationsHandler).Methods("POST")
	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	log.Fatal(http.ListenAndServe(":8080", router))
}

// registrationsHandler ... Register a user
// @Summary Add a new user to DB
// @Description add new users
// @Tags Auth
// @Accept json
// @Param Input body Credentials false "Body (raw, json)"
// @Success 200 {json} json
// @Failure 500 {json} json
// @Router /auth [post]
func registrationsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var credentials Credentials
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		log.Info(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(bson.M{
			"message": "Error while registering user. Please try again",
		})
	}

	if credentials.Username == "" || credentials.Password == "" {

		fmt.Fprintf(w, "Please enter a valid username and password.\r\n")

	} else {

		response, err := auth.RegisterUser(credentials.Username, credentials.Password)

		if err != nil {
			log.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(bson.M{
				"message": "Error while registering user. Please try again",
			})
		} else {
			log.Info(response)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(bson.M{
				"message": "New user is registered",
			})
		}
	}
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

//@desc checkToken() check if the bearer token is valid
//@parameter {Request} r. The API input
func checkToken(r *http.Request) bool {
	authToken := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]
	validToken, err := auth.ValidateToken(authToken)
	if err != nil {
		log.Error(err.Error())
	}
	return validToken
}

// searchDBHandler ... Search in DB
// @Summary Brings documents according to the criteria
// @Description get documents
// @Tags Events
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Param Input body Event false "Body (raw, json)"
// @Success 201 {json} Event
// @Failure 400, 403, 500 {json} error message
// @Router /events [post]
func storeEventsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !checkToken(r) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(bson.M{"message": "Token is missing or it is not valid"})
	}

	cfg := config.New()
	mongoClient, ctx, cancel, err := mongo.Connect(cfg.Database.Connector)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(bson.M{"message": "An error occurred. Please try again."})
	}

	inputEvent, err := createEventFromInput(r)

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(bson.M{"message": "An error occurred. Please try again."})
	}

	defer mongo.Close(mongoClient, ctx, cancel)

	db := mongoClient.Database("db")
	eventCollection := db.Collection("events")

	bsonFromEvent := createEventBson(inputEvent)
	stringFromEvent, err := createEventString(inputEvent)
	if err != nil {
		log.Error("Input to String conversion failed")
	}

	_, err = eventCollection.InsertOne(ctx, bsonFromEvent)

	if err != nil {
		log.Fatal(err.Error())
		writeToFile(stringFromEvent)
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

//@desc createEventFromInput() creates an Event from the input
//@parameter {Request} r. The API input
func createEventFromInput(r *http.Request) (Event, error) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err.Error())
		return Event{}, err
	}

	event := Event{}
	if err := json.Unmarshal(body, &event); err != nil {
		log.Error(err.Error())
		return Event{}, err
	}
	if err := json.Unmarshal(body, &event.Data); err != nil {
		log.Error(err.Error())
		return Event{}, err
	}
	delete(event.Data, "timestamp")
	delete(event.Data, "eventType")
	delete(event.Data, "service")

	return event, nil
}

//@desc createEventString() creates a string from an Event
//@parameter {Event} event. An event
func createEventString(event Event) (string, error) {
	out, err := json.Marshal(event)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	return string(out), nil
}

//@desc createEventString() creates a string from an Event
//@parameter {Event} event. An event
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

//@desc buildBsonObject() creates a bson.M from the API input
//@parameter {Request} r. The API input
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

// searchDBHandler ... Search in DB
// @Summary Brings documents according to the criteria
// @Description get documents
// @Tags Events
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Param timestamp query string false "timestamp"
// @Param service query string false "the name of the service that sends the event"
// @Param eventType query string false "the type of the event"
// @Param data query string false "extra data to search in the event body"
// @Param tags query string false "metadata given from the service when stores the events"
// @Success 200 {json} Event
// @Failure 400, 500 {json} error message
// @Router /events [get]
func searchDBHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if !checkToken(r) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(bson.M{"message": "Token is missing or it is not valid"})
	}

	cfg := config.New()
	mongoClient, ctx, cancel, err := mongo.Connect(cfg.Database.Connector)
	if err != nil {
		log.Fatal(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(bson.M{"message": "An error occurred. Please try again."})
	}

	defer mongo.Close(mongoClient, ctx, cancel)

	db := mongoClient.Database("db")
	eventsCollection := db.Collection("events")
	query := buildBsonObject(r)

	filterCursor, err := eventsCollection.Find(ctx, query)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(bson.M{"message": "An error occurred. Please try again."})
	}

	var eventsFiltered []bson.M
	if err = filterCursor.All(ctx, &eventsFiltered); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(bson.M{"message": "An error occurred. Please try again."})
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bson.M{"events": eventsFiltered})
}

// searchDBHandler ... Search in DB
// @Summary Brings documents according to the criteria
// @Description get documents
// @Tags Auth
// @Security BasicAuth
// @Success 200 {json} Event
// @Failure 400 {json} error message
// @Router /auth [get]
func authenticationHandler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()

	if ok {
		tokenDetails, err := auth.GenerateToken(username, password)

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

//@desc Monitoring
//Initialization, handling and Prometheus structs
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

var totalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Number of get requests.",
	},
	[]string{"path"},
)

var responseStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "response_status",
		Help: "Status of HTTP response",
	},
	[]string{"status"},
)

var httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_response_time_seconds",
	Help: "Duration of HTTP requests.",
}, []string{"path"})

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()

		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		rw := NewResponseWriter(w)
		next.ServeHTTP(rw, r)

		statusCode := rw.statusCode

		responseStatus.WithLabelValues(strconv.Itoa(statusCode)).Inc()
		totalRequests.WithLabelValues(path).Inc()

		timer.ObserveDuration()
	})
}

func init() {
	err := prometheus.Register(totalRequests)
	if err != nil {
		panic(err)
	}
	err = prometheus.Register(responseStatus)
	if err != nil {
		panic(err)
	}
	err = prometheus.Register(httpDuration)
	if err != nil {
		panic(err)
	}
}

//TODO create function for errors
//TODO template to show events
//TODO probably replace MongoDB with elasticsearch
//TODO add regex for better indexing
//TODO add enviroment for tags/labels
//TODO create admin enviroment
//TODO show results in html
//TODO bearer token to JWT https://blog.logrocket.com/jwt-authentication-go/
//TODO timestamp higher, between etc
//TODO get with {id}
//TODO prometheus
//TODO closures error handling
//TODO methods if necessary
//TODO concurrency thread safe
//TODO coverage tests and benchmarks

//TODO sos search mongo from data and metadata
//TODO SOS mongo secondary keys etc
//TODO SOS sort the service or db
//TODO SOS refactor (functions packages) and tests
//TODO SOS add tags to the query
//TODO SOS fix timestamp query
//TODO SOS logs and if
//TODO SOS users index username
