package main

import (
	"context"
	"fmt"
	"github.com/dmanias/mongo"
	"github.com/pampatzoglou/api/config"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

func main() {

	//	jsons, err := loadJsons()
	//	fmt.Println(jsons, err)
	handleConnections(dbConnect())
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

func dbConnect() *mongo2.Client { //??
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
	return mongoClient
}

func handleConnections(mongoClient *mongo2.Client) {
	quickstartDatabase := mongoClient.Database("db")
	podcastsCollection := quickstartDatabase.Collection("events")

	fmt.Println("sdsdasdadasd", podcastsCollection)
}
