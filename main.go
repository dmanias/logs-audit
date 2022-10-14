package main

import (
	"context"
	"encoding/json"
	"github.com/dmanias/logs-audit/auth"
	"github.com/dmanias/logs-audit/config"
	_ "github.com/dmanias/logs-audit/docs"
	mongoPack "github.com/dmanias/logs-audit/mongo"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// The Event struct creates the event from the input and add it to DB
type Event struct {
	Timestamp time.Time              `json:"timestamp"`
	Service   string                 `json:"service"`
	EventType string                 `json:"eventType"`
	Data      map[string]interface{} `json:"-"` // Rest of the fields should go here.
	Tags      string                 `json:"tags"`
}

type Creator interface {
	eventToBson() bson.M
	eventToString() (string, error)
}

//@desc method createEventBson() creates a bson.M from an Event
//@parameter {Event} event. An event
func (inputEvent Event) eventToBson() bson.M {
	//TODO add input for tags
	bsonInput := bson.A{}
	for _, value := range inputEvent.Data {
		bsonInput = append(bsonInput, value)
	}

	bsonFromJson := bson.M{
		"timestamp": inputEvent.Timestamp,
		"service":   inputEvent.Service,
		"eventType": inputEvent.EventType,
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
	router.HandleFunc("/events", makeConnectDBHandler(searchDBHandler)).Methods("GET")
	router.HandleFunc("/events", makeConnectDBHandler(storeEventsHandler)).Methods("POST")
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
// @Failure 400, 500 {json} json
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
			"message": "Error while registering user. Please try again.",
		})
		return
	}

	if credentials.Username == "" || credentials.Password == "" {

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(bson.M{
			"message": "Please enter a valid username and password.",
		})
		return

	} else {

		response, err := auth.RegisterUser(credentials.Username, credentials.Password)

		if err != nil {
			log.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(bson.M{
				"message": "Error while registering user. Please try again.",
			})
			return
		} else {
			log.Info(response)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(bson.M{
				"message": "New user is registered.",
			})
			return
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
func storeEventsHandler(w http.ResponseWriter, r *http.Request, db *mongo.Database, ctx context.Context) {
	w.Header().Set("Content-Type", "application/json")
	//Create event from input
	inputEvent, err := createEventFromInput(r)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(bson.M{"message": "An error occurred. Please try again."})
		return
	}
	//Use interface
	create := Creator(inputEvent)
	//Create bson.M from event
	bsonFromEvent := create.eventToBson()
	stringFromEvent, err := create.eventToString()

	if err != nil {
		log.Error("Input to String conversion failed")
	}

	//Add to DB
	eventCollection := db.Collection("events")
	_, err = eventCollection.InsertOne(ctx, bsonFromEvent)

	if err != nil {
		log.Fatal(err.Error())
		writeToFile(stringFromEvent)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(bson.M{
			"message": "Error while inserting event. Event is stored in temporal storage",
		})
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bson.M{
		"message": "Event has been stored.",
	})
	return
}

//@desc createEventFromInput() creates an Event from the input
//@parameter {Request} r. The API input
func createEventFromInput(r *http.Request) (Event, error) {
	//Read from body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err)
		return Event{}, err
	}

	event := Event{}
	if err := json.Unmarshal(body, &event); err != nil {
		log.Error(err)
		return Event{}, err
	}
	if err := json.Unmarshal(body, &event.Data); err != nil {
		log.Error(err)
		return Event{}, err
	}
	//remove the following data for efficiency
	delete(event.Data, "timestamp")
	delete(event.Data, "eventType")
	delete(event.Data, "service")

	return event, nil
}

func createEventString(event Event) (string, error) {
	out, err := json.Marshal(event)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	return string(out), nil
}

//@desc buildBsonObject() creates a bson.M from the API input
//@parameter {Request} r. The API input
func buildBsonObject(r *http.Request) bson.M {

	hasTimestamp := r.URL.Query().Has("timeStamp")
	hasService := r.URL.Query().Has("service")
	hasEventType := r.URL.Query().Has("eventType")
	hasData := r.URL.Query().Has("data")
	hasTags := r.URL.Query().Has("tags")

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

	if hasTags {
		query["tags"] = r.URL.Query().Get("tags")
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
func searchDBHandler(w http.ResponseWriter, r *http.Request, db *mongo.Database, ctx context.Context) {
	//TODO search greater and less than the timestamp given
	//TODO sort results if necessary

	w.Header().Set("Content-Type", "application/json")
	eventsCollection := db.Collection("events")

	//Build filter object
	query := buildBsonObject(r)
	filterCursor, err := eventsCollection.Find(ctx, query)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(bson.M{"message": "An error occurred. Please try again."})
		return
	}

	var eventsFiltered []bson.M
	if err = filterCursor.All(ctx, &eventsFiltered); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(bson.M{"message": "An error occurred. Please try again."})
		return
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
			log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(bson.M{"message": "An error occured. Please try again."})
			return
		} else {
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			enc.Encode(tokenDetails)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(bson.M{"message": "You require a username/password to get a token."})
		return
	}

}

