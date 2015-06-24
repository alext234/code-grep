package main

import (
	"config"
	"fmt"
  "github.com/gorilla/mux"
	"github.com/golang/glog"
	mgo "gopkg.in/mgo.v2"
	bson "gopkg.in/mgo.v2/bson"
	// "labix.org/v2/mgo"
	// "labix.org/v2/mgo/bson"
	
	"encoding/json"
	"net/http"
	"flag"
	"server_utils"
	"time"
  "crypto/sha1"
  "io"
	"os"
	"common"
	"strings"
	"path/filepath"
	"path"
  "strconv"
  "sync"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	"errors"




)






var (

  // mongo db related
	mgoSession    								*mgo.Session
	mgoContactFormCollection			*mgo.Collection
	mgoSignupUsersCollection			*mgo.Collection
	mgoConfirmStringCollection								*mgo.Collection	 // collection of user_id and confirm_string
	mgoResetStringCollection 								*mgo.Collection
	mgoPrjectsCollection					*mgo.Collection
	mgoProjectTreesCollection								*mgo.Collection // collection of project materialized paths, see struct TreePath
	mgoProjectDefinitionsCollection					*mgo.Collection
	mgoProjectReferencesCollection					*mgo.Collection
	mgoProjectTagsCollection								*mgo.Collection

	ConfigSections								config.WebServerConfig
)












func HandleContactUs(w http.ResponseWriter, r *http.Request) {
    
  var contactForm common.ContactForm
	
	
	decoder := json.NewDecoder(r.Body)
	
	err := decoder.Decode(&contactForm)
	
	if err != nil {
		glog.Error ("Error decode json :", err)
		server_utils.SendError (w,
		map[string][]string{"invalid_json":[]string{"something wrong with the json format"}})
		return
	}
	glog.V(1).Info (contactForm);

	// verify email address
	if matched:=server_utils.ValidEmailRegExp.Match([]byte(contactForm.Email)); !matched {
		glog.Error ("Invalid email address from contactus form :", contactForm.Email)



		return
	}
	
	contactForm.Id = bson.NewObjectId();
	contactForm.CreateTime = time.Now().UTC();
	// save to database in separate go-routine
	go func (form * common.ContactForm) {
		err := mgoContactFormCollection.Insert (form)
		
		if err !=nil {
			glog.Error("cannot save new contat form to db :", err)
		}

	} (&contactForm)
	


	
	// //  return to client
	server_utils.SendJsonResponse(w, server_utils.SuccessJson{Message:"Feedback received. We will contact you again!"} )
	
	
}

// user set new password
func HandleUpdatePassword (data map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	
	userIdHex, ok := context.GetOk(r, "user_id")
	if !ok {
		//  send back error 401
		server_utils.SendErrorWithStatusCode (w,
					map[string][]string{"message":[]string{"No user_id found"}},http.StatusUnauthorized)
		return

			
	}
	passwordPlain := data["password"].(string)
	if len(passwordPlain)==0 {
		//  send back error
		server_utils.SendError (w,
		map[string][]string{"message":[]string{"Password length is too short "}})

		return
		
	}
	
	
	// generate hashed password
	hashedPassword := generateHashedPassword (passwordPlain)
	
	// update password to db
	
	userId :=bson.ObjectIdHex(userIdHex.(string))
	change := mgo.Change{
	    ReturnNew: true,
  	  Update: bson.M{
        "$set": bson.M{
         	"hashedpassword": hashedPassword,
        }}}
	
	
	var signupUser common.SignupUser
	_, err :=  mgoSignupUsersCollection.FindId(userId).Apply(change,&signupUser)
	if err!=nil {
			server_utils.SendJsonResponseWithStatusCode (w,
		map[string][]string{"message":[]string{"Unable to update account database"}},http.StatusInternalServerError)
			return
	}
	
	// send success response
	server_utils.SendJsonResponse(w,map[string]interface{} {"Message": "Password updated successfully"})


	
}

func generateHashedPassword (plainPassword string) string {
	h := sha1.New()
	io.WriteString(h, plainPassword)
	hashedPassword  :=  fmt.Sprintf("%x", h.Sum(nil))
	return hashedPassword
	
}

// user clicks on confirmation link in email
func HandleConfirmUser (data map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	userIdHex := data["user_id"].(string)
	confirm_string := data["confirm_string"].(string)
	if userIdHex=="" ||  confirm_string==""||  !common.IsValidObjectIdHex (userIdHex) {
		// malformed json data still
		server_utils.SendError (w,
		map[string][]string{"message":[]string{"Invalid user_id or confirm_string "}})
		return
	}

	
	
	var signupUser common.SignupUser
	
	userId :=bson.ObjectIdHex(userIdHex)
  err :=  mgoSignupUsersCollection.FindId(userId).One(&signupUser)
	if err!=nil  {
		glog.Error ("error obtaining user  information :", err) // most likely not found
		server_utils.SendError (w,
		map[string][]string{"message":[]string{"Invalid user_id "}})
		return
	}
	
	if    signupUser.AccountState=="unconfirmed"{

	
		
		//err = mgoConfirmStringCollection.Remove (bson.M{"_id":userId, "confirm_string":confirm_string })
		n, err := mgoConfirmStringCollection.Find (bson.M{"_id":userId, "confirm_string":confirm_string }).Count()
		if err!=nil || n==0{
			// not found?
			
			server_utils.SendError (w,
			map[string][]string{"message":[]string{"Invalid combination of account and confirmation string"}})
			
			return
		}
		
		
		change := mgo.Change{
	    ReturnNew: true,
  	  Update: bson.M{
        "$set": bson.M{
         	"account_state": "confirmed",
        }}}
	
	
		// confirmation success, update account to confirmed
		_, err =  mgoSignupUsersCollection.FindId(userId).Apply(change,&signupUser)
		if err!=nil {
			server_utils.SendJsonResponseWithStatusCode (w,
		map[string][]string{"message":[]string{"Unable to update account database"}},http.StatusInternalServerError)
			return
		}
		
		// remove the confirm string
		err = mgoConfirmStringCollection.Remove (bson.M{"_id":userId, "confirm_string":confirm_string })
		
		// send back auth token
  tokenString :=server_utils.GenerateNewAccessToken(&signupUser)

  
  // send success response
	server_utils.SendJsonResponse(w, server_utils.SuccessJsonWithAccessToken{Message:"Account activated successfully! ", AccessToken:tokenString} )

		
		return;
	} else if  signupUser.AccountState=="confirmed" {

		// already confirmed
   	server_utils.SendError (w,
	 map[string][]string{"message":[]string{"Account is already confirmed before"}})

			
			return

		
	}
	
	server_utils.SendError (w,
	map[string][]string{"message":[]string{"Account is in unknown state"}})

}

func HandleGetUserProfile (w http.ResponseWriter, r *http.Request) {
	userIdHex, ok := context.GetOk(r, "user_id")
	if !ok {
		//  send back error 401
		server_utils.SendErrorWithStatusCode (w,
					map[string][]string{"message":[]string{"No user_id found"}},http.StatusUnauthorized)
		return

		
	} else {
		glog.V(1).Info ("user id in context =", userIdHex)
		
		var signupUser common.SignupUser

		userId :=bson.ObjectIdHex(userIdHex.(string))
		
		err :=  mgoSignupUsersCollection.FindId(userId).One(&signupUser)
		if err!=nil  {
			
     	server_utils.SendError (w,
	 map[string][]string{"message":[]string{"Unable to retrieve user profile"}})

			
			return
		}
		
		
		
		server_utils.SendJsonResponse(w,map[string]interface{} {"data": signupUser})
		
		
		
	}
	

}

