package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j/dbtype"
	"log"
	"net/http"
	"time"
)

var _db *neo4j.Driver
var _prvSet *jwk.Set
var _pubSet *jwk.Set
var tokenDuration = time.Minute * 20

func HandleUsers(mux *httprouter.Router, db *neo4j.Driver, prvSet *jwk.Set, pubSet *jwk.Set) {
	_db = db
	_prvSet = prvSet
	_pubSet = pubSet
	mux.POST("/user", register)     // create
	mux.POST("/user/login", login)  // read
	mux.PATCH("/user", updateUser)  // update
	mux.DELETE("/user", deleteUser) // delete
	mux.POST("/user/authTokenValid", isAuthTokenValid)
}

type UserAccountResponse struct {
	IsError             bool
	Token               interface{}
	HttpStatusCode      int
	HttpStatusMessage   string
	Reason              string
	UserFriendlyMessage string
}

func GetNewToken(tkn *jwt.Token) ([]byte, error) {
	cred := GetUserQueryCredentials(tkn)
	log.Println(*tkn, cred)
	query := fmt.Sprintf("MATCH (p:Person {%s:'%s'}) RETURN p", cred.queryByKey, cred.queryByValue)
	// run query
	dbSess := (*_db).NewSession(neo4j.SessionConfig{})
	defer dbSess.Close()
	res, err := dbSess.Run(query, nil)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single()
	if err != nil {
		return nil, err
	}
	// extract person details from result
	person := (rec.Values[0].(dbtype.Node)).Props
	// create and return token
	return produceSignedToken(person)
}

func produceSignedToken(person map[string]interface{}) ([]byte, error) {
	// produce the token
	timeNow := time.Now()
	timeExpire := timeNow.Add(tokenDuration)
	t, err := jwt.NewBuilder().
		Issuer("Hayabusa API Server").
		IssuedAt(timeNow).
		Audience([]string{"Hayabusa App User"}).
		Subject("Hayabusa Client-Side App").
		Expiration(timeExpire).
		Claim("nId", person["nId"]).
		Claim("name", person["name"]).
		Claim("email", person["email"]).
		Claim("email_verified", person["email_verified"]).
		Claim("phone", person["phone"]).
		Claim("phone_verified", person["phone_verified"]).
		Build()
	if err != nil {
		return nil, err
	}

	// print as JSON
	//jsonBytes, err := json.Marshal(t) // for readability: json.MarshalIndent(t, "", "  ")
	//log.Printf("%s\n", jsonBytes)

	// produce signed (and base64 encoded) token
	prvKey, _ := (*_prvSet).Get(0)
	return jwt.Sign(t, jwa.RS512, prvKey)
}

