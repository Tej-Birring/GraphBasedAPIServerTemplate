package main

import (
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"log"
	"os"
	"strconv"
)

// IsDBError returns true if neo4j error matching status code
func IsDBError(err *error, neo4jStatusCode string) bool {
	return neo4j.IsNeo4jError(*err) && (*err).(*neo4j.Neo4jError).Code == neo4jStatusCode
}

func InitializeDB() neo4j.Driver {
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
	return driver
}

// NodeSchema these are only really used to UPDATE an entity; there must be a more efficient way of doing this:
type NodeSchema map[string]string

func (schema *NodeSchema) ParseStringValToCorrectType(propertyKey string, propertyVal string) interface{} {
	if propertyVal == "null" {
		return nil
	}
	t := userSchema[propertyKey]
	var _val interface{} = nil
	var err error = nil
	switch t {
	case "string":
		return propertyVal
	case "bool":
		_val, err = strconv.ParseBool(propertyVal)
	case "int":
		_val, err = strconv.ParseInt(propertyVal, 10, 64)
	case "float":
		_val, err = strconv.ParseFloat(propertyVal, 64)
	}
	if err != nil {
		return nil
	}
	return _val
}
