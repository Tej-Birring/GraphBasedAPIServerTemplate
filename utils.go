package main

import (
	//"fmt"
	"github.com/dchest/uniuri"
)

// TODO: Implement own random string gen function
//const charsetAlpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
//const charsetAlphaNum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
//const charsetPassword = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!£$%^&*_+-=~#@?|"
//
//type RandomStringOption int64
//
//const (
//	Alpha RandomStringOption = iota
//	Alphanumeric
//	Password
//)

func RandomString(length int) string {
	return uniuri.NewLen(length)
}
