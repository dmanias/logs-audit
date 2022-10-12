package auth

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"testing"
)

//Test generateToken from authentication.go
//Takes the string input and checks the map,error output, 2 times
func Test_GenerateToken(t *testing.T) {
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
			got, err := GenerateToken(tt.args.username, tt.args.password)
			fmt.Println(got)
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

//Test createTokenBson from authentication.go
//Takes the input and checks the output 2 times
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
				userId:      "dmanias",
				authToken:   "Abs7xoFH5S1krfdeqllHsD-M72-FNRE0-gPtkNaafzk=",
				generatedAt: "2022-10-12 06:41:58",
				expiresAt:   "2022-10-12 07:41:58",
			}},
		{name: "Input 2",
			args: args{
				userId:      "dmanias",
				authToken:   "Abs7xoFH5S1krfdeqllHsD-M72-FNRE0-gPtkNaafzk=",
				generatedAt: "2022-10-12 06:41:58",
				expiresAt:   "2022-10-12 07:41:58",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createTokenBson(tt.args.userId, tt.args.authToken, tt.args.generatedAt, tt.args.expiresAt); !reflect.DeepEqual(got, tt.want) {
				fmt.Println(got)
				t.Errorf("createTokenBson() = %v, want %v", got, tt.want)
			}
		})
	}
}

//Test validateToken from authentication.go
//Takes the string input and checks the bool,error output, 2 times
func Test_ValidateToken(t *testing.T) {
	type args struct {
		authToken string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{name: "Input 1",
			args: args{
				authToken: "dmanias",
			},
			wantErr: false},

		{name: "Input 2",
			args: args{
				authToken: "Abs7xoFH5S1krfdeqllHsD-M72-FNRE0-gPtkNaafzk=",
			},
			wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateToken(tt.args.authToken)
			fmt.Println(got)
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
