package main

import (
	"context"
	"fmt"
	"github.com/dmanias/logs-audit/config"
	"github.com/dmanias/logs-audit/mongo"
	"github.com/hellofresh/health-go/v4"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

func main() {

	//	jsons, err := loadJsons()
	//	fmt.Println(jsons, err)
	handleDB()
}
func loadJsons() ([]string, error) {

	filename := "jsons.txt"
	content, err := os.ReadFile(filename)

	if err != nil {
		log.Fatal(err)
	}

	fileBody := string(content)
	split := strings.Split(fileBody, "\\n\\n")
	return split, nil
}

func handleDB() { //??
	// set global log level
	cfg := config.New()
	mongoClient, ctx, cancel, err := mongo.Connect(cfg.Database.Connector)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	defer mongo.Close(mongoClient, ctx, cancel)
	h, _ := health.New()
	err = h.Register(health.Config{
		Name:      "mongo-check",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: func(ctx context.Context) error {
			mongo.Ping(mongoClient, ctx)
			return nil
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	db := mongoClient.Database("db")
	events := db.Collection("events")

	fmt.Println("events results", events)
}