func respondUserAccountSuccess(w http.ResponseWriter, signedTkn []byte, userMessage string) {
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

func respondUserAccountError(w http.ResponseWriter, httpStatusCode int, reason string, userMessage string) {
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

func login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := r.ParseForm()
	if err != nil {
		respondUserAccountError(w, http.StatusInternalServerError, err.Error(), "Something went wrong while we were trying to log you in. Please contact us directly to resolve this issue.")
		return
	}
	email := r.PostFormValue("email")
	phone := r.PostFormValue("phone")
	password := r.PostFormValue("password")
	// CHECK WE HAVE WHAT WE NEED
	if (email == "") && (phone == "") {
		respondUserAccountError(w, http.StatusBadRequest, "Neither email address nor phone number has been specified by the user.", "You must provide either an email address or a phone number to log in to your account.")
		return
	}
	if password == "" {
		respondUserAccountError(w, http.StatusBadRequest, "The user has submitted a blank password.", "You must provide a (correct) password to log in to your account.")
		return
	}
	var loginWith string
	if len(email) > 0 {
		loginWith = "email"
	} else {
		loginWith = "phone"
	}
	// TRY FETCH THE USER INFO
	// create session
	dbSess := (*_db).NewSession(neo4j.SessionConfig{})
	defer dbSess.Close()
	// generate query string
	query := `MATCH (p:Person {` + loginWith + `:$loginWith}) RETURN p`
	// execute query
	res, err := dbSess.Run(query, map[string]interface{}{
		"loginWith": r.PostFormValue(loginWith),
	})
	if err != nil {
		respondUserAccountError(w, http.StatusInternalServerError, err.Error(), "Something went wrong while we were trying to find your account.")
		return
	}
	_rec, err := res.Single()
	if err != nil {
		respondUserAccountError(w, http.StatusNotFound, err.Error(), "Either none or more than one user was found with the credentials you specified. Please contact us directly to resolve this issue.")
		return
	}
	// check credentials
	person := (_rec.Values[0].(dbtype.Node)).Props
	_password := person["password"].(string)
	_salt := person["salt"].(string)
	hasher := sha512.New()
	hasher.Write([]byte(_salt + password))
	_matchAgainst := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))
	if _password != _matchAgainst {
		respondUserAccountError(w, http.StatusUnauthorized, "User's password did not match.", "Incorrect password.")
		return
	}
	// produce the token
	person["nId"] = (_rec.Values[0].(dbtype.Node)).Id
	signedT, err := produceSignedToken(person)
	if err != nil {
		respondUserAccountError(w, http.StatusInternalServerError, "An error occurred while signing token: "+err.Error(), "Something went wrong while trying to log you in. Please contact us directly to resolve this issue.")
		return
	}
	// send token
	respondUserAccountSuccess(w, signedT, "Login successful!")
}

func register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := r.ParseForm()
	if err != nil {
		respondUserAccountError(w, http.StatusInternalServerError, err.Error(), "Something went wrong while we were trying to register you. Please contact us directly to resolve this issue.")
		return
	}
	// get params
	name := r.PostFormValue("name")
	if name == "" {
		respondUserAccountError(w, http.StatusBadRequest, "User did not provide a name.", "Please provide your name.")
		return
	}
	email := r.PostFormValue("email")
	phone := r.PostFormValue("phone")
	if email == "" && phone == "" {
		respondUserAccountError(w, http.StatusBadRequest, "User did not provide one of either an email address or a phone.", "Please provide either an email address or a phone number.")
		return
	}
	_password := r.PostFormValue("password")
	if _password == "" {
		respondUserAccountError(w, http.StatusBadRequest, "User did not enter a password to register with.", "Please provide a password.")
		return
	}
	// get some salt & hash the password
	salt := RandomString(256)
	hasher := sha512.New()
	hasher.Write([]byte((salt + _password)))
	password := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))
	// create session
	dbSess := (*_db).NewSession(neo4j.SessionConfig{})
	defer dbSess.Close()
	// generate & execute query
	var query string
	qParams := make(map[string]interface{})
	qParams["name"] = name
	qParams["password"] = password
	qParams["salt"] = salt
	qParams["id"] = RandomString(50)
	if email != "" {
		qParams["email"] = email
		query = `CREATE (p:Person {name:$name,email:$email,email_verified:false, phone:null, phone_verified:false, password:$password, salt:$salt}) RETURN p`
	} else if phone != "" {
		qParams["phone"] = phone
		query = `CREATE (p:Person {name:$name,email:null,email_verified:false, phone:$phone, phone_verified:false, password:$password, salt:$salt}) RETURN p`
	}
	res, err := dbSess.Run(query, qParams)
	if err != nil {
		respondUserAccountError(w, http.StatusInternalServerError, err.Error(), "Something went wrong while we were trying to register you. Please contact us directly to resolve this issue.")
		return
	}
	rec, err := res.Single()
	if err != nil {
		if IsDBError(&err, "Neo.ClientError.Schema.ConstraintValidationFailed") {
			respondUserAccountError(w, http.StatusBadRequest, err.Error(), "An account with this email address or phone number already exists. Please contact us directly to resolve this issue.")
			return
		} else {
			respondUserAccountError(w, http.StatusInternalServerError, err.Error(), "An unknown error occurred while we were trying to register you. Please contact us directly to resolve this issue.")
			return
		}
	}
	// produce signed token and return
	person := (rec.Values[0].(dbtype.Node)).Props
	signedT, err := produceSignedToken(person)
	if err != nil {
		respondUserAccountError(w, http.StatusInternalServerError, "An error occurred while signing token: "+err.Error(), "Something went wrong while trying to log you in. Please contact us directly to resolve this issue.")
		return
	}
	respondUserAccountSuccess(w, signedT, "Your account has been registered successfully")
}