func HandleVerifyResetString (data map[string]interface{}, w http.ResponseWriter, r *http.Request) {
	userIdHex := data["user_id"].(string)
	reset_string := data["reset_string"].(string)
	if userIdHex=="" ||  reset_string==""||  !common.IsValidObjectIdHex (userIdHex) {
		// malformed json data still
		server_utils.SendError (w,
		map[string][]string{"message":[]string{"Invalid user_id or reset_string "}})
		return
	}

	
	
	var signupUser common.SignupUser
	
	userId :=bson.ObjectIdHex(userIdHex)
  err :=  mgoSignupUsersCollection.FindId(userId).One(&signupUser)
	if err!=nil  {
		glog.Error ("error obtaining user  information :", err) // most likely not found
		server_utils.SendError (w,
		map[string][]string{"message":[]string{"Invalid user_id "}})
		return
	}
	
	

	
		
		n, err := mgoResetStringCollection.Find (bson.M{"user_id":userId, "reset_string":reset_string }).Count()
		if err!=nil || n==0{
			// not found?
			
			server_utils.SendError (w,
			map[string][]string{"message":[]string{"Invalid combination of account and reset string"}})
			
			return
		}
		
		
		// clear hashedpassword
		change := mgo.Change{
	    ReturnNew: true,
  	  Update: bson.M{
        "$set": bson.M{
         	"hashedpassword": "",
        }}}
	
	

		_, err =  mgoSignupUsersCollection.FindId(userId).Apply(change,&signupUser)
		if err!=nil {
			server_utils.SendJsonResponseWithStatusCode (w,
		map[string][]string{"message":[]string{"Unable to update account database"}},http.StatusInternalServerError)
			return
		}
		

		
		// remove the reset string(s)
	err = mgoResetStringCollection.Remove (bson.M{"user_id":userId })
		
		// send back auth token
  tokenString :=server_utils.GenerateNewAccessToken(&signupUser)

  
  // send success response
	server_utils.SendJsonResponse(w, server_utils.SuccessJsonWithAccessToken{Message:"Your password is successfully reset!", AccessToken:tokenString} )

		
	
}
func HandleUserLogin (w http.ResponseWriter, r *http.Request) {
	email  	:= r.FormValue("email")
	passwordPlain  	:= r.FormValue("password")
	hashedPassword := generateHashedPassword (passwordPlain)
	
	var user []common.SignupUser = make ([]common.SignupUser, 0 , 1)

	if len(email)==0|| len (passwordPlain)==0 {
		server_utils.SendError (w,
		map[string][]string{"message":[]string{"Email or password is empty"}})

		return
	}


	QueryDbWithTimeout(mgoSignupUsersCollection, bson.M{"email": email, "hashedpassword": hashedPassword,"account_state":"confirmed"} , nil, &user, 0, 1 , ConfigSections.WebServer.DbQueryTimeoutMilliseconds*3)
	



	if len(user)==0  {
		server_utils.SendError (w,
		map[string][]string{"message":[]string{"Login failed. Please try again."}})

		return
	}
	// send back auth token
	tokenString :=server_utils.GenerateNewAccessToken(&user[0])


	// send back auth token
	server_utils.SendJsonResponse(w, server_utils.SuccessJsonWithAccessToken{Message:"Successful login!", AccessToken:tokenString} )

	
		

}
func HandleResetPassword (w http.ResponseWriter, r *http.Request) {
	email  	:= r.FormValue("email")
	
	// verify email address format
	if matched:=server_utils.ValidEmailRegExp.Match([]byte(email)); !matched {
		

		server_utils.SendError (w,
		map[string][]string{"message":[]string{"Invalid email address"}})


		return
	}
	
		// check if email already exist: if not, send back error

  var existingUser common.SignupUser
  err :=  mgoSignupUsersCollection.Find(bson.M{"email": email}).One(&existingUser)
  if err != nil {
    // email not  exist
    	server_utils.SendError (w,
        	map[string][]string{"message":[]string{"The email address does not exist. Please check again"}})
    return
  }

   // generate a reset hash to send via email
	h := sha1.New()
	io.WriteString(h, time.Now().String())
	io.WriteString(h, existingUser.Id.Hex())
	io.WriteString(h, existingUser.Email)
	io.WriteString(h, "CODE/GREP")
	resetHash  :=  fmt.Sprintf("%x", h.Sum(nil))
	
	
	err = mgoResetStringCollection.Insert(bson.M{"user_id": existingUser.Id, "reset_string":resetHash})
	
	if err != nil {
		glog.Error ("Cannot save reset hash into reset string collection ", err)
		return
	}

  
  // request backend workers to do email work
  server_utils.AddToSendResetEmailQueue(&existingUser, resetHash)
  
  
  // send respose
	server_utils.SendJsonResponse(w, map[string]string{"message":"An email will be sent to your address will the instruction to reset your password."})
	

}


// HandleGetUserData send back user profile or other information depending on "type"
// also handle login when type="login"
func HandleUserLoginOrResetPassword (w http.ResponseWriter, r *http.Request) {
	getDataType := r.FormValue("type")
	glog.V(1).Info ("Get type "+ getDataType)
	if getDataType=="login" {
		HandleUserLogin(w, r)
		

		return
	 } else if getDataType=="password_reset" {
	 	HandleResetPassword (w, r)
	 	return
	 	
	 	
	 	
	 }
	
	// unknow operation Type
	server_utils.SendError (w,
	map[string][]string{"message":[]string{"Unknown operation type"}})
	
}


// handle put method; e.g. when user verify confirm string or update password etc
func HandleUserUpdate (w http.ResponseWriter, r *http.Request) {

	var putData  map[string]interface{} = make (map[string]interface{})
	
	decoder := json.NewDecoder(r.Body)
	
	err := decoder.Decode(&putData)
	
	
	
	if err != nil {
		glog.Error ("Error decode json :", err)
		server_utils.SendError (w,
		map[string][]string{"message":[]string{"something wrong with the json format"}})
		return
	}
	
	putType := putData ["type"]
	if putType=="verify_confirm_string" {
		
		HandleConfirmUser (putData,  w , r)
		
		return
		
	} else	if putType=="verify_reset_string" {
		
		HandleVerifyResetString (putData,  w , r)
		
		return
		
	} else if putType=="update_password" {
		HandleUpdatePassword (putData,  w , r)

		return
	}

	
	// unknow operation Type
	server_utils.SendError (w,
	map[string][]string{"message":[]string{"Unknown operation type"}})

	
}

