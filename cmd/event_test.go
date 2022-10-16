package main

import (
	"go.mongodb.org/mongo-driver/bson"
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
	tests := []struct {
		name    string
		fields  fields
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
			if err := event.store(); (err != nil) != tt.wantErr {
				t.Errorf("store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_search(t *testing.T) {
	type args struct {
		query bson.M
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
			got, err := search(tt.args.query)
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

func Test_writeToFile(t *testing.T) {
	type args struct {
		jsonInput string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeToFile(tt.args.jsonInput)
		})
	}
}
