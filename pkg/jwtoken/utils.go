package jwtoken

import (
	"encoding/json"
	"github.com/buildboxapp/app/pkg/model"
	"github.com/buildboxapp/lib"
	"github.com/dgrijalva/jwt-go"
)

func (s *jwtoken) Verify(tokenString string) (statue bool, body *model.Token, refreshToken string, err error) {
	var in model.Token
	var jtoken = map[string]string{}

	jsonToken, err := lib.Decrypt([]byte(s.cfg.ProjectKey), tokenString)
	if err != nil {
		return false, nil, refreshToken, err
	}

	err = json.Unmarshal([]byte(jsonToken), &jtoken)
	if err != nil {
		return false, nil, refreshToken, err
	}

	tokenAccess := jtoken["access"]
	refreshToken = jtoken["refresh"]

	token, err := jwt.ParseWithClaims(tokenAccess, &in, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.ProjectKey), nil
	})

	if !token.Valid {
		return false, nil, refreshToken, s.msg.TokenValidateFail.Error()
	}
	tbody := token.Claims.(*model.Token)

	return true, tbody, refreshToken, err
}