// new user signup
func HandleSignup(w http.ResponseWriter, r *http.Request) {

  var signupUser common.SignupUser
	
	
	decoder := json.NewDecoder(r.Body)
	
	err := decoder.Decode(&signupUser)
	
	if err != nil {
		glog.Error ("Error decode json :", err)
		server_utils.SendError (w,
		map[string][]string{"message":[]string{"something wrong with the json format"}})
		return
	}
	

	// verify email address
	if matched:=server_utils.ValidEmailRegExp.Match([]byte(signupUser.Email)); !matched {
		glog.Error ("Invalid email address from contactus form :", signupUser.Email)

		server_utils.SendError (w,
		map[string][]string{"message":[]string{"Invalid email address"}})


		return
	}
	
	
	// check if email already exist: send back error

  var anyExistingUser common.SignupUser
  err =  mgoSignupUsersCollection.Find(bson.M{"email": signupUser.Email}).One(&anyExistingUser)
  if err == nil {
    // email already exist
    	server_utils.SendJsonResponseWithStatusCode (w,
        	map[string][]string{"message":[]string{"The email address is already in use. Please specify a different one"}}, http.StatusConflict) // HTTP 409
    return
  }
  
  
  signupUser.AccountState = "unconfirmed"
  signupUser.Id = bson.NewObjectId();
  signupUser.CreateTime = time.Now().UTC();
  signupUser.Quota = ConfigSections.Common.NewSignupQuota
  signupUser.UsedSpace = 0
 
  
	err = mgoSignupUsersCollection.Insert (signupUser)
	if err !=nil {
		glog.Error("Cannot save new signup user to db : ", err)
		
		// send internal server error
		server_utils.SendJsonResponseWithStatusCode (w,
		map[string][]string{"message":[]string{"Unable to save new signup user to DB"}},http.StatusInternalServerError)



		return
	}


  
	   // generate a confirmation hash to be sent to user email
	h := sha1.New()
	io.WriteString(h, time.Now().String())
	io.WriteString(h, signupUser.Id.Hex())
	io.WriteString(h, signupUser.Email)
	io.WriteString(h, "CODE/GREP")
	confirmHash  :=  fmt.Sprintf("%x", h.Sum(nil))
	
	err = mgoConfirmStringCollection.Insert(bson.M{"_id": signupUser.Id, "confirm_string":confirmHash})
	
	if err != nil {
		glog.Error ("Cannot save into confirm string collection ", err)
		return
	}

  
  // request backend workers to do email work
  server_utils.AddToSendConfirmEmailQueue(&signupUser, confirmHash)

  // generate access token
  tokenString :=server_utils.GenerateNewAccessToken(&signupUser)

  //glog.V(1).Info (signupUser)


  // send success response
	server_utils.SendJsonResponse(w, server_utils.SuccessJsonWithAccessToken{Message:"Thank you for signing up! We have sent you a confirmation email. Kindly follow the instructions in the email to fully activate your account and to set your password. In the meantime, you can start using and add your own projects.", AccessToken:tokenString} )


}

// new project added via url method
func HandleNewProjectInfo (w http.ResponseWriter, r *http.Request) {
	var newProject common.Project

	userIdHex, ok := context.GetOk(r, "user_id") // make sure again we get userid from session
	if !ok {
		//  send back error 401
		server_utils.SendErrorWithStatusCode (w,
					map[string][]string{"message":[]string{"No user_id found"}},http.StatusUnauthorized)
		return

			
	}

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&newProject)

	if err != nil {
		glog.Error ("Error decode json :", err)
		server_utils.SendError (w,
		map[string][]string{"message":[]string{"something wrong with the json format"}})
		return
	}
	
	//
	glog.V(1).Info ("new project added by url :",newProject.FetchUrl); // should have only FetchUrl
	// also perform trimmming of spaces
	newProject.FetchUrl = strings.TrimSpace (newProject.FetchUrl)

	// perform regex check on url again
	if match:=server_utils.ValidUrlRegExp.Match([]byte(newProject.FetchUrl)); !match {
		server_utils.SendError (w,
		map[string][]string{"message":[]string{"Invalid URL"}})

		return
	}


	newProject.Id = bson.NewObjectId()
	newProject.CreateTime = time.Now().UTC()
	newProject.Name = newProject.FetchUrl[strings.LastIndex(newProject.FetchUrl, "/")+1:];
	newProject.BallName = fmt.Sprintf ("%d", time.Now().UnixNano()) +"_"+newProject.Name
	newProject.WorkDir = "" // to be assigned and used by backend
	newProject.Status = "url_received"
	newProject.TotalSize = 0 // only updated once scanning is done by backend
	newProject.UserId =bson.ObjectIdHex(userIdHex.(string))// for back end to use
	newProject.ViewPermission = "private" // original default value
	
  // store into mongo
  err = mgoPrjectsCollection.Insert (newProject)
	if err !=nil {
		glog.Error("Cannot save project  to db : ", err)
		// send internal server error
		server_utils.SendJsonResponseWithStatusCode (w,
		map[string][]string{"message":[]string{"Unable to save new project to DB"}},http.StatusInternalServerError)

		return
	}
  

	// also update to User's ProjectList
	change := mgo.Change{
	    ReturnNew: true,
  	  Update: bson.M{
        "$push": bson.M{
         	"project_list": newProject.Id,
        }}}
	var signupUser common.SignupUser
	_, err =  mgoSignupUsersCollection.FindId(newProject.UserId).Apply(change,&signupUser)
	
	if err !=nil {
		glog.Error("Cannot add new project to user's project list  : ", err)
		server_utils.SendJsonResponseWithStatusCode (w,
		map[string][]string{"message":[]string{"Unable to add new project to user profile "}},http.StatusInternalServerError)

		return
		
		
	}
	
		_, _, _ =CalculateUsedAndRemainingSpaceForUser (bson.ObjectIdHex(userIdHex.(string))) // just to make sure it updates the database so backend can get out the information

	//glog.V(1).Info ("new project user id = ", newProject.UserId)
  // request backend to start processing new project
	server_utils.AddToProjectToBeProcessedQueue (&newProject)
	
  server_utils.SendJsonResponse(w, server_utils.UploadSuccessJson{Message:"Project URL is received. Start working on it...", ProjectId: newProject.Id.Hex()})


}




func GetUserListProjects (userId bson.ObjectId) (projectList []common.Project, err error) {
	
	var user []common.SignupUser = make ([]common.SignupUser, 0 , 1)

	
	// retrieve list of all projects belong to the user
	QueryDbWithTimeout(mgoSignupUsersCollection, bson.M{"_id": userId }, bson.M{"project_list":1}, &user, 0, 1000 , ConfigSections.WebServer.DbQueryTimeoutMilliseconds*5)
	

	
	
	if len(user)==0 {
		
		return []common.Project{}, errors.New("No user id ")
	}
	
	

	if len(user[0].ProjectList)==0 {
		
		return []common.Project{},nil
		
		
	}
	
	var tempList []common.Project = make ([]common.Project, 0 , 1000)

	QueryDbWithTimeout(mgoPrjectsCollection, bson.M{"_id":bson.M{"$in": user[0].ProjectList} ,"status": bson.M{"$ne":"error"}}, bson.M{"_id":1, "name":1, "status":1, "total_size":1 ,"view_permission":1}, &tempList, 0, cap (tempList) , ConfigSections.WebServer.DbQueryTimeoutMilliseconds*10)

	return tempList, nil
	
}

// get list of prjects belonging to a user
func HandleGetListProjects (w http.ResponseWriter, r *http.Request) {

	userIdHex, ok := context.GetOk(r, "user_id")
	if !ok {	// first check: does user own the project?
			server_utils.SendError (w,
			map[string][]string{"message":[]string{"No user id"}})
			return

	}
	
	projectList, err := GetUserListProjects(bson.ObjectIdHex(userIdHex.(string)))
	
	if err!=nil {
		server_utils.SendError (w,
		map[string][]string{"message":[]string{err.Error()}})
		
	}
	
	var responseJson map[string]interface{} = make (map[string]interface{})
	
	responseJson["projects"] = projectList

	server_utils.SendJsonResponse(w, responseJson)
	
	
}

