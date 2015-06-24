package server_utils

import (

  "config"
  "common"
  zmq "github.com/pebbe/zmq4"
  "github.com/golang/glog"
  "encoding/json"
  
)

// message queues communication to backend
var (
	sendConfirmEmailQueueSender *zmq.Socket
	sendResetEmailQueueSender *zmq.Socket
	projectToBeProcessedQueueSender *zmq.Socket
	ConfigSections					*config.WebServerConfig
	pubSocket								*zmq.Socket // for sending commands to backend workers
)



func init () {

}

func InitBackendQueues () {
  glog.V(1).Info ("mq backend init")
	// add monitoring

  //////
  sendConfirmEmailQueueSender, _ = zmq.NewSocket(zmq.PUSH)

	var err error
  err = sendConfirmEmailQueueSender.Connect("tcp://"+ConfigSections.Common.ZmqProxyHost+":"+ConfigSections.Common.ZmqConfirmEmailPushPort)
	

	if err!=nil {
		println ("Error connecting to zmq socket:", err);
	}
	
	///////////
	sendResetEmailQueueSender, _ = zmq.NewSocket(zmq.PUSH)


  err = sendResetEmailQueueSender.Connect("tcp://"+ConfigSections.Common.ZmqProxyHost+":"+ConfigSections.Common.ZmqResetEmailPushPort)
	

	if err!=nil {
		println ("Error connecting to zmq socket:", err);
	}
	
	/////////////////
  projectToBeProcessedQueueSender, _ = zmq.NewSocket(zmq.PUSH)

	connectProjectQueueStr :="tcp://"+ConfigSections.Common.ZmqProxyHost+":"+ConfigSections.Common.ZmqProjectToBeProcessedPushPort
	glog.V(1).Info ("connect str to project queue :", connectProjectQueueStr)
  err = projectToBeProcessedQueueSender.Connect(connectProjectQueueStr)
  
	if err!=nil {
			println ("Error connecting to zmq socket:", err);
		}

	// socket to pubsub proxy
	pubSocket, _ = zmq.NewSocket(zmq.PUB)
	connectStr:="tcp://"+ConfigSections.Common.ZmqProxyHost+":"+ConfigSections.Common.ZmqGeneralPubPort
	pubSocket.SetLinger (1)
	glog.V(1).Info ("Connect to pub proxy string: ", connectStr)
	err= pubSocket.Connect(connectStr)
	if err != nil {
		glog.Error("Error connecting to zmq socket :", err)
		return
	}


}

func AddToQueue (sender * zmq.Socket, data interface{}) {

  s, err  := json.Marshal (data)
  if err!=nil {
    glog.Error ("failed to serialize  object when adding to  queue ", err, "     data = ", data)
    return
  }
  // uncomment this line to see the whole json serialization of data - for debugging purpose
  glog.V(1).Info (string(s))
  
  _, err =sender.SendBytes (s, zmq.DONTWAIT);

  if err!=nil {
  	glog.Error ("Error sending to ZMQ :",err)
  }
	glog.V(1).Info ("Add to queue done")
}

// add to task queue for backend workers to handler
func AddToSendConfirmEmailQueue (user * common.SignupUser, confirmString string) {
	glog.V(1).Info ("Add to confirm email queue")
	var data  map[string]interface{} = make (map[string]interface{})
	data ["id"] = user.Id.Hex()
	data ["confirm_string"] = confirmString
	data ["email"]  = user.Email
	

	AddToQueue(sendConfirmEmailQueueSender, data)


}



// add to task queue for backend workers to handler
func AddToSendResetEmailQueue (user * common.SignupUser, resetString string) {
	glog.V(1).Info ("Add to reset password email queue")
	var data  map[string]interface{} = make (map[string]interface{})
	data ["id"] = user.Id.Hex()
	data ["reset_string"] = resetString
	data ["email"]  = user.Email
	

	AddToQueue(sendResetEmailQueueSender, data)


}

// add new project to queue of projects to be processed
func  AddToProjectToBeProcessedQueue (proj *common.Project) {
	glog.V(1).Info ("Add to projects to be processed queue")
	AddToQueue(projectToBeProcessedQueueSender, proj)
	
}


// send directly to the worker that's processing the project
func SendCommandToProjectWorkder (projectIdHex string, command string) {
	fullPubStr :=projectIdHex+":command:"+command
	glog.V(1).Info ("command : ", fullPubStr)
	pubSocket.Send(  fullPubStr,zmq.DONTWAIT)


}