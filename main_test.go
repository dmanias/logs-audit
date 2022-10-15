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

//Test eventToBson method from main.go
//Takes the Event input and checks the bson.M output, 2 times
func TestEvent_eventToBson(t *testing.T) {

	data1 := map[string]interface{}{
		"test": "delicious",
	}

	event1 := Event{
		Service:   "BILLING",
		EventType: "event",
		Data:      data1,
		Timestamp: time.Now(),
		Tags:      "test",
	}

	byte1 := bson.M{
		"service":   event1.Service,
		"eventType": event1.EventType,
		"data":      event1.Data,
		"timestamp": event1.Timestamp,
		"tags":      event1.Tags,
	}

	data2 := map[string]interface{}{
		"test": "delicious",
	}

	event2 := Event{
		Service:   "BILLING",
		EventType: "event",
		Data:      data2,
		Timestamp: time.Now(),
		Tags:      "test",
	}

	tests := []struct {
		name   string
		fields Event
		want   bson.M
	}{
		{
			name:   "Input 1",
			fields: event1,
			want:   byte1,
		},
		{
			name:   "Input 2 with the same output as 1",
			fields: event2,
			want:   byte1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputEvent := Event{
				Timestamp: tt.fields.Timestamp,
				Service:   tt.fields.Service,
				EventType: tt.fields.EventType,
				Data:      tt.fields.Data,
				Tags:      tt.fields.Tags,
			}
			if got := inputEvent.eventToBson(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("eventToBson() = %v, want %v", got, tt.want)
			}
		})
	}
} //require equal values

//Test eventToString method from main.go
//Takes the Event input and checks the string output, 2 times
func TestEvent_eventToString(t *testing.T) {
	type fields struct {
		Timestamp time.Time
		Service   string
		EventType string
		Data      map[string]interface{}
		Tags      string
	}

	data1 := map[string]interface{}{
		"test": "delicious",
	}

	event1 := fields{
		Service:   "BILLING",
		EventType: "event",
		Data:      data1,
		Timestamp: time.Now(),
		Tags:      "test",
	}

	byte1, _ := json.Marshal(event1)

	data2 := map[string]interface{}{
		"test": "delicious",
	}

	event2 := fields{
		Service:   "BILLING",
		EventType: "event",
		Data:      data2,
		Timestamp: time.Now(),
		Tags:      "test",
	}
	byte2, _ := json.Marshal(event1)

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "Input 1",
			fields: event1,
			want:   string(byte1),
		},
		{
			name:   "Input 2 with the same output as 1",
			fields: event2,
			want:   string(byte2),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputEvent := Event{
				Timestamp: tt.fields.Timestamp,
				Service:   tt.fields.Service,
				EventType: tt.fields.EventType,
				Data:      tt.fields.Data,
				Tags:      tt.fields.Tags,
			}
			got, _ := inputEvent.eventToString()

			if got != tt.want {
				t.Errorf("eventToString() got = %v, want %v", got, tt.want)
			}
		})
	}
}