// only handle update name at the moment
func HandleUpdateProjectInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
 	projectId :=vars["projectId"] // id string in hex  - from the url itself
	if !common.IsValidObjectIdHex (projectId) {
			server_utils.SendError (w,
			map[string][]string{"message":[]string{"Invalid projectId"}})
			return
	}
	
	var putData  map[string]interface{} = make (map[string]interface{})
	
	decoder := json.NewDecoder(r.Body)
	
	err := decoder.Decode(&putData)
	
	
	
	if err != nil {
		glog.Error ("Error decode json :", err)
		server_utils.SendError (w,
		map[string][]string{"message":[]string{"something wrong with the json format"}})
		return
	}
	
	
	
	
	var change mgo.Change
	var project  common.Project
	updateBson := bson.M{}
	anyThingToUpdate :=false
	if putData["name"]!="" && putData["view_permission"]!=""  {
		updateBson["$set"] =bson.M{	"name":putData["name"], "view_permission":putData["view_permission"] }
		anyThingToUpdate = true
	}
	
	
	if anyThingToUpdate {
	// update project status
		change = mgo.Change{
	    ReturnNew: true,
	    Update: updateBson,
		}
	        
	        
		if _, err := mgoPrjectsCollection.FindId(bson.ObjectIdHex(projectId)).Apply(change,&project); err != nil {
	    glog.Error ("error updating project message   : ", err)
	    
			server_utils.SendError (w,
				map[string][]string{"message":[]string{"There was error deleting the project"}})
	
			return
	
		}
	
		glog.V(1).Info (project)
		return
	}
	// nothing to update??
	server_utils.SendError (w,
				map[string][]string{"message":[]string{"Nothing to update"}})
	

	server_utils.SendJsonResponse (w,project)


}

// user request to delete or cancel the project
	func HandleDeleteProjectInfo (w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
 	projectId :=vars["projectId"] // id string in hex
	if !common.IsValidObjectIdHex (projectId) {
			server_utils.SendError (w,
			map[string][]string{"message":[]string{"Invalid projectId"}})
			return

	}


	// update to error state
	var change mgo.Change
	var project  common.Project
	
	// update project status
	change = mgo.Change{
    ReturnNew: true,
    Update: bson.M{
        "$set": bson.M{
         	"message": "Project was cancelled or deleted",
         	"status":"error",
        }}}
        
        
	if _, err := mgoPrjectsCollection.FindId(bson.ObjectIdHex(projectId)).Apply(change,&project); err != nil {
    glog.Error ("error updating project message   : ", err)
    
		server_utils.SendError (w,
			map[string][]string{"message":[]string{"There was error deleting the project"}})

		return

	}


	
	// send zmq PUB to backend to delete/cancel downloading or whatever processing
	server_utils.SendCommandToProjectWorkder(projectId, "cancel");
	
	server_utils.SendJsonResponse (w,project)


}

// r may or may not contain user_id injected by HandleAuth
func VerifyViewPermission ( projectId bson.ObjectId ,w http.ResponseWriter ,r *http.Request) bool {
	//TODO also cache the permissions for fast access


	var isAllowed bool =  false
	userIdHex, ok := context.GetOk(r, "user_id")
	if ok {	// first check: does user own the project?
	
		var user []common.SignupUser = make ([]common.SignupUser, 0 , 1)
		QueryDbWithTimeout(mgoSignupUsersCollection, bson.M{"_id": bson.ObjectIdHex(userIdHex.(string)),"project_list":projectId} , nil, &user, 0, 1 , ConfigSections.WebServer.DbQueryTimeoutMilliseconds*10)
		
		isAllowed = len(user)>0
	}
	
	// check if the project allows public access?
	if !isAllowed {
		
		var project []common.Project = make ([]common.Project, 0 , 1)

		QueryDbWithTimeout(mgoPrjectsCollection, bson.M{"_id": projectId} , nil, &project, 0, 1 , ConfigSections.WebServer.DbQueryTimeoutMilliseconds*5)

		if len(project)!=0 {
			if project[0].ViewPermission =="public" {
				isAllowed = true
			}
			
		}

	}
	
	
	if !isAllowed {
		server_utils.SendErrorWithStatusCode (w,
							map[string][]string{"message":[]string{"You are not allowed to access this resource. Please try to login."}}, http.StatusUnauthorized)
		return false
	}
	

	// allowed
	return true
	
}

func HandleGetProjectInfo (w http.ResponseWriter, r *http.Request) {
	
	

	vars := mux.Vars(r)
 	projectIdHex :=vars["projectId"] // id string in hex
 	
	if !common.IsValidObjectIdHex (projectIdHex) {
			server_utils.SendError (w,
			map[string][]string{"message":[]string{"Invalid projectId"}})

	}
	projectId :=bson.ObjectIdHex(projectIdHex)

	if !VerifyViewPermission(projectId,  w, r) { // not allowed to view ?
		return
	}

	var project []common.Project = make ([]common.Project, 0 , 1) // slice of 1 only

	QueryDbWithTimeout(mgoPrjectsCollection, bson.M{"_id": projectId} , nil, &project, 0, 1 , ConfigSections.WebServer.DbQueryTimeoutMilliseconds*5)


  
	if len(project)==0 {
		// not found or error occured
		server_utils.SendError (w,
			map[string][]string{"message":[]string{"There was error obtaining project information"}})
		return

	}

	updateProjectMessage (&project[0])

	server_utils.SendJsonResponse(w,project[0])



}

// note this does not update to database
func updateProjectMessage (project *common.Project) {
	if len(project.Message)>0  {
		return;
	}

	switch (project.Status) {
		case "error": //  this message should already be updated by backend
			project.Message = "There was error processing the project"

		case "uploaded":
			project.Message = "Project is received successfully. Processing will start soon..."
		case "extracting":
			project.Message = "Extracting in progress..."
		case  "url_received":
		  project.Message = "Project URL is received. Processing will start soon..."
		case "fetching":
			if project.Message =="" {
				project.Message = "Fetching project  URL in progress..."
			}
		case "tree_scanning":
			project.Message = "Scanning project directory tree..."
		case "analyzing":
			project.Message = "Analyzing files..."
		case "ready":
			project.Message = "Project is ready for browsing"

	}

}


// TODO for future improvement: all these calculation should be done in backend  - after tree scanning is done
// calculate and also update to dabase used space
func CalculateUsedAndRemainingSpaceForUser (userId bson.ObjectId) (usedSpace int64 , remainingSpace int64 ,err error){
	//
	projectList, err := GetUserListProjects(userId)

	if err!=nil {
		return 0, 0, err
	}
	// user.Quota
	var totalUsed  int64=0
	for _, project:=range projectList {
		totalUsed =totalUsed+project.TotalSize

	}
	
	///// update and also retrieve latest information
	change := mgo.Change{
	    ReturnNew: true,
  	  Update: bson.M{
        "$set": bson.M{
         	"used_space": totalUsed,
        }}}
	var user common.SignupUser
	_, err =  mgoSignupUsersCollection.FindId(userId).Apply(change,&user)
	
	if err !=nil {
		return 0,0, err
	}
	
	remainingSpace  = user.Quota - totalUsed
	if remainingSpace<0 {
		remainingSpace=0
	}
	if remainingSpace>ConfigSections.Common.MaxUploadLimit {
		ConfigSections.Common.MaxUploadLimit = ConfigSections.Common.MaxUploadLimit
	}
	return totalUsed, remainingSpace, nil


}