func makeConnectDBHandler(fn func(http.ResponseWriter, *http.Request, *mongo.Database, context.Context)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//Authentication check
		if !checkToken(r) {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(bson.M{"message": "Token is missing or it is not valid."})
			return
		}
		//Connect to DB
		cfg := config.New()
		mongoClient, ctx, cancel, err := mongoPack.Connect(cfg.Database.Connector)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(bson.M{"message": "An error occurred. Please try again."})
			return
		}
		defer mongoPack.Close(mongoClient, ctx, cancel)
		db := mongoClient.Database("db")
		fn(w, r, db, ctx)
	}
}

//@desc Monitoring
//Initialization, handling and Prometheus structs
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

//@desc response writer for prometheus
func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

//@desc write header for prometheus
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

//registered variable
var totalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Number of get requests.",
	},
	[]string{"path"},
)

//registered variable
var responseStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "response_status",
		Help: "Status of HTTP response",
	},
	[]string{"status"},
)

//registered variable
var httpDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "http_response_time_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path"})

//@desc Prometheus handler
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

// Prometheus Initialization (Below the metrics are shown in the metrics page)
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

//TODO methods if necessary
//TODO concurrency thread safe
//TODO SOS sort the service or db

//TODO validation //https://medium.com/@apzuk3/input-validation-in-golang-bc24cdec1835#id_token=eyJhbGciOiJSUzI1NiIsImtpZCI6Ijk1NTEwNGEzN2ZhOTAzZWQ4MGM1NzE0NWVjOWU4M2VkYjI5YjBjNDUiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJuYmYiOjE2NjU2NzUwMjUsImF1ZCI6IjIxNjI5NjAzNTgzNC1rMWs2cWUwNjBzMnRwMmEyamFtNGxqZGNtczAwc3R0Zy5hcHBzLmdvb2dsZXVzZXJjb250ZW50LmNvbSIsInN1YiI6IjExNDAxOTQ2MTU1OTY0OTk4ODE3MyIsImVtYWlsIjoiZGltb3N0aGVuaXMubWFuaWFzQGdtYWlsLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJhenAiOiIyMTYyOTYwMzU4MzQtazFrNnFlMDYwczJ0cDJhMmphbTRsamRjbXMwMHN0dGcuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJuYW1lIjoiRGltb3N0aGVuaXMgTWFuaWFzIiwicGljdHVyZSI6Imh0dHBzOi8vbGgzLmdvb2dsZXVzZXJjb250ZW50LmNvbS9hL0FMbTV3dTBIVXdfX1YwSVJRMDF1dmhWQUFJRjNGNUswcE9RcFl1a0YtdTlDQUE9czk2LWMiLCJnaXZlbl9uYW1lIjoiRGltb3N0aGVuaXMiLCJmYW1pbHlfbmFtZSI6Ik1hbmlhcyIsImlhdCI6MTY2NTY3NTMyNSwiZXhwIjoxNjY1Njc4OTI1LCJqdGkiOiIzZmU4NjM3MzI1MTVmNmI3YzI4ZjA1NjI1ZjI4NzUyNDNhNWNiNDMyIn0.Xl-JQDzevM5iJu-tCEhruXIWtS6aR_IPHV2pzsojeXYlbJEvk81AR7Iu8_k88cgBaC4cJ_kyXuF6FfAvyJW6AxsRD_Mmxx-bnt-0PzG8pAMDNH6fwiygps184Qq7Ha3PYkXfbfAlg_cHrmWFrz_9jW3_rkeNPEchAxHV9r1W7GWrBSgM93Sf4UZcWdbEZ-o3UgNw7waD4RMOff4n4rW5pjiF0l7ym7dxS4rmR6lxu44fwkE2xJ6d-tbEA19l9SnjH_tPCPFM435mSRmbZ_0KzEtYJ6xh3uurWti-s7kn_Siq9jfKDgxk02eUwBMIr0v1orhtzXXS4xzpmmXKXk7muA
