# Log Audit

This is a log audit service. Events are aggregated and the user can run queries on them. The events are indexed by their invariant parts. The variant parts are stored all together under the name data. The events endpoints are protected with bearer token authentication.

The service uses the following technologies:
- Go
- MongoDB
- Mongo Express
- Docker
- Docker Compose
- Swagger
- Prometheus

### To run:
```shell
docker-compose up --detach --build
```

### API
Port: 8080  
Base path: /api/v1  

**POST** /auth - register a user  
```json
{
  "username": "user",
  "password": "password"
}
```

**GET** /auth - returns a token  
Basic auth with the user credentials  

**POST** /events - add an event  
```json
{
  "timestamp": "2017-11-22T08:44:22.309Z",
  "service": "ADMINISTRATION",
  "eventType": "event",
  "eventNumber": 1,
  "status": 0,
  "action": "creation",
  "customerName": "Babis",
  "customerID": "12354",
  "customerRights": "USER"
}
```
Invariant parts: timestamp, service, eventType
The rest are variant parts that are stored under the name data.
All invariant parts are indexed as strings and variant parts (data) are indexed as text.
With every event are stored some tags to make the search easier and the service more user-friendly.

**GET** /events - returns the events resulted from the query  
Parameters: timestamp, eventType, service, data, tags
```azure
http://localhost:8080/api/v1/events?timestamp=2017-11-22&eventType=event&service=ADMINISTRATION&data=Babis&tags=test
```

### DB Initialization
The database is initialized from the script mongo/init-mongo.js file. The script creates the DB, the collections and the indexes.

### Testing
In the test/test_endpoints.sh file are the curl calls for the API. 
For the GET /events call are presented some benchmark results as well. The search operation the results are exported in the benchmarks.txt file and the metrics are presented in the console.  
The Unit tests are in the main and auth folders and they run with the docker compose up command.

### Documentation
The API is documented with the Swagger UI.
```azure
http://localhost:8080/api/v1/swagger/index.html
```
To compile the swagger documentation run the following command:
```shell
swag init
```
Web-based MongoDB admin interface
```azure
http://localhost:8081
```

### Example of Language specifics 

### Monitoring
Monitoring is done with Prometheus. The metrics are exposed on the /metrics endpoint.
```azure
http://localhost:8080/api/v1/metrics
```