// return current limit for the user
func HandleCheckUploadLimit (w http.ResponseWriter, r *http.Request) {
	userIdHex, _:= context.GetOk(r, "user_id")
	
	_, remainingSpace, err :=CalculateUsedAndRemainingSpaceForUser (bson.ObjectIdHex(userIdHex.(string)))
	
	if err!=nil {
     server_utils.SendError (w,
		 map[string][]string{"message":[]string{err.Error()}})

		return
	}

	
	server_utils.SendJsonResponse(w, map[string]int64{"upload_limit":remainingSpace})
}

// handle upload
func HandleUpload(w http.ResponseWriter, r *http.Request) {

	userIdHex, ok := context.GetOk(r, "user_id") // make sure again we get userid from session
	if !ok {
		//  send back error 401
		server_utils.SendErrorWithStatusCode (w,
					map[string][]string{"message":[]string{"No user_id found"}},http.StatusUnauthorized)
		return

			
	}

  glog.V(1).Info ("Change req.Body to use MaxBytesReader to limit incoming size")
  r.Body = http.MaxBytesReader(w, r.Body, ConfigSections.Common.MaxUploadLimit)


  file, header, err := r.FormFile("file")

  if err != nil {
    glog.Error ("Error with Formfile  (maybe user cancels at client side) or size is over the limit ", err)

     server_utils.SendError (w,
		 map[string][]string{"message":[]string{"Upload is canceled or size is over the limit"}})
	
    return
  }
  
  defer file.Close()
  
  generatedFilename := fmt.Sprintf ("%d", time.Now().UnixNano()) +"_"+header.Filename
  out, err := os.Create(ConfigSections.Common.DirProjectsUploaded+"/"+generatedFilename)
  
  if err != nil {
    glog.Error ( "Unable to create the file for writing. Check your write access privilege : ", err)
    
    server_utils.SendJsonResponseWithStatusCode (w,
		map[string][]string{"message":[]string{"Unable to store uploaded file on server"}},http.StatusInternalServerError)

    
    return
  }

  defer out.Close()

	_, remainingSpace, _ :=CalculateUsedAndRemainingSpaceForUser (bson.ObjectIdHex(userIdHex.(string)))

  // write the content from POST to the file
  size, err := io.CopyN(out, file, remainingSpace) // TODO: define this limit, maybe per user
  if err != nil {
    // note: because we use CopyN, we should expect error when EOF reached
  }
  out.Sync()

  glog.V(1).Info ("File uploaded successfully : ", size)
  glog.V(1).Info (header.Filename)

  var newProject common.Project

	newProject.Id = bson.NewObjectId()
	newProject.CreateTime = time.Now().UTC()
	newProject.Name = header.Filename
	
	
	newProject.BallName = generatedFilename
	newProject.FetchUrl = ""
	newProject.WorkDir = "" // to be assigned and used by backend
	newProject.Status = "uploaded"
	newProject.UserId =bson.ObjectIdHex(userIdHex.(string))// for back end to use
	newProject.ViewPermission = "private" // original default value

  // store into mongo
  err = mgoPrjectsCollection.Insert (newProject)
	if err !=nil {
		glog.Error("Cannot save project  to db : ", err)
		// send internal server error
		server_utils.SendJsonResponseWithStatusCode (w,
		map[string][]string{"message":[]string{"Unable to save new project to DB"}},http.StatusInternalServerError)

		return
	}
  

	// also update to User's ProjectList
	change := mgo.Change{
	    ReturnNew: true,
  	  Update: bson.M{
        "$push": bson.M{
         	"project_list": newProject.Id,
        }}}
	var signupUser common.SignupUser
	_, err =  mgoSignupUsersCollection.FindId(newProject.UserId).Apply(change,&signupUser)
	
	if err !=nil {
		glog.Error("Cannot add new project to user's project list  : ", err)
		server_utils.SendJsonResponseWithStatusCode (w,
		map[string][]string{"message":[]string{"Unable to add new project to user profile "}},http.StatusInternalServerError)

		return
		
		
	}


  // request backend to start processing new project
	server_utils.AddToProjectToBeProcessedQueue (&newProject)
	
  server_utils.SendJsonResponse(w, server_utils.UploadSuccessJson{Message:"Project is received successfully. Start analyzing...", ProjectId: newProject.Id.Hex()})


}


// if isDownload is true, browser will display Save as prompt
func handleGetProjectRawOrDownload (w http.ResponseWriter, r*http.Request, isDownload bool) {
	vars := mux.Vars(r)
 	projectId :=vars["projectId"] // id string in hex
 	fullPath :="/"+vars["path"] // id string in hex
	
	if !common.IsValidObjectIdHex (projectId) {
			server_utils.SendRawError (w, "Invalid projectId") // send without json
			return
	}
		

	if !VerifyViewPermission(bson.ObjectIdHex(projectId),  w, r) { // not allowed to view ?
		return
	}

	glog.V(1).Info ("raw file for project ", projectId, "  path = ", fullPath)
	// check if project exist
	var project common.Project

  err :=  mgoPrjectsCollection.FindId(bson.ObjectIdHex(projectId)).One(&project)
	if err!=nil  {
		glog.Error ("error obtaining project information :", err) // most likely not found

		server_utils.SendRawError (w, "Project does not exist")
		return

	}
	
	// check if path exist in db
	baseName :=filepath.Base(fullPath)
	parentPath :=fullPath[:len(fullPath)-len(baseName)-1]
	glog.V(1).Info ("check if exist name = ", baseName, " with path=",parentPath)
	// query mongo
	
	
	query :=  mgoProjectTreesCollection.Find(
	bson.M{"project_id": bson.ObjectIdHex(projectId), "path":parentPath, "name":baseName, "is_dir":false })
	
	n,err := query.Count()
	if n==0 || err!=nil {
		server_utils.SendRawError (w, "Path does not exist")
		return
	}
	
	if isDownload{
	
		// try to display download prompt
		w.Header().Set("Content-Disposition", "attachment; filename="+baseName)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	}

	
	
	// now server the file
	absolutePath :=project.WorkDir + fullPath
	glog.V(1).Info ("abs path = ", absolutePath)
	http.ServeFile(w,r,absolutePath)
	

}


// get raw file content
func HandleGetProjectRaw  (w http.ResponseWriter, r *http.Request) {
	handleGetProjectRawOrDownload (w,r, false)
}


// file download
func HandleGetProjectDownload  (w http.ResponseWriter, r *http.Request) {
	handleGetProjectRawOrDownload (w,r, true)
}


