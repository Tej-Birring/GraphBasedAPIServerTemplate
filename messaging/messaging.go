package messaging

import (
	"fmt"
	"net/http"
	"os"
	"net/url"
)

var twilioAccountSid string
var twilioAuthToken string

func initializeTwilio() {
	var found bool = false
	twilioAccountSid, found = os.LookupEnv("TWILIO_ACCOUNT_SID")
	if found != true {
		panic("Could not find account ID for Twilio!")
	}
	twilioAuthToken, found = os.LookupEnv("TWILIO_AUTH_TOKEN")
	if found != true {
		panic("Could not find auth token for Twilio!")
	}
}

func InitializeMessaging() {
	initializeTwilio()
}

// SendSMS
// see: https://www.twilio.com/docs/usage/requests-to-twilio
func SendSMS(to string, body string) bool {
	apiUrl := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", twilioAccountSid)
	//http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(""))
	resp, err := http.PostForm(apiUrl, url.Values{
		"to":{"+447472239190"},
		"body":{"This is a text.\nReally...\nIt is!!"},
	})
	resp.

	req, err := http.NewRequest("POST", url)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set(twilioAccountSid, twilioAuthToken)
}
