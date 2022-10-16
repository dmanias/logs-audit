package eventsHelper

import (
	"fmt"
	"testing"
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
