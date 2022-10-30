package handlers

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var sdb *mongo.Database
var collQosList, collQosDataset, collUsers *mongo.Collection

func loadEnv() {
	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func createDBInstance() {
	// DB connection string
	connectionURI := os.Getenv("DB_URI")

	// Database Name
	db := os.Getenv("DB_MONIQQ")
	sessionsdb := os.Getenv("DB_SESSIONS")

	// Collections name
	dbCollQosList := os.Getenv("DB_COLL_QOS_LIST")
	dbCollQosDataset := os.Getenv("DB_COLL_QOS_DATASET")
	dbCollUsers := os.Getenv("DB_COLL_USERS")

	// Set client options and context
	clientOptions := options.Client().ApplyURI(connectionURI)

	// connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB!")

	// Session db
	sdb = client.Database(sessionsdb)

	// Collections
	collQosList = client.Database(db).Collection(dbCollQosList)
	collQosDataset = client.Database(db).Collection(dbCollQosDataset)
	collUsers = client.Database(db).Collection(dbCollUsers)
}
