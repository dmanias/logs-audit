#!/bin/bash -x
#register user
res=$(curl -X POST -F 'username=dmanias' -F 'password=1234' http://localhost:8080/api/v1/auth)
echo $res

#ask for token
token=$(curl -u dmanias:1234 -k http://localhost:8080/api/v1/auth --get -d output=json | jq '.auth_token' | tr -d '"')
echo $token

#add event
res=$(curl -X POST http://localhost:8080/api/v1/events -H "accept: application/json" -H "Authorization: Bearer $token" -d '{"timestamp": "2017-11-22T08:44:22.309Z","service": "ADMINISTRATION","eventType": "event","eventNumber": 1,"status": 0,"action": "creation","customerName": "Babis","customerID": "12354","customerRights": "USER"}')
echo $res

#query DB for events
out=$(curl -X GET "http://localhost:8080/api/v1/events?eventType=event&service=ADMINISTRATION" -H "accept: application/json" -H "Authorization: Bearer $token")
echo $out

#curl -v for verbose output