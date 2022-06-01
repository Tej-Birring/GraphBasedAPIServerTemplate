package configs

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

var Configs = struct {
	TwilioPhoneNumber            string
	AuthTokenValidForMins        int
	VerificationCodeValidForMins int
	AppName                      string
	EmailFromAddress             string
	EmailFromName                string
	VerificationEmailTemplateId  int
	Port                         int
}{}

func InitializeConfigs() {
	// Load credential from environment vars
	err := godotenv.Load("configs/.env")
	if err != nil {
		log.Println("Could not find .env file. Assuming environment variables have already been set...")
	}

	// Read configs from JSON
	configData, err := ioutil.ReadFile("configs/.configs.json")
	if err != nil {
		log.Fatal("Failed to read config file! " + err.Error())
	}
	err = json.Unmarshal(configData, &Configs)
	if err != nil {
		log.Fatal("Failed to parse config file! " + err.Error())
	}

	// Set correct port if specified...
	_portStr := os.Getenv("PORT") // <-- defined by Heroku
	if len(_portStr) < 1 {
		return
	}
	// Port has been specified, so we have to use it
	_portInt, err := strconv.Atoi(_portStr)
	if err != nil {
		log.Fatal("Error parsing port from env variable.")
	}
	Configs.Port = _portInt
}
