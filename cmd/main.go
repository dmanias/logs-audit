package main

import (
	"encoding/json"
	"github.com/dmanias/logs-audit/auth"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// @title Logs Audit API documentation
// @version 1.0.0
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.basic BasicAuth

//@desc main() exposes 4 endpoints for user registration, user authentication, logs aggregation and queries
func main() {
	a := App{}
	a.initializeRoutes()
	a.Run(":8080")
}

type App struct {
	Router *mux.Router
}

func (a *App) initializeRoutes() {
	a.Router = mux.NewRouter().PathPrefix("/api/v1").Subrouter() //New router with base path for all routes
	a.Router.Use(prometheusMiddleware)
	a.Router.Handle("/metrics", promhttp.Handler())
	a.Router.HandleFunc("/events", searchDBHandler).Methods("GET")
	a.Router.HandleFunc("/events", storeEventsHandler).Methods("POST")
	a.Router.HandleFunc("/auth", authenticationHandler).Methods("GET")
	a.Router.HandleFunc("/auth", registrationsHandler).Methods("POST")
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
func storeEventsHandler(w http.ResponseWriter, r *http.Request) {
	//Authentication check
	if _, err := checkToken(r); err != nil {
		errorResponse(w, http.StatusForbidden, err.Error())
		return
	}
	//Create event from input
	//Read from body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error(err.Error())
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	event := Event{}
	if err := json.Unmarshal(body, &event); err != nil {
		log.Error(err.Error())
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := json.Unmarshal(body, &event.Data); err != nil {
		log.Error(err.Error())
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	//remove the following keys from data map
	delete(event.Data, "timestamp")
	delete(event.Data, "eventType")
	delete(event.Data, "service")

	event.store()

	if err != nil {
		log.Error(err.Error())
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	okResponse(w, http.StatusCreated, "Event has been stored.")
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
func registrationsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var credentials Credentials
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		log.Info(err.Error())
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if credentials.Username == "" || credentials.Password == "" {
		errorResponse(w, http.StatusBadRequest, "Please enter a valid username and password.")
		return
	}

	response, err := auth.RegisterUser(credentials.Username, credentials.Password)

	if err != nil {
		log.Error(err.Error())
		errorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Info(response)
	okResponse(w, http.StatusOK, "New user is registered.")
	return
}

//@desc checkToken() check if the bearer token is valid
//@parameter {Request} r. The API input
func checkToken(r *http.Request) (bool, error) {
	authToken := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]
	validToken, err := auth.ValidateToken(authToken)
	if err != nil {
		log.Error(err.Error())
		return false, err
	}
	return validToken, nil
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
func authenticationHandler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if ok {
		tokenDetails, err := auth.GenerateToken(username, password)

		if err != nil {
			log.Error(err.Error())
			errorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tokenDetails)
		return
	}
	errorResponse(w, http.StatusBadRequest, "You require a username/password to get a token.")
	return
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

type SearchHandlerResponse struct {
	Message []bson.M `json:"events"`
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
func searchDBHandler(w http.ResponseWriter, r *http.Request) {
	//TODO search greater and less than the timestamp given
	//TODO sort results if necessary

	w.Header().Set("Content-Type", "application/json")
	//Authentication check
	if _, err := checkToken(r); err != nil {
		errorResponse(w, http.StatusForbidden, err.Error())
		return
	}
	query := buildBsonObject(r)
	eventsFiltered, err := search(query)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SearchHandlerResponse{Message: eventsFiltered})
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type OkResponse struct {
	Message string `json:"message"`
}

func errorResponse(w http.ResponseWriter, code int, message string) {
	response, err := json.Marshal(ErrorResponse{Error: message})
	if err != nil {
		log.Error(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func okResponse(w http.ResponseWriter, code int, message string) {
	response, err := json.Marshal(OkResponse{Message: message})
	if err != nil {
		log.Error(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
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
//ta statuses
//to bson.m sto search response kai to reponse sto token

//pantelis, vasi kai na allsko to config

//aleksandros provlima me to koino db
