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

// @title Logs Audit API documentation
// @version 1.0.0
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.basic BasicAuth

//@desc main() exposes 4 endpoints for user registration, user authentication, logs aggregation and queries
func main() {
	a := App{}
	a.initialize()
	a.Run(":8080")
	defer mongoPack.Close(a.DB, a.Context.ctx, a.Context.cancel)
}

type App struct {
	Router  *mux.Router
	DB      *mongo.Client
	Context Context
}

type Context struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (a *App) initialize() {
	//Connect to DB
	cfg := config.New()
	mongoClient, ctx, cancel, err := mongoPack.Connect(cfg.Database.Connector)
	if err != nil {
		log.Error(err)
	}
	a.DB = mongoClient
	a.Context.ctx = ctx
	a.Context.cancel = cancel
	a.Router = mux.NewRouter().PathPrefix("/api/v1").Subrouter() //New router with base path for all routes
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	a.Router.Use(prometheusMiddleware)
	a.Router.Handle("/metrics", promhttp.Handler())
	a.Router.HandleFunc("/events", a.searchDBHandler).Methods("GET")
	a.Router.HandleFunc("/events", a.storeEventsHandler).Methods("POST")
	a.Router.HandleFunc("/auth", a.authenticationHandler).Methods("GET")
	a.Router.HandleFunc("/auth", a.registrationsHandler).Methods("POST")
	a.Router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

// searchDBHandler ... Search in DB
// @Summary Brings documents according to the criteria
// @Description get documents
// @Tags Events
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Param Input body Event false "Body (raw, json)"
// @Success 201 {json} Event
// @Failure 400 {json} error message
// @Failure 403 {json} error message
// @Failure 500 {json} error message
// @Router /events [post]
func (a *App) storeEventsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//Authentication check
	if !checkToken(r) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(bson.M{"message": "Token is missing or it is not valid."})
		return
	}
	//Create event from input
	//Read from body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(bson.M{
			"message": "Error while inserting event. Event is stored in temporal storage",
		})
		return
	}

	event := Event{}
	if err := json.Unmarshal(body, &event); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(bson.M{
			"message": "Error while inserting event. Event is stored in temporal storage",
		})
		return
	}
	if err := json.Unmarshal(body, &event.Data); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(bson.M{
			"message": "Error while inserting event. Event is stored in temporal storage",
		})
		return
	}
	//remove the following keys from data map
	delete(event.Data, "timestamp")
	delete(event.Data, "eventType")
	delete(event.Data, "service")

	event.store(a.DB, a.Context.ctx)

	if err != nil {
		log.Error(err.Error())
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

// registrationsHandler ... Register a user
// @Summary Add a new user to DB
// @Description add new users
// @Tags Auth
// @Accept json
// @Param Input body Credentials false "User credentials"
// @Success 200 {json} json
// @Failure 400 {json} json
// @Failure 500 {json} json
// @Router /auth [post]
func (a *App) registrationsHandler(w http.ResponseWriter, r *http.Request) {
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

	}

	response, err := auth.RegisterUser(credentials.Username, credentials.Password)

	if err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(bson.M{
			"message": "Error while registering user. Please try again.",
		})
		return
	}
	log.Info(response)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bson.M{
		"message": "New user is registered.",
	})
	return
}

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

// The Credentials struct handles and stores the user credentials to the DB
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// authenticationHandler ... Brings a token
// @Summary Brings a new token for the user
// @Description Brings a new token
// @Tags Auth
// @Security BasicAuth
// @Success 200 {json} json
// @Failure 400 {json} json
// @Router /auth [get]
func (a *App) authenticationHandler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	w.Header().Set("Content-Type", "application/json")
	if ok {
		tokenDetails, err := auth.GenerateToken(username, password)

		if err != nil {
			log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(bson.M{"message": "An error occured. Please try again."})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenDetails)
		return

	}
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(bson.M{"message": "You require a username/password to get a token."})
	return

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
func (a *App) heckToken(r *http.Request) bool {
	authToken := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]
	validToken, err := auth.ValidateToken(a.DB, a.Context.ctx, authToken)
	if err != nil {
		log.Error(err.Error())
	}
	return validToken
}

type storeEventsResponse struct {
	Message string `json:"message"`
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
// @Param timestamp query string false "2017-11-22"
// @Param service query string false "the name of the service that sends the event"
// @Param eventType query string false "the type of the event"
// @Param data query string false "extra data to search in the event body"
// @Param tags query string false "metadata given from the service when stores the events"
// @Success 200 {json} Event
// @Failure 400 {json} json
// @Failure 500 {json} json
// @Router /events [get]
func (a *App) searchDBHandler(w http.ResponseWriter, r *http.Request) {
	//TODO search greater and less than the timestamp given
	//TODO sort results if necessary

	w.Header().Set("Content-Type", "application/json")
	//Authentication check
	if !checkToken(r) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(bson.M{"message": "Token is missing or it is not valid."})
		return
	}
	query := buildBsonObject(r)
	eventsFiltered, err := search(a.DB, a.Context.ctx, query)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(bson.M{"message": "An error occurred. Please try again."})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bson.M{"events": eventsFiltered})
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

//Init , initialization
// ta polla orismata sto ctx (closedb)
// sos to event store exei 2 pointers to opoio den tou aresei
// use interface (ta esvisa, na ta afiso?)
//sos ta messages
//sos ta test
