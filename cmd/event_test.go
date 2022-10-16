package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	"testing"
	"time"
)

func TestEvent_store(t *testing.T) {
	type fields struct {
		Timestamp time.Time
		Service   string
		EventType string
		Data      map[string]interface{}
		Tags      string
	}
	type args struct {
		client *mongo.Client
		ctx    context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := Event{
				Timestamp: tt.fields.Timestamp,
				Service:   tt.fields.Service,
				EventType: tt.fields.EventType,
				Data:      tt.fields.Data,
				Tags:      tt.fields.Tags,
			}
			if err := event.store(tt.args.client, tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_search(t *testing.T) {
	type args struct {
		client *mongo.Client
		ctx    context.Context
		query  bson.M
	}
	tests := []struct {
		name    string
		args    args
		want    []bson.M
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := search(tt.args.client, tt.args.ctx, tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("search() got = %v, want %v", got, tt.want)
			}
		})
	}
}

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
