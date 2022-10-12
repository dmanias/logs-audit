package main

import (
	"fmt"
	"testing"
)

//Test generateToken from registration.go
//Takes the string input and checks the string,error output, 2 times
func Test_registerUser(t *testing.T) {
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "Input 1",
			args: args{
				username: "dmanias",
				password: "1234",
			},
			wantErr: true},

		{name: "Input 2",
			args: args{
				username: "dmanias",
				password: "12345",
			},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := registerUser(tt.args.username, tt.args.password)
			fmt.Println(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("registerUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("registerUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}
