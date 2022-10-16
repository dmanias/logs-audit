package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/dmanias/logs-audit/config"
	"github.com/dmanias/logs-audit/mongo"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	UserId          string `bson:"_id"`
	Username        string `bson:"username"`
	AccountPassword string `bson:"accountPassword"`
}

type Token struct {
	UserId      string `bson:"userId"`
	AuthToken   string `bson:"auth_token"`
	GeneratedAt string `bson:"generatedAt"`
	ExpiresAt   string `bson:"expiresAt"`
}

//@desc generateToken() generates a map[string]interface with the token
//@parameter {string} username. The username
//@parameter {string} password. The password
func GenerateToken(username string, password string) (map[string]interface{}, error) {
	//DB connection
	cfg := config.New()
	mongoClient, ctx, cancel, err := mongo.Connect(cfg.Database.Connector)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	defer mongo.Close(mongoClient, ctx, cancel)

	db := mongoClient.Database("db")
	usersCollection := db.Collection("users")

	user := User{}
	err = usersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.AccountPassword), []byte(password))
	if err != nil {
		return nil, errors.New("Invalid username or password.\r\n")
	}

	randomToken := make([]byte, 32)

	_, err = rand.Read(randomToken)
	if err != nil {
		return nil, err
	}

	authToken := base64.URLEncoding.EncodeToString(randomToken)

	const timeLayout = "2006-01-02 15:04:05"

	dt := time.Now()
	expiryTime := time.Now().Add(time.Minute * 60)

	generatedAt := dt.Format(timeLayout)
	expiresAt := expiryTime.Format(timeLayout)

	tokensCollection := db.Collection("authentication_tokens")

	tokenBson := createTokenBson(user.UserId, authToken, generatedAt, expiresAt)

	_, err = tokensCollection.InsertOne(ctx, tokenBson)

	if err != nil {
		return nil, err
	}

	tokenDetails := map[string]interface{}{
		"token_type":   "Bearer",
		"auth_token":   authToken,
		"generated_at": generatedAt,
		"expires_at":   expiresAt,
	}

	return tokenDetails, nil
}

func createTokenBson(userId string, authToken string, generatedAt string, expiresAt string) bson.M {
	bson := bson.M{
		"userId":      userId,
		"authToken":   authToken,
		"generatedAt": generatedAt,
		"expiresAt":   expiresAt,
	}
	return bson
}
func ValidateToken(authToken string) (bool, error) {

	//Connect to DB
	cfg := config.New()
	mongoClient, ctx, cancel, err := mongo.Connect(cfg.Database.Connector)
	if err != nil {
		log.Error(err)
		return false, err
	}

	defer mongo.Close(mongoClient, ctx, cancel)

	db := mongoClient.Database("db")

	tokensCollection := db.Collection("authentication_tokens")

	token := Token{}

	err = tokensCollection.FindOne(ctx, bson.M{"authToken": authToken}).Decode(&token)
	if err != nil {
		return false, err
	}

	const timeLayout = "2006-01-02 15:04:05"

	expiryTime, _ := time.Parse(timeLayout, token.ExpiresAt)
	currentTime, _ := time.Parse(timeLayout, time.Now().Format(timeLayout))

	if expiryTime.Before(currentTime) {
		return false, errors.New("The token is expired.\r\n")
	}

	return true, nil
}
