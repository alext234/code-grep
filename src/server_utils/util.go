package server_utils

import (
	"encoding/json"
	"net/http"
	"github.com/golang/glog"
	"regexp"
)


type ErrorJson struct {
	ErrorMessage map[string][]string `json:"errors"`
}

type SuccessJson struct {
	Message string  `json:"message"`
}

type SuccessJsonWithAccessToken struct {
	Message string  `json:"message"`
	AccessToken string `json:"access_token"`
}

type UploadSuccessJson struct {
	Message string  `json:"message"`
	ProjectId string  `json:"project_id"`
}






//////////////////////////////////////////

var ValidEmailRegExp=  regexp.MustCompile(".+@.+\\..+") // regulare expression for email address verification
var ValidUrlRegExp = regexp.MustCompile(`(http|https|git):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?`);
////////////////////////////////////////



// send status code only
func SendOnlyStatusCode (w http.ResponseWriter, statusCode int ) {

	w.WriteHeader(statusCode)

	w.Write(nil)
}




func SendJsonResponseWithStatusCode (w http.ResponseWriter, data interface{}, statusCode int ) {
	jresponse, err := json.Marshal(data)
	if err != nil {
		glog.Error ("Error marshelling json :", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	w.Write(jresponse)
}

func SendJsonResponse (w http.ResponseWriter, data interface{}) {
	SendJsonResponseWithStatusCode (w, data, http.StatusOK)
}


// using status code  BadRequest
func SendErrorWithStatusCode(w http.ResponseWriter, messages map[string][]string , statusCode int)  {
	j, err := json.Marshal(ErrorJson{ErrorMessage: messages})
	if err != nil {
		glog.Error ("Error making error  json response for msg :", messages)
	}
	w.Header().Set("Content-Type", "application/json")
	
	w.WriteHeader(statusCode)
	w.Write(j)
}

// using status code  BadRequest
func SendError(w http.ResponseWriter, messages map[string][]string )  {
	SendErrorWithStatusCode(w, messages, http.StatusBadRequest)
	
}



func SendRawError(w http.ResponseWriter, message string )  {

	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

