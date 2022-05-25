package db

import (
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"log"
	"os"
)

type Controller struct {
	driver neo4j.Driver
}

func IsDBError(err *error) bool {
	return neo4j.IsNeo4jError(*err)
}

func IsThisDBError(err *error, neo4jStatusCode string) bool {
	return neo4j.IsNeo4jError(*err) && (*err).(*neo4j.Neo4jError).Code == neo4jStatusCode
}

func InitializeDB() Controller {
	// parse var
	neo4jUri, found := os.LookupEnv("NEO4J_URI")
	if !found {
		panic("NEO4J_URI not set!")
	}
	neo4jUsername, found := os.LookupEnv("NEO4J_USERNAME")
	if !found {
		panic("NEO4J_USERNAME not set!")
	}
	neo4jPassword, found := os.LookupEnv("NEO4J_PASSWORD")
	if !found {
		panic("NEO4J_PASSWORD not set!")
	}
	// connect to DB
	log.Println("Connecting to DB...")
	driver, err := neo4j.NewDriver(neo4jUri, neo4j.BasicAuth(neo4jUsername, neo4jPassword, ""))
	if err != nil {
		panic("Failed to connect to DB!")
	}
	log.Println("Connected to DB.")
	return Controller{driver}
}
