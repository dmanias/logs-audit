package main

import (
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

//Test writeToFile from main
//Check if writes 2 jsons in the file when the db is down
func Test_writeToFile(t *testing.T) {

	bson1 := bson.M{
		"timestamp":      "2017-11-22T08:44:22.309Z",
		"service":        "ADMINISTRATION",
		"eventType":      "event",
		"eventNumber":    1,
		"status":         0,
		"action":         "creation",
		"customerName":   "Babis",
		"customerID":     "12354",
		"customerRights": "USER",
	}

	bson2 := bson.M{
		"timestamp":      "2017-11-22T08:44:22.309Z",
		"service":        "ADMINISTRATION",
		"eventType":      "event",
		"eventNumber":    1,
		"status":         0,
		"action":         "creation",
		"customerName":   "Babis",
		"customerID":     "12354",
		"customerRights": "USER",
	}

	type args struct {
		jsonInput bson.M
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "First input",
			args: args{
				jsonInput: bson1,
			},
		},
		{name: "Second input",
			args: args{
				jsonInput: bson2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeToFile(tt.args.jsonInput)
		})
	}
}
