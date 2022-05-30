package httpControllers

import (
	"GraphBasedServer/configs"
	"GraphBasedServer/db"
	. "GraphBasedServer/httpMiddleware"
	"GraphBasedServer/messaging"
	"GraphBasedServer/utils"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
)

func HandleUserPhoneVerification(mux *httprouter.Router) {
	mux.GET("/user/verifyPhone", JSONOnly(handleSendPhoneVerificationCode))
	mux.POST("/user/verifyPhone", JSONOnly(handleConfirmPhoneVerification))
}

var handleSendPhoneVerificationCode = NewAuthHandle(func(tknData *map[string]interface{}, userId string, r *http.Request, p httprouter.Params) AuthHandleResponse {
	phone, found := (*tknData)["phone"]
	if found != true {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest,
			"No phone number in user token.",
			"Your account is not associated with a phone number. Please add a phone number.")
	}
	phoneStr, ok := phone.(string)
	if ok != true {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest,
			"Failed to parse phone number to string.",
			"The phone number associated with your account seems invalid. Please contact us directly to resolve this issue.")
	}
	// Compute verify token duration
	verifyTokenDuration := time.Minute * time.Duration(configs.Configs.VerificationCodeValidForMins)
	// Generate a token and store when it was created
	verificationCode := utils.GetRandomString2(utils.Numeric, utils.UpperCase, 6)
	verificationCodeExpires := time.Now().Add(verifyTokenDuration)
	n := db.Node{db.Labels{"User"}, db.Properties{"id": userId}}
	err := n.Update(_dbc, db.Properties{
		"phoneVerificationCode":        verificationCode,
		"phoneVerificationCodeExpires": verificationCodeExpires.UTC(),
	})
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusInternalServerError,
			err.Error(),
			"Failed to update your account information. Please contact us directly to resolve this issue.")
	}
	// Send the message
	msg := fmt.Sprintf("Hello from %s 👋\nYour verification code is: %s", configs.Configs.AppName, verificationCode)
	err = messaging.SendSMS(phoneStr, msg)
	//
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest,
			err.Error(),
			"Failed to send SMS to your phone number. Please contact us directly to resolve this issue.")
	}
	return NewAuthHandleSuccessResponse(_dbc, userId, nil, "Verification code sent!")
})

var handleConfirmPhoneVerification = NewAuthHandle(func(tknData *map[string]interface{}, userId string, r *http.Request, p httprouter.Params) AuthHandleResponse {
	// parse JSON body
	req := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest,
			"Failed to parse JSON request.",
			"Something went wrong while trying to process your verification request. Please contact us directly to resolve this issue.")
	}
	codeStr, found := req["code"]
	if found != true {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest,
			"Verification code not specified in request body (JSON).",
			"Please provide a valid verification code.")
	}
	// Get user info
	n := db.Node{db.Labels{"User"}, db.Properties{"id": userId}}
	node, err := n.GetOne(_dbc)
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusInternalServerError,
			err.Error(),
			"Failed to update your account information. Please contact us directly to resolve this issue.")
	}
	// Extract verification code and expiry info
	_verificationCode, found := node.Props["phoneVerificationCode"]
	if found != true {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest,
			"User account not associated with a verification code.",
			"Verification request is invalid. Please contact us directly to resolve this issue.")
	}
	verificationCode, _ := _verificationCode.(string) // assumes no errors
	_verificationCodeExpires, found := node.Props["phoneVerificationCodeExpires"]
	verificationCodeExpires, _ := _verificationCodeExpires.(time.Time) // assumes no errors
	if found != true {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest,
			"User account not associated with a verification code.",
			"Verification request is invalid. Please contact us directly to resolve this issue.")
	}
	// compare verification codes
	if codeStr != verificationCode {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest, "Verification code does not match!", "The verification code you submitted is incorrect. Please try again.")
	}
	// check if verification code has not expired
	if time.Now().After(verificationCodeExpires) {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest, "Verification code has expired!", "The verification you submitted has expired. Please request another.")
	}
	// change db if successful
	err = n.Update(_dbc, map[string]interface{}{
		"phoneVerified":                true,
		"phoneVerificationCode":        nil,
		"phoneVerificationCodeExpires": nil,
	})
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusInternalServerError, err.Error(), "Something went wrong while attempting to verify your phone number. Please contact us directly to resolve this issue.")
	}
	//
	return NewAuthHandleSuccessResponse(_dbc, userId, nil, "Your phone number has been verified successfully.")
})
