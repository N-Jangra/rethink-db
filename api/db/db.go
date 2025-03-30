package db

import (
	"fmt"
	"log"

	"github.com/go-redis/redis"
	r "github.com/rethinkdb/rethinkdb-go"
	"github.com/spf13/viper"
)

var DB *r.Session

func InitDB() *r.Session {

	// Set the configuration file name and type (JSON)
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	// Read the config file
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Error reading config file:", err)
	}

	// Get the environment variables
	dbn := viper.GetString("DB_NAME")
	host := viper.GetString("DB_HOST")
	port := viper.GetString("DB_PORT")

	fmt.Println("Using Database:", dbn) // Debugging line

	if dbn == "" {
		log.Fatal("DB_NAME is not set in config.json")
	}

	// Establish a connection to RethinkDB
	session, err := r.Connect(r.ConnectOpts{
		Address:  fmt.Sprintf("%s:%s", host, port),
		Database: dbn, // Explicitly set the database here
	})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	// Switch to the specified database
	tables, err := r.DB(dbn).TableList().Run(session) // dbn to select the database/db name
	if err != nil {
		log.Fatal("Failed to use the specified database:", err)
	}
	defer tables.Close()

	fmt.Println("Successfully connected to RethinkDB.")

	// Store the session for later use
	DB = session
	return session
}

var redisClient *redis.Client

func InitRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	return
}

func GetRedisClient() *redis.Client {
	return redisClient
}
