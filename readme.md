# Graph database-based API Server Template
 Template for a business API server based on neo4j graph database to store and query data.

## Features
- [X] JWT-based authentication with either phone or email address, and password
- [X] Email address verification (code sent via templated email)
- [X] Phone number verification (code sent via SMS)
- [X] Messaging: SMS (via Twilio integration)
- [X] Messaging: Basic & Templated Emails (via Mailjet integration)
- [X] Query neo4j nodes by specifying any combination of labels and properties
- [ ] **TODO** Query neo4j relationships by specifying any combination of directionality, labels, properties \
      including node labels and properties.
- [ ] **TODO** Payments: One-off and regular (via Stripe integration)
- [ ] Get status of sent emails and SMS. 
- [ ] **TODO** Randomized tests for workflows: \
      - User registration, login, and refresh \
      - Email verification (includes 'send email' test) \
      - Phone verification (includes 'send SMS' test)
- [ ] Serve static files.

## How to use this template
1. Clone this repository to a local directory.
2. Create a `.jwkSigPairSet.json` file in `./configs`. \
    This will be the key that the server will use to sign (and verify) JWK tokens used for authentication. \
    You may create the key using [this tool.](https://mkjwk.org/).
3. Create a `.env` file in `./configs`, specifying environment vars:
   - NEO4J_URI — URI of the neo4j database.
   - NEO4J_USERNAME — Username to log into the neo4j database.
   - NEO4J_PASSWORD — Password to log into the neo4j database.
   - MAILJET_API_KEY — To use Mailjet to send emails. Obtain from web console.
   - MAILJET_SECRET_KEY — To use Mailjet to send emails. Obtain from web console.
   - TWILIO_ACCOUNT_SID — To use Twilio to send SMS. Obtain from web console.
   - TWILIO_AUTH_TOKEN — To use Twilio to send SMS. Obtain from web console.
4. Create a `configs.go` file in `./configs`, paste the following:

```
package configs

import (
	"github.com/joho/godotenv"
	"log"
	"os"
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
}{
	"<Twilio phone number here>",
	20,
	10,
	"<App name here, spaces allowed>",
	"<From email address>",
	"<From email name>",
	<Mailjet template ID for verification email>
	8080,
}

func InitializeConfigs() {
	// Load environment vars
	err := godotenv.Load("configs/.env")
	if err != nil {
		log.Fatal("Error loading .env file.")
	}

	// Use port 80 if RUN_IN_PRODUCTION is defined in env
	productionMode := os.Getenv("RUN_IN_PRODUCTION")
	if productionMode == "true" {
		Configs.Port = 80
	}
}
```


