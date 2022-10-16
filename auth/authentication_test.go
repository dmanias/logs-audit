package auth

import (
	"go.mongodb.org/mongo-driver/bson"
	"reflect"
	"testing"
)

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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createTokenBson(tt.args.userId, tt.args.authToken, tt.args.generatedAt, tt.args.expiresAt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createTokenBson() = %v, want %v", got, tt.want)
			}
		})
	}
}
