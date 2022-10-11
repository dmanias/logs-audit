package main

import (
	"fmt"
	"testing"
)

//Test writeToFile from main
//Check if writes 2 jsons in the file when the db is down

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

func Test_createEventString(t *testing.T) {
	type args struct {
		event Event
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "Input 1",
			event: Event{
				Timestamp: 1257894000,
				Service:   "BILLING",
				EventType: "event",
				Data:      map[string]interface{"foo": "foo", "bar": "bar"},
			}},
		{},
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
