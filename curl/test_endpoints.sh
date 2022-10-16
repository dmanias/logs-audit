#!/bin/bash -x
#register user
#res=$(curl -X POST -F 'username=dmanias' -F 'password=1234' http://localhost:8080/api/v1/auth)
#res=$(curl -X POST -F 'username=dmanias' -F 'password=1234' http://localhost:8080/api/v1/auth)

res=$(curl -X POST -H "Content-Type: application/json" -d \
    '{"username":"dmanias", "password":"1234"}' "http://localhost:8080/api/v1/auth");

#ask for token
token=$(curl -u dmanias:1234 -k http://localhost:8080/api/v1/auth --get -d output=json | jq '.auth_token' | tr -d '"')

#add event
res=$(curl -X POST http://localhost:8080/api/v1/events -H "accept: application/json" -H "Authorization: Bearer $token" -d '{"timestamp": "2017-11-22T08:44:22.309Z","service": "ADMINISTRATION","eventType": "event","eventNumber": 1,"status": 0,"action": "creation","customerName": "Babis","customerID": "12354","customerRights": "USER"}')

#query DB for events, with metrics
out=$(curl -w "@curl-format.txt" -o results.txt -X GET "http://localhost:8080/api/v1/events?eventType=event&service=ADMINISTRATION" -H "accept: application/json" -H "Authorization: Bearer $token")

#curl -v for verbose output