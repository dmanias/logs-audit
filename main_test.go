package main

import (
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func Test_writeToFile(t *testing.T) {
	type args struct {
		json bson.M
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeToFile(tt.args.json)
		})
	}
}