//called from HandleGetProjectDetails
// NOTE: this function might be implemented again in the future. at the momemnt, autocomplete is disabled
func HandleIdentifierAutocomplete (projectIdHex string, searchString string, w http.ResponseWriter) {
	
	var results = make ([]string, 0, 50)

	resultChan := make (chan string)
	go func (maxNumItems int, resultChan chan string) {
		
		var queryItem common.Tag
		resultAlreadyAdded := make (map[string]bool) // simple set implementation to check with one already added to avoid duplication
		// option "i" for case-insensitive
		//glog.V(1).Info ("searchString =", searchString)
		query :=  mgoProjectDefinitionsCollection.Find(bson.M{"project_id": bson.ObjectIdHex(projectIdHex),  "tag":bson.M{"$regex":bson.RegEx{`^`+searchString,""} } }).Select (bson.M{"tag":1,"_id":0})
		iter := query.Limit( maxNumItems ).Iter()
		
		for iter.Next(&queryItem) {
			
			tag :=queryItem.Tag
			//glog.V(1).Info ("Tag = ", tag)
			if !resultAlreadyAdded[tag] {
				resultAlreadyAdded[tag]=true
				resultChan <- tag
			}
		}
		iter.Close()
		close(resultChan)
		
	}(cap(results), resultChan)

QUERY_WAIT:
	for {
			select {
				case <-time.After ( time.Duration(ConfigSections.WebServer.DbQueryTimeoutMilliseconds*3)*time.Millisecond):// it takes longer so multiply by 3
					glog.V(1).Info ("query timeout");
					break QUERY_WAIT
				case nextResult, ok :=<-resultChan:
					if (ok){
						results = append (results, nextResult)
					} else {
						//glog.V(1).Info ("result channel is closed")
						break QUERY_WAIT
					}
			}
	}
	
	
	var responseJson map[string]interface{} = make (map[string]interface{})
		
	responseJson["results"] = results

	
 	server_utils.SendJsonResponse(w,responseJson)

	
}


	
func QueryDbWithTimeout (  collection *mgo.Collection,  queryBson  interface{}, querySelectBson interface{}, resultList interface{}, skip int, size int, timeoutMilli int64) {
	completedChan := make (chan bool)
	go func (resultList interface{}, completedChan chan bool) {
		
		//var queryItem common.Tag
		
		
		 query :=  collection.Find(queryBson)
		 if querySelectBson!=nil {
		 	query.Select (querySelectBson)
		 }

		//maxScan :=1500000// TODO: still experimental stage; maybe later finetune this value
		

		//query.SetMaxScan(maxScan)
		 
		 query.Skip (skip)// TODO: skip is actually expensive; to improve in the future
		
		
		iter := query.Limit( size ).Iter()
		
		// var explainData bson.M
		// query.Explain (&explainData)
		// glog.V(1).Infof ("explain query %#v",explainData)
		// count, _ :=query.Count()
		// glog.V(1).Info ("query count ", count)
		
		err := query.All (resultList)
		if err!=nil {
			glog.V(1).Info ("Error querying !!", err)
		}
		
		iter.Close()
		close(completedChan)
		
	}( resultList, completedChan)

waitTimer := time.NewTimer(time.Duration(timeoutMilli)*time.Millisecond)
QUERY_WAIT:
	for {
			select {
				case <-waitTimer.C:
					//glog.V(1).Info ("query timeout");
					break QUERY_WAIT
				case  <-completedChan:
					//glog.V(1).Info ("query return ")
					break QUERY_WAIT
			}
	}

}

func HandleTagSearch(projectIdHex string,  searchTerm string,  searchPage string, w http.ResponseWriter  ) {
	var pageSize int =100// TODO: this should be configurable per user?
	var pageNum int;
  pageNum, err := strconv.Atoi(searchPage)
  if err!=nil {
  	// something wrong with page num
		server_utils.SendJsonResponseWithStatusCode (w,
  	map[string][]string{"message":[]string{"Page number error"}}, http.StatusNotFound)
  	return

  }

	
	var results = make ([]common.Tag, 0, pageSize)


	var collection *mgo.Collection  = mgoProjectTagsCollection
	// TODO: continue from here

	QueryDbWithTimeout(collection, bson.M{"project_id": bson.ObjectIdHex(projectIdHex),  "tag":bson.M{"$regex":bson.RegEx{"^"+searchTerm,""} } }, bson.M{"tag":1,"_id":1}, &results, pageSize*pageNum, pageSize+1 , ConfigSections.WebServer.DbQueryTimeoutMilliseconds*5)
	


	
	var responseJson map[string]interface{} = make (map[string]interface{})
	
	if len(results)==pageSize+1 {
		responseJson["hasMore"]	= true
		responseJson["results"] = results[:pageSize]
	} else {
		responseJson["hasMore"]	= false
		responseJson["results"] = results
	}
		
	

	
 	server_utils.SendJsonResponse(w,responseJson)


}

// quick popup search for both defs and references
func HandleQuickDefsAndRefsSearch (projectIdHex string, tagIdHex string, searchTerm string,  searchPage string, w http.ResponseWriter  ) {
	
	var pageSize int = 5
	var resultDefs = make ([]common.TagLocation, 0, pageSize)
	var resultRefs = make ([]common.TagLocation, 0, pageSize)
  pageNum, err := strconv.Atoi(searchPage)
  if err!=nil {
  	pageNum  = 0
	}

	var wg sync.WaitGroup
	wg.Add(2)
	
	
	
	go func (wg * sync.WaitGroup) {
		QueryDbWithTimeout(mgoProjectDefinitionsCollection, bson.M{"tag_id": bson.ObjectIdHex(tagIdHex)},bson.M{"_id":0,"path":1, "line_image":1, "line_number":1}, &resultDefs, pageSize*pageNum, pageSize , ConfigSections.WebServer.DbQueryTimeoutMilliseconds*5)
		wg.Done()
	}(&wg)
	
	go func (wg * sync.WaitGroup) {
		QueryDbWithTimeout(mgoProjectReferencesCollection, bson.M{"tag_id": bson.ObjectIdHex(tagIdHex)}, bson.M{"_id":0,"path":1, "line_image":1, "line_number":1}, &resultRefs, pageSize*pageNum, pageSize , ConfigSections.WebServer.DbQueryTimeoutMilliseconds*5)
		wg.Done()
	}(&wg)

	wg.Wait()// wait for all the queries to finish
	var responseJson map[string]interface{} = make (map[string]interface{})


	responseJson["result_defs"] = resultDefs
	responseJson["result_refs"] = resultRefs
	server_utils.SendJsonResponse(w,responseJson)
	
}

// search type should be 'definitions' or 'references'
func HandleDefinitionOrReferenceSearch (projectIdHex string, tagIdHex string, searchType string, searchTerm string, searchPage string, w http.ResponseWriter ){
	var pageSize int =20// TODO: this should be configurable per user?
	var pageNum int;
  pageNum, err := strconv.Atoi(searchPage)
  if err!=nil {
  	// something wrong with page num
		server_utils.SendJsonResponseWithStatusCode (w,
  	map[string][]string{"message":[]string{"Page number error"}}, http.StatusNotFound)
  	return

  }

	
	var results = make ([]common.TagLocation, 0, pageSize)


	var collection *mgo.Collection
	if searchType=="references" {
		collection =mgoProjectReferencesCollection
	} else {
		collection =mgoProjectDefinitionsCollection
	}
	
	
	
	QueryDbWithTimeout(collection, bson.M{"tag_id": bson.ObjectIdHex(tagIdHex)}, bson.M{"_id":0,"path":1, "line_image":1, "line_number":1}, &results, pageSize*pageNum, pageSize+1 , ConfigSections.WebServer.DbQueryTimeoutMilliseconds*5)
	


	
	var responseJson map[string]interface{} = make (map[string]interface{})
	
	if len(results)==pageSize+1 {
		responseJson["hasMore"]	= true
		responseJson["results"] = results[:pageSize]
	} else {
		responseJson["hasMore"]	= false
		responseJson["results"] = results
	}
		
	

	
 	server_utils.SendJsonResponse(w,responseJson)

	
}

