package main

import "testing"

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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := registerUser(tt.args.username, tt.args.password)
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
