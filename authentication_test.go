package main

import (
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"testing"
)

//Test generateToken from authentication.go
//Check if writes 2 jsons in the file when the db is down
func Test_generateToken(t *testing.T) {
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{name: "Input 1",
			args: args{
				username: "dmanias",
				password: "1234",
			}},
		{name: "Input 2",
			args: args{
				username: "dmanias",
				password: "12345",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateToken(tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createTokenBson(t *testing.T) {
	type args struct {
		userId      string
		authToken   string
		generatedAt string
		expiresAt   string
	}
	tests := []struct {
		name string
		args args
		want bson.M
	}{
		{name: "Input 1",
			args: args{
				userId:    "dmanias",
				authToken: "1234",
				generatedAt string
				expiresAt   string
			}},
		{name: "Input 2",
			args: args{
				username: "dmanias",
				password: "12345",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createTokenBson(tt.args.userId, tt.args.authToken, tt.args.generatedAt, tt.args.expiresAt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createTokenBson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateToken(t *testing.T) {
	type args struct {
		authToken string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateToken(tt.args.authToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validateToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}
