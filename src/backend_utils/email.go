package backend_utils

import (

  "github.com/sendgrid/sendgrid-go"
  "fmt"
  "runtime"
  zmq "github.com/pebbe/zmq4"
  "github.com/golang/glog"
  "time"
  "encoding/json"
)


//send confirmation email to signup users
func SendConfirmEmailQueueHandler (receiver * zmq.Socket ) {
  var err error

  sg :=sendgrid.NewSendGridClient(ConfigSections.BackendWorker.SendGridUser,ConfigSections.BackendWorker.SendGridPassword)

  for { // loop forever;
  	
		time.Sleep (time.Second * 1)
		select {
			
			default:
		    msgbytes, _ := receiver.RecvBytes(0)

		    

				var data  map[string]interface{} = make (map[string]interface{})
		
			
		    //err = json.Unmarshal(msgbytes, &signupUser)
		    err = json.Unmarshal(msgbytes, &data)
		    if err!= nil {
		      glog.Error ("Failed to unmarshal ", err)
		      continue
		    }
		    
		    var confirmString string
		    var userIdHex string
		    var email string
				userIdHex = data["id"].(string)
				confirmString = data["confirm_string"].(string)
				email  = data["email"].(string)
				
		    
		    message := sendgrid.NewMail()
		    message.AddTo(email)
		    message.SetSubject("Account confirmation")
		
		    var confirmUrl string
		    var urlPort string
		
		
		    urlPort =""
		
		    confirmUrl = fmt.Sprintf ("https://%s%s/view/edit-profile?user_id=%s&confirm_string=%s",
		    ConfigSections.Common.BaseUrl,
		    urlPort,
		    userIdHex,
		    confirmString)
		
		    text :=`
		    
Hi,

Please click on the link below to activate your account and to set your password:

` + confirmUrl+
`


Thank you,

Code/Grep - Code Browsing Made Easy


		    `

		    message.SetText(text)

		    message.SetFrom("Code/Grep <contact@code-grep.com>")

		    if err= sg.Send(message); err == nil {

		    } else {
		        glog.Error ("Error sending confirmation email : ", err)
		    }

		}
    runtime.Gosched() // avoid hogging cpu
  }
}



//send reset email to  users
func SendResetEmailQueueHandler (receiver * zmq.Socket ) {
  var err error

  sg :=sendgrid.NewSendGridClient(ConfigSections.BackendWorker.SendGridUser,ConfigSections.BackendWorker.SendGridPassword)

  for { // loop forever;
  	
		time.Sleep (time.Second * 1)
		select {
			
			default:
		    msgbytes, _ := receiver.RecvBytes(0)

		    

				var data  map[string]interface{} = make (map[string]interface{})
		
			
		    //err = json.Unmarshal(msgbytes, &signupUser)
		    err = json.Unmarshal(msgbytes, &data)
		    if err!= nil {
		      glog.Error ("Failed to unmarshal ", err)
		      continue
		    }
		    
		    var resetString string
		    var userIdHex string
		    var email string
				userIdHex = data["id"].(string)
				resetString = data["reset_string"].(string)
				email  = data["email"].(string)
				
		    
		    message := sendgrid.NewMail()
		    message.AddTo(email)
		    message.SetSubject("Password reset")
		
		    var resetUrl string
		    var urlPort string
		
		
		    urlPort =""
		
		    resetUrl = fmt.Sprintf ("https://%s%s/view/edit-profile?user_id=%s&reset_string=%s",
		    ConfigSections.Common.BaseUrl,
		    urlPort,
		    userIdHex,
		    resetString)
		
		    text :=`
		    
Welcome to Code/Grep,

Please click on the link below to reset your password:

` + resetUrl+
`


Thank you,

Code/Grep - Code Browsing Made Easy


		    `

		    message.SetText(text)

		    message.SetFrom("Code/Grep <contact@code-grep.com>")

		    if err= sg.Send(message); err == nil {

		    } else {
		        glog.Error ("Error sending confirmation email : ", err)
		    }

		}
    runtime.Gosched() // avoid hogging cpu
  }
}

