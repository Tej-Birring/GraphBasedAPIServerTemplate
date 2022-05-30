package messaging

import (
	"GraphBasedServer/configs"
	"encoding/base64"
	"fmt"
	mailjet "github.com/mailjet/mailjet-apiv3-go"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var httpClient = http.Client{}

var twilioAccountSid string
var twilioAuthToken string
var twilioNumber string
var mailjetApiKey string
var mailjetSecretKey string

func initializeTwilio() {
	found := false
	twilioAccountSid, found = os.LookupEnv("TWILIO_ACCOUNT_SID")
	if found != true {
		panic("Could not find account ID for Twilio!")
	}
	twilioAuthToken, found = os.LookupEnv("TWILIO_AUTH_TOKEN")
	if found != true {
		panic("Could not find auth token for Twilio!")
	}
	twilioNumber = configs.Configs.TwilioPhoneNumber
}

func initializeMailjet() {
	found := false
	mailjetApiKey, found = os.LookupEnv("MAILJET_API_KEY")
	if found != true {
		panic("Could not find Mailjet API key!")
	}
	mailjetSecretKey, found = os.LookupEnv("MAILJET_SECRET_KEY")
	if found != true {
		panic("Could not find Mailjet secret key!")
	}
}

func InitializeMessaging() {
	initializeTwilio()
	initializeMailjet()
}

// SendSMS
// see: https://www.twilio.com/docs/usage/requests-to-twilio
func SendSMS(to string, body string) error {
	apiUrl := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", twilioAccountSid)

	postFormData := url.Values{}
	postFormData.Set("To", to)
	postFormData.Set("From", twilioNumber)
	postFormData.Set("Body", body)
	fmt.Println(strings.NewReader(postFormData.Encode()))
	//return nil
	rqst, err := http.NewRequest("POST", apiUrl, strings.NewReader(postFormData.Encode()))
	if err != nil {
		return err
	}
	basicAuth := base64.URLEncoding.EncodeToString([]byte(twilioAccountSid + ":" + twilioAuthToken))
	rqst.Header.Set("Authorization", "Basic "+basicAuth) // for auth
	rqst.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rqst.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(rqst)
	if err != nil {
		return err
	}

	// To Do: Handle error in response!
	bodystr, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(bodystr))

	return nil
}

func SendBasicEmail(toEmail string, toName string, subject string, textPart string, htmlPart string) error {
	client := mailjet.NewMailjetClient(mailjetApiKey, mailjetSecretKey)
	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: configs.Configs.EmailFromAddress,
				Name:  configs.Configs.EmailFromName,
			},
			To: &mailjet.RecipientsV31{
				{
					Email: toEmail,
					Name:  toName,
				},
			},
			Subject:  subject,
			TextPart: textPart,
			HTMLPart: htmlPart,
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := client.SendMailV31(&messages)
	if err != nil {
		return err
	}
	return nil
}

func SendTemplatedEmail(toEmail string, toName string, templateID int, subject string, variables *map[string]interface{}) error {
	client := mailjet.NewMailjetClient(mailjetApiKey, mailjetSecretKey)
	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: configs.Configs.EmailFromAddress,
				Name:  configs.Configs.EmailFromName,
			},
			To: &mailjet.RecipientsV31{
				{
					Email: toEmail,
					Name:  toName,
				},
			},
			TemplateID:       templateID,
			TemplateLanguage: true,
			Subject:          subject,
			Variables:        (*variables),
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := client.SendMailV31(&messages)
	if err != nil {
		return err
	}
	return nil
}
