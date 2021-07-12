package model

import "github.com/dgrijalva/jwt-go"

type Token struct {
	Uid string
	Role string
	Profile string
	Groups string
	Local string
	Type string
	Session string
	jwt.StandardClaims
}

type Roles struct {
	Title string
	Uid string
}