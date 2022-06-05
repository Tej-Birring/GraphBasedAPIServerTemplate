package auth

import (
	"GraphBasedServer/configs"
	"GraphBasedServer/db"
	"GraphBasedServer/utils"
	"crypto/sha512"
	"encoding/base64"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"time"
)

var authTokenDuration time.Duration

var SigKeySetPrv jwk.Set
var SigKeySetPub jwk.Set

func InitializeAuth() {
	var err error
	//  load signature key sets
	SigKeySetPrv, err = jwk.ReadFile("configs/.jwkSigPairSet.json")
	if err != nil {
		panic("Error loading JWK set from file.")
	}
	SigKeySetPub, err = jwk.PublicSetOf(SigKeySetPrv)
	if err != nil {
		panic("Error producing public JWK set from private one.")
	}
	// define token
	authTokenDuration = time.Minute * time.Duration(configs.Configs.AuthTokenValidForMins)
	if authTokenDuration < time.Minute {
		panic("Invalid VERIFICATION_CODE_VALID_FOR_MINS. Please check env!")
	}
}

func NewAuthToken(controller *db.Controller, id string, prvKey interface{}) ([]byte, error) {
	n, err := db.GetById(controller, db.Labels{"User"}, id)
	if err != nil {
		return nil, err
	}
	user := n.Props
	// produce the token
	timeNow := time.Now()
	timeExpire := timeNow.Add(authTokenDuration)
	tknBuilder := jwt.NewBuilder().
		Issuer(configs.Configs.AppName + " API Server").
		IssuedAt(timeNow).
		Audience([]string{configs.Configs.AppName + " App User"}).
		Subject(configs.Configs.AppName + " Client-Side App").
		Expiration(timeExpire)

	// append all user props apart from password and salt
	for k, v := range user {
		if k == "password" || k == "salt" {
			continue
		}
		tknBuilder.Claim(k, v)
	}

	tkn, err := tknBuilder.Build()
	if err != nil {
		return nil, err
	}

	// print token as JSON
	//jsonBytes, err := json.MarshalIndent(tkn, "", "  ")
	//log.Printf("%s\n", jsonBytes)

	// produce signed (and base64 encoded) token
	return jwt.Sign(tkn, jwa.RS512, prvKey)
}

type password = string
type salt = string

// HashPasswordNewSalt mine new salt & hash the password
func HashPasswordNewSalt(newPassword string) (salt, password) {
	salt := utils.GetRandomString(256)
	hasher := sha512.New()
	hasher.Write([]byte((salt + newPassword)))
	password := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))
	return salt, password
}

func HashPasswordExistingSalt(passwordToCheck string, salt string) string {
	hasher := sha512.New()
	hasher.Write([]byte((salt + passwordToCheck)))
	hashedPasswordToCheck := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))
	return hashedPasswordToCheck
}

func CheckPasswordMatch(passwordToCheckUnhashed string, currentPasswordHashed string, currentSalt string) bool {
	hasher := sha512.New()
	hasher.Write([]byte((currentSalt + passwordToCheckUnhashed)))
	passwordToCheckHashed := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))
	return passwordToCheckHashed == currentPasswordHashed
}