// autocomplete of identifier, reference and def search
func HandleGetProjectDetails(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	projectId :=vars["projectId"] // id string in hex
	tokenAutocomplete :=r.FormValue("token-autocomplete")
	
	searchType :=r.FormValue("search_type")
 	searchTerm :=r.FormValue("search_term")
 	searchPage :=r.FormValue("search_page")
 	tagIdHex :=r.FormValue ("tag_id")
 	
	if !VerifyViewPermission(bson.ObjectIdHex(projectId),  w, r) { // not allowed to view ?
		return
	}


	if len(tokenAutocomplete)>0 {
		
		HandleIdentifierAutocomplete(projectId,tokenAutocomplete,w)
		return
		

	}
	
	if searchType =="definitions" || searchType=="references" {
		HandleDefinitionOrReferenceSearch (projectId,tagIdHex, searchType, searchTerm, searchPage, w)
		return
	}
	
	if searchType=="tag" {
		HandleTagSearch (projectId, searchTerm, searchPage, w)
		return
	}
	
	if searchType=="quick_defs_and_refs" {
		HandleQuickDefsAndRefsSearch (projectId, tagIdHex, searchTerm, searchPage, w)
		return
	}
	
	server_utils.SendJsonResponseWithStatusCode (w,
  	map[string][]string{"message":[]string{"No result found"}}, http.StatusNotFound) // HTTP 404

}

func HandleGetProjectTreeSearch(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	projectId :=vars["projectId"] // id string in hex
	searchFor :=r.FormValue("search_for")
	var responseJson map[string]interface{} = make (map[string]interface{})
	results := make ([]string, 0, 50) // TODO: make this length a settings?
	
	
	

	
	resultChan := make (chan string)
	go func (maxNumItems int, resultChan chan string) {
		var query *mgo.Query
		var queryItem common.TreePath
		// option "i" for case-insensitive
		
		// first we try query making use of index
		query =  mgoProjectTreesCollection.Find(
		bson.M{"project_id": bson.ObjectIdHex(projectId), "full_path":bson.M{"$regex":bson.RegEx{"^"+searchFor,"i"} } })
		n, _:=query.Count()
		if n==0 {
			
			// if no result, more broader range
			query =  mgoProjectTreesCollection.Find(
			bson.M{"project_id": bson.ObjectIdHex(projectId), "full_path":bson.M{"$regex":bson.RegEx{".*"+searchFor+".*","i"} } })
			
		}

		iter := query.Limit( maxNumItems ).Iter()
		for iter.Next(&queryItem) {
			resultChan <- queryItem.Path+"/"+queryItem.Name
		
		}
		iter.Close()
		close(resultChan)
		
	}(cap(results), resultChan)

QUERY_WAIT:
	for {
			select {
				case <-time.After ( time.Duration(ConfigSections.WebServer.DbQueryTimeoutMilliseconds)*time.Millisecond):
					//glog.V(1).Info ("query timeout");
					break QUERY_WAIT
				case nextResult, ok :=<-resultChan:
					if (ok){
						results = append (results, nextResult)
					} else {
						//glog.V(1).Info ("result channel is closed")
						break QUERY_WAIT
					}
			}
	}
	
	

	
		
	
		

	if len (results)==0 {

    	server_utils.SendJsonResponseWithStatusCode (w,
        	map[string][]string{"message":[]string{"No result found"}}, http.StatusNotFound) // HTTP 404
		return
	}
	responseJson["results"] =results
	// return
 server_utils.SendJsonResponse(w,responseJson)


}

func HandleGetProjectTree (w http.ResponseWriter, r *http.Request) {
	
	vars := mux.Vars(r)
 	projectId :=vars["projectId"] // id string in hex
	if !common.IsValidObjectIdHex (projectId) {
			server_utils.SendError (w,
			map[string][]string{"message":[]string{"Invalid projectId"}})
			return
	}
		

	if !VerifyViewPermission(bson.ObjectIdHex(projectId),  w, r) { // not allowed to view ?
		return
	}


	var responseJson map[string]interface{} = make (map[string]interface{})
	
	
	if r.FormValue("search_for")!="" {
		//  handle search here
		HandleGetProjectTreeSearch(w, r)
		return
	}
	

	fullPath :=r.FormValue("full_path")
	// handle normal access
	if fullPath=="/" {
		fullPath=""
		responseJson["is_dir"] = true
	} else {

		//  make sure the path exist and also check whether it is file or directory
		baseName :=filepath.Base(fullPath)
		parentPath :=fullPath[:len(fullPath)-len(baseName)-1]
		//glog.V(1).Info ("check if exist name = ", baseName, " with path=",parentPath)
		// query mongo
		var existingDir common.TreePath
  	query :=  mgoProjectTreesCollection.Find(
  	bson.M{"project_id": bson.ObjectIdHex(projectId), "path":parentPath, "name":baseName })
  	
  	n,err := query.Count()
  	if n==0 || err!=nil {
  		// not found, return error
			server_utils.SendError (w,
			map[string][]string{"message":[]string{"Path not found"}})
			return

  	}
  	
  	query.One(&existingDir)
  	
  	// check whether is directory or file
  	responseJson["is_dir"] = existingDir.IsDir

		
	}
	
	
	if responseJson["is_dir"].(bool)  {

		glog.V(1).Info ("directory query  for : path=", fullPath," project :", projectId)
		// if it is directory then query
		iter := mgoProjectTreesCollection.Find (bson.M{"project_id": bson.ObjectIdHex(projectId), "path":fullPath}).Limit(100000).Iter()
		
		
		
		var result []common.TreePath
		
		
		err := iter.All(&result)
		iter.Close()
		if err!= nil {
			glog.Error("There was error querying path :", err)
			server_utils.SendError (w,
			map[string][]string{"message":[]string{"Error querying path"}})
			
			return
		}
		
		responseJson["dir_list"] = &result
		

	} else {
		// TODO: send file content here
		//responseJson["file_content"] = ??
		glog.V(1).Info ("TODO: add content of file here")
	}
	
	
	// return
	 server_utils.SendJsonResponse(w,responseJson)
	


	
	
}

// when user access a long url directly, we have to send back index.html so it can build up the angularjs app structure
func catchAccessViewDirectly(w http.ResponseWriter, r *http.Request) {
	
	indexHtmlPath:=ConfigSections.WebServer.WWWDir+"/index.html"
	glog.V(1).Info ("catch ", r.URL.Path," count = ", strings.Count (r.URL.Path,"/"))
	
	if strings.Count (r.URL.Path,"/")>=4 ||path.Ext(r.URL.Path) == "" { // for example /view/project/5463e4583b7d9c52b3000005/linux-3.18-rc2 or /view/project
		//glog.V(1).Info ("serve index file")
		http.ServeFile(w, r, indexHtmlPath)
		return
	}
	//
	w.WriteHeader(http.StatusNotFound)
  w.Write([]byte("Page not found"))
	
}

