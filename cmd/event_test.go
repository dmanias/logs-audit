package main

import (
	"testing"
	"time"
)

func TestEvent_store(t *testing.T) {

	map1 := map[string]interface{}{"data1": 1, "data2": 2}
	event1 := Event{
		Timestamp: time.Now(),
		Service:   "ADMINISTRATION",
		EventType: "event",
		Data:      map1,
		Tags:      "curl",
	}

	event2 := Event{
		Timestamp: time.Now(),
		Service:   "ADMINISTRATION",
		EventType: "event",
		Data:      map1,
		Tags:      "test2",
	}

	tests := []struct {
		name    string
		event   Event
		wantErr bool
	}{
		{name: "Input 1",
			event: event1,
		},
		{name: "Input 2",
			event: event2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := Event{
				Timestamp: tt.event.Timestamp,
				Service:   tt.event.Service,
				EventType: tt.event.EventType,
				Data:      tt.event.Data,
				Tags:      tt.event.Tags,
			}
			if err := event.store(); (err != nil) != tt.wantErr {
				t.Errorf("store() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
