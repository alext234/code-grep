package server_utils

import (
	jwt "github.com/dgrijalva/jwt-go"
	"time"
	"io/ioutil"
	"github.com/golang/glog"
  "common"
)



// keys are held in global variables in RAM
var (
	verifyKey, signKey []byte
)

func init () {

}


func LoadKeysForAccessToken (privFile string, pubFile string) {
	var err error

	signKey, err = ioutil.ReadFile(privFile)

	if err != nil {
		glog.Error ("Fatal error: failed to read priv key ")
		return
	}

	verifyKey, err = ioutil.ReadFile(pubFile)

	if err != nil {
		glog.Error ("Error reading public  key")
		return
	}
	glog.V(1).Info ("Successfully read keys for signing access tokens");

}

func  GenerateNewAccessToken (user *common.SignupUser) string {
  

	t := jwt.New(jwt.GetSigningMethod("RS256"))
	t.Claims["id"] = user.Id.Hex()

	//t.Claims["exp"] = time.Now().Add(time.Hour * 24*30).Unix()// make it long, 30 days a month
	
	t.Claims["exp"] = time.Now().Add(time.Minute * time.Duration(ConfigSections.WebServer.AccessTokenExpMinutes)).Unix()
	
	tokenString, err := t.SignedString(signKey) // sign the token
	if err!=nil {
		glog.Error  ("Error signing token for user :", user.Email," :", err)
		return ""
	}
	return  tokenString
}
 
func ParseAccessToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
	// since we only use the one private key to sign the tokens,
	// we also only use its public counter part to verify

  
  
		return verifyKey, nil
	})
	return token, err

}