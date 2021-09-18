package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func constructURI() string {
	username := os.Getenv("username")
	password := os.Getenv("password")
	host := "aditya.qg15l.mongodb.net/"
	URI := "mongodb+srv://%s:%s@%stesting?retryWrites=true&w=majority"
	return fmt.Sprintf(URI, username, password, host)
}

func Conn() (client *mongo.Client) {
	URI := constructURI()
	client, err := mongo.NewClient(options.Client().ApplyURI(URI))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err) // Without Mongo Connection, not much point in starting the app
	}
	return client
}

func GetCTX() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	return ctx
}

func FindLatestDoc() (cursor *mongo.Cursor, e error) {

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	collection := Conn().Database("testing").Collection("covid")
	opts := options.Find()
	opts.SetSort(bson.D{{"recordDate", -1}})
	opts.SetLimit(1)
	return collection.Find(ctx, bson.D{}, opts)
}