/////////
// check for authentication token coming from client
//shouldSend401IfExpired - 	set to true to enforce strick auth access
//													set to false for more relaxed auth check. in this case, the handler function will have to send proper response
func   HandleAuth(f http.HandlerFunc , shouldSend401IfExpired bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		
		authCookie, err := r.Cookie ("auth_token")
		if err!=nil {
			// no cookie  found
			
		}  else {
			
			//glog.V(1).Info ("auth_token cookie : ", authCookie.Value)
			
			token, err :=server_utils.ParseAccessToken(authCookie.Value)
			switch err.(type) {
		 
				case nil: // no error

						
						if userIdHex, ok := token.Claims["id"].(string); ok {
							glog.V(1).Info ("inject context userId from token is ", userIdHex)
							context.Set(r, "user_id", userIdHex)

							
						} else {
							glog.V(1).Info ("type assertion failed ")
						}
						
		
		
				case *jwt.ValidationError: // something was wrong during the validation
					vErr := err.(*jwt.ValidationError)
			 		//println("Error validating token :", vErr.Errors);
					if (vErr.Errors & jwt.ValidationErrorExpired !=0) {
						glog.V(1).Info ( "Token Expired, send unAuthorized .")
						if shouldSend401IfExpired {
							// send back 401
							server_utils.SendJsonResponseWithStatusCode (w,
							map[string][]string{"message":[]string{"Access token expired"}, "reason":[]string{"token_expired"}}, http.StatusUnauthorized)
				return
							return
							
						}
						
					}
					if (vErr.Errors & jwt.ValidationErrorSignatureInvalid !=0) {
						glog.V(1).Info ("ValidationErrorSignatureInvalid")
					}
		
					
				default: // something else went wrong
				//		w.WriteHeader(http.StatusInternalServerError)
					glog.V(1).Info  ("error parsing token 2: ", err)
		
			}

			
			// decode cookie and query user information; before that can also check against the cache first
			// inject user information into request here, using gorilla context
			// http://www.gorillatoolkit.org/pkg/context
			
			
			
			
			
		}
		// still call handler function
		f(w,r);
		
	}
}
//////////////////////////////////////////////////////////////////////////////////////////////

// should only be called when mgoSession is already setup
func initializeDbCollections () {
	
	mgoContactFormCollection 	= mgoSession.DB("Users").C("contact_us") // contact us forms submissions
	mgoSignupUsersCollection	= mgoSession.DB("Users").C("sign_up")	   // those who sign up with email address
	mgoConfirmStringCollection = mgoSession.DB("Users").C("confirm_strings") // _id which is user id and confirm string
	mgoResetStringCollection = mgoSession.DB("Users").C("reset_strings")  //_id, user_id, reset_string
	mgoPrjectsCollection			= mgoSession.DB("Users").C("projects") // project collections
	mgoProjectTreesCollection  = mgoSession.DB("ProjectDetails").C("trees") // materialized paths collection
	mgoProjectDefinitionsCollection = mgoSession.DB("ProjectDetails").C("definitions")
	mgoProjectReferencesCollection = mgoSession.DB("ProjectDetails").C("references")
	mgoProjectTagsCollection= mgoSession.DB("ProjectDetails").C("tags")
}
/////////



///
// to display log to console:
//go run src/webserver/server.go  -logtostderr=true -v=1
// to output log files to a directory:
//go run src/webserver/server.go  -log_dir=/var/log
// for www dir -www_dir=...
func main() {

	var configFiles string
  flag.StringVar(&configFiles, "config", "www.conf,common.conf", "configuration files (comma separated)")
	////
	flag.Parse() // needed for glog
	
	config.ParseConfigFiles (configFiles, &ConfigSections)


	//////

	glog.V(0).Info("starting mongo db session")
	var err error
	
	mgoSession, err = mgo.Dial(ConfigSections.Common.MongoDBIp)
	
	if err != nil {
		panic(err)
	}
	mgoSession.SetMode(mgo.Strong, true)
	mgoSession.SetSocketTimeout(1 * time.Hour) //
	

	initializeDbCollections()
	
	// go func () { // goroutine to periodically ping db and reconnect if necessary
	// 	for {
	// 		time.Sleep (1*time.Second)
	// 		select {
				
	// 			default:
	// 					err := mgoSession.Ping()
	// 					if err!=nil {
	// 						glog.V(1).Info ("Database ping failed. reconnect")
							
	// 						for { // keep trying to connect
	// 							mgoSession, err = mgo.Dial(ConfigSections.Common.MongoDBIp)
								
	// 							if err==nil { // connect ok
	// 								mgoSession.SetMode(mgo.Strong, true)
	// 								mgoSession.SetSocketTimeout(1 * time.Hour) //
	// 								initializeDbCollections()
			
									
	// 								break
	// 							}
	// 							// keep trying to connect again
	// 							glog.V(1).Info ("Reconnect DB failed. try again")
			
	// 							time.Sleep (1*time.Second)
	// 						}
				
			
				
						
	// 					}
				
				
	// 		}

			
	// 	}
	// } ()
	
	

	glog.V(0).Info("Start web server ")
	
	////
	server_utils.ConfigSections = &ConfigSections
	server_utils.InitBackendQueues()
	server_utils.LoadKeysForAccessToken (ConfigSections.WebServer.AccessTokenPrivKeyPath,ConfigSections.WebServer.AccessTokenPubKeyPath)
	

	///////
	
	r := mux.NewRouter()
	
	
	r.HandleFunc("/apiv1/contact-us", 	HandleContactUs).Methods("POST")
  r.HandleFunc("/apiv1/upload", 			HandleAuth(HandleUpload,true)).Methods("POST") //new project via upload method
  r.HandleFunc("/apiv1/project-info/{projectId}", HandleAuth(HandleGetProjectInfo, false)).Methods("GET")
  r.HandleFunc("/apiv1/project-info/{projectId}", HandleAuth(HandleDeleteProjectInfo, true)).Methods("DELETE")
  r.HandleFunc("/apiv1/project-info/{projectId}", HandleAuth(HandleUpdateProjectInfo, true)).Methods("PUT")
  r.HandleFunc("/apiv1/project-info", HandleAuth(HandleNewProjectInfo, true)).Methods("POST") // new project add-via-url method
  r.HandleFunc("/apiv1/manage-projects", HandleAuth(HandleGetListProjects, true)).Methods("GET") // retrieve list of projects


	r.HandleFunc("/apiv1/user", 			HandleSignup).Methods("POST") // new user signup
	r.HandleFunc("/apiv1/user", 			HandleAuth(HandleUserUpdate, false)).Methods("PUT") // various update operations, including confirming new uer and setting password
	r.HandleFunc("/apiv1/user-not-loggedin", 			HandleAuth(HandleUserLoginOrResetPassword, false)).Methods("GET") //login, reset password, confirm email
	r.HandleFunc("/apiv1/user", 			HandleAuth(HandleGetUserProfile, true)).Methods("GET") // retrieve profile information

	
	r.HandleFunc("/apiv1/upload-limit", 			HandleAuth(HandleCheckUploadLimit, true)).Methods("GET")
	
	r.HandleFunc("/apiv1/project-tree/{projectId}", HandleAuth(HandleGetProjectTree, false)).Methods("GET") // directory listing & searching
	r.HandleFunc("/apiv1/project-raw/{projectId}/{path:.*}", HandleAuth(HandleGetProjectRaw , false)).Methods("GET") // get raw content of file
	r.HandleFunc("/apiv1/project-download/{projectId}/{path:.*}", HandleAuth(HandleGetProjectDownload , false)).Methods("GET") // get raw content of file, browser pop up Save as dialog
	r.HandleFunc("/apiv1/project-details/{projectId}", HandleAuth(HandleGetProjectDetails , false)).Methods("GET") // token autocomplete, definitions and references search
	
	
	http.Handle("/apiv1/", r)
	
	rCatch :=mux.NewRouter()
	
	// catch when user go to the url directly

	rCatch.NotFoundHandler =http.HandlerFunc(catchAccessViewDirectly)
	
	http.Handle("/view/", rCatch)

	
	http.Handle("/", http.FileServer(http.Dir(ConfigSections.WebServer.WWWDir)))

	http.ListenAndServe(fmt.Sprintf(":%d", ConfigSections.WebServer.HttpPort), nil)
	

}

