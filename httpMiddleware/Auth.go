package httpMiddleware

import (
	"HayabusaBackend/auth"
	"HayabusaBackend/db"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"log"
	"net/http"
)

/*
	AuthProtectedHandleError
	Error bool — Indicates that this is an error response
	TokenValid — If error, indicates that the token is invalid
	TokenExpired — If error, indicates that the token has expired (reason for it being invalid)
	NewToken — This is a new token sent upon every request where authentication is successful, basically replaces existing JWT with one that has an up-to-date/extended expiration time
*/
type AuthHandleResponse struct {
	IsError             bool
	TokenValid          bool
	TokenExpired        bool
	NewToken            interface{}
	HttpStatusCode      int
	HttpStatusMessage   string
	Reason              string
	UserFriendlyMessage string
	Data                interface{}
}

func NewAuthHandleSuccessResponse(dbc *db.Controller, userId string, data interface{}, userMessage string) AuthHandleResponse {
	// generate new token for the user — TODO handle error
	prvSigKey, _ := auth.SigKeySetPrv.Get(0)
	newTkn, err := auth.NewAuthToken(dbc, userId, prvSigKey)
	if err != nil {
		return NewAuthHandleErrorResponse(true, true, http.StatusInternalServerError, "Couldn't generate new token for this user session: "+err.Error(), "Failed to maintain your session. Please contact us directly to resolve this issue.")
	}
	// send response
	return AuthHandleResponse{
		false,
		true,
		false,
		string(newTkn),
		http.StatusOK,
		http.StatusText(http.StatusOK),
		"",
		userMessage,
		data,
	}
}

func NewAuthHandleErrorResponse(tokenValid bool, tokenExpired bool, httpStatusCode int, reason string, userMessage string) AuthHandleResponse {
	return AuthHandleResponse{
		true,
		tokenValid,
		tokenExpired,
		nil,
		httpStatusCode,
		http.StatusText(httpStatusCode),
		reason,
		userMessage,
		nil,
	}
}

type AuthHandleWork func(tknData *map[string]interface{}, userId string, r *http.Request, p httprouter.Params) AuthHandleResponse

func NewAuthHandle(work AuthHandleWork) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")

		pubKey, _ := auth.SigKeySetPub.Get(0)
		tkn, err := jwt.ParseRequest(r, jwt.WithVerify(jwa.RS512, pubKey)) //jwt.WithValidate(true)
		if err != nil {
			log.Printf("Token verification failed: %s\n", err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(NewAuthHandleErrorResponse(false, false, http.StatusUnauthorized, "Token failed verification with public key: "+err.Error(), "Authorisation failed. Your session could not be authenticated."))
			return
		}

		err = jwt.Validate(tkn)
		if err != nil {
			log.Printf("Token validation failed: %s\n", err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(NewAuthHandleErrorResponse(false, err == jwt.ErrTokenExpired(), http.StatusUnauthorized, "Token passed verification with public key BUT failed validation! "+err.Error(), "Authorisation failed. Your session could not be authenticated."))
			return
		}

		err = r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(NewAuthHandleErrorResponse(true, true, http.StatusInternalServerError, "Failed to parse form parameters: "+err.Error(), "Something went wrong while we were trying to fulfil your request. Please contact us directly to resolve this issue."))
			return
		}

		tknData := tkn.PrivateClaims()
		userId, found := tknData["id"]
		//fmt.Println(*tkn, userId)
		if found != true {
			NewAuthHandleErrorResponse(false, false, http.StatusBadRequest, "The received token doesn't contain a user id! This is NOT normal.", "An error occurred.")
			return
		}
		strUserId, ok := userId.(string)
		if ok != true {
			NewAuthHandleErrorResponse(false, false, http.StatusBadRequest, "The received token doesn't contain a valid user id (of type string)! This is NOT normal.", "An error occurred.")
			return
		}

		workResponse := work(&tknData, strUserId, r, p)
		if workResponse.IsError {
			w.WriteHeader(workResponse.HttpStatusCode)
			json.NewEncoder(w).Encode(workResponse)
			return
		} else {
			json.NewEncoder(w).Encode(workResponse)
			return
		}
	}
}
