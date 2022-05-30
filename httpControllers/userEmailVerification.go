package httpControllers

import (
	"GraphBasedServer/configs"
	"GraphBasedServer/db"
	. "GraphBasedServer/httpMiddleware"
	"GraphBasedServer/messaging"
	"GraphBasedServer/utils"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strings"
	"time"
)

func HandleUserEmailVerification(mux *httprouter.Router) {
	mux.GET("/user/verifyEmail", JSONOnly(handleSendEmailVerificationCode))
	mux.POST("/user/verifyEmail", JSONOnly(handleConfirmEmailVerification))
}

var handleSendEmailVerificationCode = NewAuthHandle(func(tknData *map[string]interface{}, userId string, r *http.Request, p httprouter.Params) AuthHandleResponse {
	// get email from tkn
	email, found := (*tknData)["email"]
	if found != true {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest,
			"Email address not in token.",
			"Your account is not associated with an email address. Please add an email address.")
	}
	emailStr, _ := email.(string) // assume no errors for brevity
	// get name from tkn
	var name string
	firstName := (*tknData)["firstName"]
	if firstName != nil {
		name = firstName.(string) + " "
	}
	lastName := (*tknData)["lastName"]
	if lastName != nil {
		name = name + lastName.(string)
	}
	name = strings.Trim(name, " ")
	if len(name) < 1 {
		name = "user"
	}
	// generate a token and store when it was created
	verificationCode := utils.GetRandomString2(utils.Numeric, utils.UpperCase, 6)
	verifyTokenDuration := time.Minute * time.Duration(configs.Configs.VerificationCodeValidForMins)
	verificationCodeExpires := time.Now().Add(verifyTokenDuration)
	n := db.Node{db.Labels{"User"}, db.Properties{"id": userId}}
	err := n.Update(_dbc, db.Properties{
		"emailVerificationCode":        verificationCode,
		"emailVerificationCodeExpires": verificationCodeExpires.UTC(),
	})
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusInternalServerError,
			err.Error(),
			"Failed to update your account information. Please contact us directly to resolve this issue.")
	}
	// Send the message
	err = messaging.SendTemplatedEmail(emailStr, emailStr, configs.Configs.VerificationEmailTemplateId, "Please verify your email.", &map[string]interface{}{
		"code": verificationCode,
		"name": name,
	})
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusInternalServerError,
			err.Error(),
			"Failed to send verification email. Please contact us directly to resolve this issue.")
	}
	return NewAuthHandleSuccessResponse(_dbc, userId, nil, "Verification code sent!")
})

var handleConfirmEmailVerification = NewAuthHandle(func(tknData *map[string]interface{}, userId string, r *http.Request, p httprouter.Params) AuthHandleResponse {
	// parse JSON body
	req := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest,
			"Failed to parse JSON request.",
			"Something went wrong while trying to process your verification request. Please contact us directly to resolve this issue.")
	}
	// get user submitted code
	code, found := req["code"]
	if found != true {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest,
			"Verification code not specified in request body (JSON).",
			"Please provide a valid verification code.")
	}
	// get emailVerificationCode and emailVerificationCodeExpires
	n := db.Node{db.Labels{"User"}, db.Properties{"id": userId}}
	node, err := n.GetOne(_dbc)
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusInternalServerError,
			err.Error(),
			"Failed to update your account information. Please contact us directly to resolve this issue.")
	}
	_emailVerificationCode, found := node.Props["emailVerificationCode"]
	if found != true {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest,
			"User account not associated with a verification code.",
			"Verification request is invalid. Please contact us directly to resolve this issue.")
	}
	emailVerificationCode := _emailVerificationCode.(string)
	_emailVerificationCodeExpires, found := node.Props["emailVerificationCodeExpires"]
	if found != true {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest,
			"User account not associated with a verification code.",
			"Verification request is invalid. Please contact us directly to resolve this issue.")
	}
	emailVerificationCodeExpires := _emailVerificationCodeExpires.(time.Time)
	// return error if code not match
	if code != emailVerificationCode {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest,
			"Verification code does not match!",
			"The verification code you submitted is incorrect. Please try again.")
	}
	// return error if time expired
	if time.Now().After(emailVerificationCodeExpires) {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest,
			"Verification code has expired!",
			"The verification you submitted has expired. Please request another.")
	}
	// otherwise we're good...
	// change db
	err = n.Update(_dbc, map[string]interface{}{
		"emailVerified":                true,
		"emailVerificationCode":        nil,
		"emailVerificationCodeExpires": nil,
	})
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusInternalServerError, err.Error(), "Something went wrong while attempting to verify your email address. Please contact us directly to resolve this issue.")
	}
	//
	return NewAuthHandleSuccessResponse(_dbc, userId, nil, "Your email address has been verified successfully!")
})