var isAuthTokenValid = NewAuthProtectedHandle(func(tkn *jwt.Token, credentials UserQueryCredentials, r *http.Request, p httprouter.Params) AuthProtectedHandleResponse {
	return NewAuthProtectedHandleSuccessResponse(nil, tkn, "Token is valid!")
})

var deleteUser = NewAuthProtectedHandle(func(tkn *jwt.Token, credentials UserQueryCredentials, r *http.Request, p httprouter.Params) AuthProtectedHandleResponse {
	// create session
	dbSess := (*_db).NewSession(neo4j.SessionConfig{})
	defer dbSess.Close()
	// execute query
	query := `MATCH (p:Person {` + credentials.queryByKey + `:$value}) DELETE p`
	_, err := dbSess.Run(query, map[string]interface{}{
		"value": credentials.queryByValue,
	})
	if err != nil {
		return NewAuthProtectedHandleErrorResponse(
			true,
			true,
			http.StatusInternalServerError,
			err.Error(),
			"Something went wrong while we were trying to delete your account. Please contact us directly to resolve this issue.")
	}
	// return success
	return AuthProtectedHandleResponse{false, false, false, nil, http.StatusOK, http.StatusText(http.StatusOK), "", "Your account has ben deleted successfully."}
})

func updateUserPassword(credentials UserQueryCredentials, newPassword string) error {
	dbSess := (*_db).NewSession(neo4j.SessionConfig{})
	defer dbSess.Close()
	// get some salt & hash the password
	salt := RandomString(256)
	hasher := sha512.New()
	hasher.Write([]byte((salt + newPassword)))
	password := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))
	// produce and run query
	queryStr := `MATCH (p:Person {` + credentials.queryByKey + `:$queryByVal}) SET p.salt=$salt, p.password=$password RETURN p`
	_, err := dbSess.Run(queryStr, map[string]interface{}{
		"queryByVal": credentials.queryByValue,
		"salt":       salt,
		"password":   password,
	})
	return err
}

var userSchema = NodeSchema{
	"name":           "string",
	"email":          "string",
	"phone":          "string",
	"email_verified": "bool",
	"phone_verified": "bool",
	"password":       "string",
}

var updateUser = NewAuthProtectedHandle(func(tkn *jwt.Token, credentials UserQueryCredentials, r *http.Request, p httprouter.Params) AuthProtectedHandleResponse {
	dbSess := (*_db).NewSession(neo4j.SessionConfig{})
	defer dbSess.Close()
	_, err := dbSess.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		for k, v := range r.PostForm {
			if userSchema[k] == "" { // if the key is not in the schema, there is nothing to do
				continue
			}
			if v[0] == "" { // i.e. if value is blank, there is nothing to do!
				continue
			}
			if k == "password" { // separate execution flow for passwords because we need to hash them
				err := updateUserPassword(credentials, v[0])
				if err != nil {
					return nil, err
				}
				continue
			}
			queryStr := `MATCH (p:Person {` + credentials.queryByKey + `:$queryByVal}) SET p.` + k + `=$val RETURN p`
			_, err := dbSess.Run(queryStr, map[string]interface{}{
				"queryByVal": credentials.queryByValue,
				"val":        userSchema.ParseStringValToCorrectType(k, v[0]),
			})
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		return NewAuthProtectedHandleErrorResponse(true, true, http.StatusInternalServerError, err.Error(), "Something went wrong while we were trying to update your account information. Please contact us directly to resolve this issue.")
	}
	return NewAuthProtectedHandleSuccessResponse(nil, tkn, "Account updated successfully.")
})
