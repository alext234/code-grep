// run various message broker services such as queues for processing confirmation email
// or queue for processing new projects
package main

import (

  "github.com/golang/glog"
	zmq "github.com/pebbe/zmq4"
  "flag"
  "config"
)
  
var (
	ConfigSections config.MsgProxyConfig
	)

// see explanation here http://api.zeromq.org/4-0:zmq-proxy#toc2
func RunStreamerProxy (pushPort string, pullPort string ) {
	var err error
	
	glog.V(1).Info ("push port :", pushPort, " pull port :", pullPort)
	//  Socket facing web servers (front end)
	
	frontend, _ := zmq.NewSocket(zmq.PULL)
	defer frontend.Close()
	err = frontend.Bind("tcp://*:"+pushPort)
	if err != nil {
		glog.Error("Binding frontend:", err)
	}

	//  Socket facing backend workers

	backend, _ := zmq.NewSocket(zmq.PUSH)
	defer backend.Close()
	err = backend.Bind("tcp://*:"+pullPort)
	if err != nil {
		glog.Error("Binding backend:", err)
	}


	//  Start the proxy
	err = zmq.Proxy(frontend, backend, nil)
	
	
	glog.Error("Proxy interrupted:", err)
	

}


// see explanation here http://api.zeromq.org/4-0:zmq-proxy#toc2
func RunForwarderProxy (pubPort string, subPort string ) {
	var err error
	
	glog.V(1).Info ("pub port :", pubPort, " sub port :", subPort)
	//  Socket facing clients
	
	frontend, _ := zmq.NewSocket(zmq.XSUB)
	defer frontend.Close()
	err = frontend.Bind("tcp://*:"+pubPort)
	if err != nil {
		glog.Error("Error Binding frontend:", err)
	}
 

	//  Socket facing services

	backend, _ := zmq.NewSocket(zmq.XPUB)
	defer backend.Close()
	err = backend.Bind("tcp://*:"+subPort)
	if err != nil {
		glog.Error("Error Binding backend:", err)
	}


	//  Start the proxy
	err = zmq.Proxy(frontend, backend, nil)
	
	
	glog.Error("Proxy interrupted:", err)
	

}




func main () {
	var configFiles string
  flag.StringVar(&configFiles, "config", "common.conf", "configuration files (comma separated)")
	////
	flag.Parse() // needed for glog
	
	config.ParseConfigFiles (configFiles, &ConfigSections)


	

	

	// see here for the terms http://api.zeromq.org/3-2:zmq-proxy
	go RunStreamerProxy (ConfigSections.Common.ZmqConfirmEmailPushPort, ConfigSections.Common.ZmqConfirmEmailPullPort)
	go RunStreamerProxy (ConfigSections.Common.ZmqResetEmailPushPort, ConfigSections.Common.ZmqResetEmailPullPort)
	go RunStreamerProxy (ConfigSections.Common.ZmqProjectToBeProcessedPushPort, ConfigSections.Common.ZmqProjectToBeProcessedPullPort)
	
	go RunForwarderProxy (ConfigSections.Common.ZmqGeneralPubPort, ConfigSections.Common.ZmqGeneralSubPort)
	
	
	select {}
}