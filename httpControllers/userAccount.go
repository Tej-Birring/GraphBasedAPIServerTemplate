package httpControllers

import (
	"HayabusaBackend/auth"
	"HayabusaBackend/db"
	"HayabusaBackend/utils"
	"bytes"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"net/http"
	"strings"
)

// UserAccountRequest Only used for conversion from JSON request body into local vars, and documentation purposes
// these fields, once set, can not be nullified/deleted/unset, only the value can be updated
type UserAccountRequest struct {
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	Password      string `json:"password"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	PhoneVerified bool   `json:"phoneVerified"`
	EmailVerified bool   `json:"emailVerified"`
}

// UserAccountResponse Is used to standardize the output/response from this section of the API
type UserAccountResponse struct {
	IsError             bool
	Token               interface{}
	HttpStatusCode      int
	HttpStatusMessage   string
	Reason              string
	UserFriendlyMessage string
}

var _dbc *db.Controller

func HandleUserAccount(mux *httprouter.Router, dbController *db.Controller) {
	_dbc = dbController
	// note: handler for read not required because info available in signed JWT
	mux.POST("/user", handleRegister) // create
	mux.POST("/login", handleLogin)   // "read" #1 => start user session by getting JWT
	//mux.GET("/user", handleIsAuthTokenValid)	// "read" #2 => get status of JWT
	mux.PATCH("/user", handleUpdate)       // update
	mux.POST("/user/delete", handleDelete) // delete, DELETE doesn't seem to be include form data in some API clients
	//mux.GET("/user/resetPassword", handleSendResetPasswordInstructions)
	//mux.POST("/user/resetPassword", handleResetPassword)
}

func respondWithUserAccountSuccess(w http.ResponseWriter, signedTkn []byte, userMessage string) {
	response := UserAccountResponse{
		false,
		string(signedTkn),
		http.StatusOK,
		http.StatusText(http.StatusOK),
		"",
		userMessage,
	}
	w.WriteHeader(response.HttpStatusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func respondWithUserAccountError(w http.ResponseWriter, httpStatusCode int, reason string, userMessage string) {
	response := UserAccountResponse{
		true,
		nil,
		httpStatusCode,
		http.StatusText(httpStatusCode),
		reason,
		userMessage,
	}
	w.WriteHeader(response.HttpStatusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleRegister(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	req := UserAccountRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithUserAccountError(w, http.StatusBadRequest, err.Error(), "Something went wrong while we were trying to register your account. Please contact us directly to resolve this issue.")
		return
	}

	_email := req.Email
	_phone := req.Phone
	_password := req.Password

	if len(_email) < 1 && len(_phone) < 1 {
		respondWithUserAccountError(w, http.StatusBadRequest, "No email address or phone number entry in form.", "One of either an email address or a phone number is required to register an account.")
		return
	}
	if len(_password) < 1 {
		respondWithUserAccountError(w, http.StatusBadRequest, "No password entry in form.", "A password is required to register an account.")
		return
	}

	salt, password := auth.HashPasswordNewSalt(_password)

	id := utils.GetRandomString(50)

	n := db.Node{MatchLabels: db.Labels{"User"}, MatchProperties: db.Properties{
		"id":            id,
		"firstName":     req.FirstName,
		"lastName":      req.LastName,
		"phone":         _phone,
		"email":         _email,
		"phoneVerified": false,
		"emailVerified": false,
		"salt":          salt,
		"password":      password,
	}}
	err = n.Create(_dbc)

	if err != nil {
		respondWithUserAccountError(w, http.StatusInternalServerError, err.Error(), "Something went wrong while we were trying to register your account. Please contact us directly to resolve this issue.")
		return
	}

	// get token
	prvKey, _ := auth.SigKeySetPrv.Get(0)
	signedTkn, err := auth.NewAuthToken(_dbc, id, prvKey)
	if err != nil {
		respondWithUserAccountError(w, http.StatusInternalServerError, err.Error(), "Something went wrong while we were trying to register your account. Please contact us directly to resolve this issue.")
		return
	}

	// return
	respondWithUserAccountSuccess(w, signedTkn, "Congratulations! Your account has been registered!")
}

var handleUpdate = NewAuthHandle(func(tknData *map[string]interface{}, userId string, r *http.Request, p httprouter.Params) AuthHandleResponse {
	buff := bytes.Buffer{}
	_, err := buff.ReadFrom(r.Body)
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest, err.Error(), "Something went wrong while we were trying to update your account. Please contact us directly to resolve this issue.")
	}

	// We don't use the struct because we only want to update the variables specified in the request
	// If we parse struct, we are forced to update all, or omit zero values — even if the request
	// intended to set the property to a zero value!
	req := map[string]interface{}{}
	err = json.Unmarshal(buff.Bytes(), &req)
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest, err.Error(), "Something went wrong while we were trying to update your account. Please contact us directly to resolve this issue.")
	}

	// omit fields that can only be set by this server
	delete(req, "id")
	delete(req, "salt")

	// parse to JSON *only* to check input is of correct type for important fields, and below
	tmp := UserAccountRequest{}
	err = json.Unmarshal(buff.Bytes(), &tmp)
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest, err.Error(), "Something went wrong while we were trying to update your account. Please contact us directly to resolve this issue.")
	}

	// using JSON, omit fields that can not be nullified/deleted/unset once they've been set (during creation/update)
	nonNullableKeys := utils.GetJsonKeysUsedByStruct(tmp)
	for _, jsonKey := range nonNullableKeys {
		if val, ok := req[jsonKey]; ok {
			if val == nil {
				//delete(req, jsonKey)
				return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest, "Attempted to delete one or more of the following non-nullable fields: "+strings.Join(nonNullableKeys, ", "), "Something went wrong while we were trying to update your account. Please contact us directly to resolve this issue.")
			}
		}
	}

	// if request specifies a password change, we need to process it i.e. produce salt + hash
	rawPassword, found := req["password"]
	if found == true {
		// see as string
		strPassword, ok := rawPassword.(string)
		if ok != true {
			return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest, "Password could not be asserted as string.", "Incorrect password.")
		}
		// hash the password
		salt, password := auth.HashPasswordNewSalt(strPassword)
		req["salt"] = salt
		req["password"] = password
	}

	n := db.Node{db.Labels{"User"}, db.Properties{"id": userId}}

	err = n.Update(_dbc, req)
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusInternalServerError, err.Error(), "Something went wrong while we were trying to update your account. Please contact us directly to resolve this issue.")
	}

	return NewAuthHandleSuccessResponse(userId, nil, "Account updated successfully.")
})

func handleLogin(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	req := UserAccountRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithUserAccountError(w, http.StatusBadRequest, err.Error(), "Something went wrong while we were trying to log you in. Please contact us directly to resolve this issue.")
		return
	}

	// get form vals
	if err != nil {
		respondWithUserAccountError(w, http.StatusInternalServerError, err.Error(), "Something went wrong while we were trying to log you in. Please contact us directly to resolve this issue.")
		return
	}

	email := req.Email
	phone := req.Phone
	passwordToCheck := req.Password

	if len(email) < 1 && len(phone) < 1 {
		respondWithUserAccountError(w, http.StatusBadRequest, "No email address or phone number in form!", "You must provide either an email address or a phone number to log in to your account.")
		return
	}
	if len(passwordToCheck) < 1 {
		respondWithUserAccountError(w, http.StatusBadRequest, "No user password in the form!", "Please enter your password to log in.")
		return
	}
	// get user node
	var userNode *neo4j.Node
	if len(email) > 0 {
		n := db.Node{db.Labels{"User"}, db.Properties{
			"email": email,
		}}
		userNode, err = n.GetOne(_dbc)
		if err != nil {
			respondWithUserAccountError(w, http.StatusInternalServerError, err.Error(), "Failed to retrieve your account. Please contact us directly to resolve this issue.")
			return
		}
	} else if len(phone) > 0 {
		n := db.Node{db.Labels{"User"}, db.Properties{
			"phone": phone,
		}}
		userNode, err = n.GetOne(_dbc)
		if err != nil {
			respondWithUserAccountError(w, http.StatusInternalServerError, err.Error(), "Failed to retrieve your account. Please contact us directly to resolve this issue.")
			return
		}
	}
	// check credentials
	currentPasswordHashed, ok := userNode.Props["password"].(string)
	if ok != true {
		respondWithUserAccountError(w, http.StatusInternalServerError, "Error parsing password retrieved from user node.", "Something went wrong while we were trying to log you in. Please contact us directly to resolve this issue.")
		return
	}
	salt, ok := userNode.Props["salt"].(string)
	if ok != true {
		respondWithUserAccountError(w, http.StatusInternalServerError, "Error parsing salt retrieved from user node.", "Something went wrong while we were trying to log you in. Please contact us directly to resolve this issue.")
		return
	}
	match := auth.CheckPasswordMatch(passwordToCheck, currentPasswordHashed, salt)
	if match == false {
		respondWithUserAccountError(w, http.StatusBadRequest, "Incorrect password!", "The password you entered is incorrect.")
		return
	}
	// match! now get token
	userId, ok := userNode.Props["id"].(string)
	if ok != true {
		respondWithUserAccountError(w, http.StatusInternalServerError, "Error parsing id retrieved from user node.", "Something went wrong while we were trying to log you in. Please contact us directly to resolve this issue.")
		return
	}
	prvKey, _ := auth.SigKeySetPrv.Get(0)
	signedTkn, err := auth.NewAuthToken(_dbc, userId, prvKey)
	if err != nil {
		respondWithUserAccountError(w, http.StatusInternalServerError, err.Error(), "Something went wrong while we were trying to log you in. Please contact us directly to resolve this issue.")
		return
	}
	respondWithUserAccountSuccess(w, signedTkn, "Login successful!")
}

var handleDelete = NewAuthHandle(func(tknData *map[string]interface{}, userId string, r *http.Request, p httprouter.Params) AuthHandleResponse {
	req := UserAccountRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest, err.Error(), "Something went wrong while we were trying to delete your account. Please contact us directly to resolve this issue.")
	}

	passwordToCheck := req.Password
	if len(passwordToCheck) < 1 {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest, "No user password in the form!", "Please enter your password to delete your account.")
	}

	userNode, err := db.GetById(_dbc, db.Labels{"User"}, userId)

	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusInternalServerError, err.Error(), "Something went wrong while we were trying to delete your account. Please contact us directly to resolve this issue.")
	}
	currentPasswordHashed, ok := userNode.Props["password"].(string)
	if ok != true {
		return NewAuthHandleErrorResponse(true, false, http.StatusInternalServerError, "Password in user node could not be converted to string!", "Something went wrong while we were trying to delete your account. Please contact us directly to resolve this issue.")
	}
	salt, ok := userNode.Props["salt"].(string)
	if ok != true {
		return NewAuthHandleErrorResponse(true, false, http.StatusInternalServerError, "Salt in user node could not be converted to string!", "Something went wrong while we were trying to delete your account. Please contact us directly to resolve this issue.")
	}

	isPasswordMatch := auth.CheckPasswordMatch(passwordToCheck, currentPasswordHashed, salt)
	if isPasswordMatch != true {
		return NewAuthHandleErrorResponse(true, false, http.StatusBadRequest, "The password user provided did not match.", "The password you entered is incorrect.")
	}

	n := db.Node{db.Labels{"User"}, db.Properties{"id": userId}}
	err = n.Delete(_dbc)
	if err != nil {
		return NewAuthHandleErrorResponse(true, false, http.StatusInternalServerError, err.Error(), "Something went wrong while we were trying to delete your account. Please contact us directly to resolve this issue.")
	}

	return AuthHandleResponse{
		IsError:             false,
		TokenValid:          false,
		TokenExpired:        false,
		NewToken:            nil,
		HttpStatusCode:      http.StatusOK,
		HttpStatusMessage:   http.StatusText(http.StatusOK),
		Reason:              "",
		UserFriendlyMessage: "Your account has been deleted.",
		Data:                nil,
	}
})
