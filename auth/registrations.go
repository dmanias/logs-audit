package auth

import (
	"github.com/dmanias/logs-audit/config"
	"github.com/dmanias/logs-audit/mongo"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

//@desc registerUser() register a new user
//@parameter {string} username. The username
//@parameter {string} password. The password
func RegisterUser(username string, password string) (string, error) {
	//DB connection
	cfg := config.New()
	mongoClient, ctx, cancel, err := mongo.Connect(cfg.Database.Connector)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	defer mongo.Close(mongoClient, ctx, cancel)

	db := mongoClient.Database("db")
	usersCollection := db.Collection("users")

	userBson, err := createUserBson(username, password)
	if err != nil {
		log.Error(err)
		return "", err
	}

	_, err = usersCollection.InsertOne(ctx, userBson)

	if err != nil {
		log.Error(err)
		return "", err
	}

	return "Success", nil
}

//@desc createUserBson() creates a bson.M from the user credentials
//@parameter {string} username. The username
//@parameter {string} password. The password
func createUserBson(username string, password string) (bson.M, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return bson.M{}, err
	}
	bson := bson.M{
		"username":        username,
		"accountPassword": hashedPassword,
	}
	return bson, nil
}
