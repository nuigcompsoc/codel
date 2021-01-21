/*
 *	Auth and Auth Middleware
 *	Verifies tokens and also verifies if an admin (optional)
 */

package main

import (
	"encoding/json"
	//	"fmt"
	"crypto/x509"
	b64 "encoding/base64"
	"encoding/pem"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
)

var realmPublicKey = []byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAs+6jMR7u2nBq+MxtDfcH
t/jX78GcCwvLTRVOBAIOuS2XYCM26Ttxnxl9u5j4WJqlsw75XtfpPkELO36eJsUh
2yWV9yOh4JL+nDkU9SsHKxKCzmExoevZYbiq2QbngOz/25V6IdWTIVMzykmv6c9u
iYm2ZV/CAH3sE77Te9Do9QBFo6++kZJvAAzc0109Ey1H4GAKDG9/hs8T1bKE7+2+
3Vo6lQui/y2HaBRZVcda7cyRd+qf7cxqSNKHZfMMdmAP+Px6Om9yg/jKklvz88mP
Nre4Tp9k/v5LlCMvvoG8V9XBpjR5SVednIJiIlXoV5Qf3RDxK3tKiQv0qP4AvP8T
fQIDAQAB
-----END PUBLIC KEY-----`)

type SSOToken struct {
	Exp       int      `json:"exp"`
	Iat       int      `json:"iat"`
	Iss       string   `json:"iss"`
	Name      string   `json:"name"`
	UID       string   `json:"preferred_username"`
	FirstName string   `json:"given_name"`
	LastName  string   `json:"family_name"`
	Email     string   `json:"email"`
	GidNumber []int    `json:"gidNumber"`
	Groups    []string `json:"groupMapper"`
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		_, err := VerifyToken(req)
		if err != nil {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}
		next(res, req)
	}
}

func CheckUserIsAdminUser(ssoToken SSOToken) bool {
	for _, group := range ssoToken.Groups {
		if strings.Contains(group, "adminteam") {
			return true
		}
	}
	return false
}

func CheckUserIsAdminUserAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// ExtractTokenMetadata also verifies token, so there is no need for AuthMiddleware before this
		ssoToken, err := ExtractTokenMetadata(req)
		if err != nil {
			res.WriteHeader(http.StatusUnauthorized)
			return
		}

		for _, group := range ssoToken.Groups {
			if strings.Contains(group, "adminteam") {
				next(res, req)
				return
			}
		}

		res.WriteHeader(http.StatusUnauthorized)
		return
	}
}

func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	//normally Authorization the_token_xxx
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		block, _ := pem.Decode(realmPublicKey)
		if block == nil {
			panic("failed to parse PEM block containing the public key")
		}

		rsaPub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			panic("failed to parse DER encoded public key: " + err.Error())
		}
		return rsaPub, nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func ExtractTokenMetadata(r *http.Request) (SSOToken, error) {
	_, err := VerifyToken(r)
	if err != nil {
		return SSOToken{}, err
	}

	tokenString := ExtractToken(r)
	tokenStringArr := strings.Split(tokenString, ".")
	tokenStringDecoded, _ := b64.RawStdEncoding.DecodeString(tokenStringArr[1]) // RawStdEncoding as this JWT does not contain padding at the end of the string

	var ssoToken SSOToken
	error := json.Unmarshal(tokenStringDecoded, &ssoToken)
	return ssoToken, error
}
