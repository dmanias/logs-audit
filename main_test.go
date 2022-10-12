package main

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"testing"
	"time"
)

//Test writeToFile from main.go
//Takes the string input and writes is to mongo/temp.json file, 2 times
func Test_writeToFile(t *testing.T) {
	type args struct {
		jsonInput string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "Input 1",
			args: args{
				jsonInput: `{"page": 1, "fruits": ["apple", "peach"]}`,
			}},
		{name: "Input 2",
			args: args{
				jsonInput: `{"page": 2, "fruits": ["apple", "peach"]}`,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(tt.args.jsonInput)
			writeToFile(tt.args.jsonInput)
		})
	}
}

//Test createEventString from main.go
//Takes the Event input and checks the string,error output, 2 times
func Test_createEventString(t *testing.T) {
	data1 := map[string]interface{}{
		"test": "delicious",
	}

	event1 := Event{
		Service:   "BILLING",
		EventType: "event",
		Data:      data1,
		Timestamp: time.Now(),
	}

	byte1, _ := json.Marshal(event1)

	data2 := map[string]interface{}{
		"test": "delicious",
	}

	event2 := Event{
		Service:   "BILLING",
		EventType: "event",
		Data:      data2,
		Timestamp: time.Now(),
	}

	byte2, _ := json.Marshal(event2)
	type args struct {
		event Event
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Input 1",
			args: args{
				event: event1,
			},
			want: string(byte1),
		},
		{
			name: "Input 2",
			args: args{
				event: event2,
			},
			want: string(byte2),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createEventString(tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("createEventString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("createEventString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

//Test createEventBson from main.go
//Takes the Event input and checks the bson.M output, 2 times
func Test_createEventBson(t *testing.T) {
	data1 := map[string]interface{}{
		"test": "delicious",
	}

	event1 := Event{
		Service:   "BILLING",
		EventType: "event",
		Data:      data1,
		Timestamp: time.Now(),
	}

	byte1 := bson.M{
		"timestamp": event1.Timestamp,
		"service":   event1.Service,
		"eventType": event1.EventType,
		"data":      event1.Data,
		"tags":      bson.A{"coding", "test"},
	}

	data2 := map[string]interface{}{
		"test": "delicious",
	}

	event2 := Event{
		Service:   "BILLING",
		EventType: "event",
		Data:      data2,
		Timestamp: time.Now(),
	}
	/*
		byte2 := bson.M{
			"timestamp": event2.Timestamp,
			"service":   event2.Service,
			"eventType": event2.EventType,
			"data":      event2.Data,
			"tags":      bson.A{"coding", "test"},
		}*/

	type args struct {
		inputEvent Event
	}
	tests := []struct {
		name string
		args args
		want bson.M
	}{
		{
			name: "Input 1",
			args: args{
				inputEvent: event1,
			},
			want: byte1,
		},
		{
			name: "Input 2 with the same output as 1",
			args: args{
				inputEvent: event2,
			},
			want: byte1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createEventBson(tt.args.inputEvent); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createEventBson() = %v, want %v", got, tt.want)
			}
		})
	}
}
