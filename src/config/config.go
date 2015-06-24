package config

import (
	"runtime"
	"github.com/golang/glog"
	"strings"
	"code.google.com/p/gcfg"

)

type CommonConfig struct {
	Common struct {
		BaseUrl string
		DirProjectsUploaded string
		DirProjectsWorkDir	string
		
		MyIp		string
		MongoDBIp	string
		MaxUploadLimit  int64
		NewSignupQuota	int64
		
		ZmqProxyHost string
		ZmqConfirmEmailPushPort string
		ZmqConfirmEmailPullPort string

		ZmqResetEmailPushPort string
		ZmqResetEmailPullPort string

  
		ZmqProjectToBeProcessedPushPort string
		ZmqProjectToBeProcessedPullPort string
		
		ZmqGeneralPubPort string
		ZmqGeneralSubPort string
		
		ZmqGeneralReqPort string
		ZmqGeneralRepPort string
		


	}
}


type WebServerConfig struct {

	CommonConfig
	WebServer struct {

		HttpPort  int64
		AccessTokenPrivKeyPath string
		AccessTokenPubKeyPath string
		AccessTokenExpMinutes int64
		WWWDir	string
		
		DbQueryTimeoutMilliseconds int64
	}
}



type MsgProxyConfig struct {

	CommonConfig
	
}


type BackendWorkderConfig struct {
	CommonConfig
	BackendWorker struct {
		SendGridUser string
		SendGridPassword string
		MaxDownloadTimeout int64
		MaxGitCloneTimeout int64
		MaxExtractTimeout int64
		MaxGtagTimeout int64
		Git2Path string
		GtagPath string
		

	}
}

func init()  {
	runtime.GOMAXPROCS(runtime.NumCPU())

}




func 	ParseConfigFiles (configFiles string, ConfigSections interface{}) {
	fileList :=strings.Split (configFiles, ",")
	if len (fileList)==0 {
		panic ("No configuration files specified!!")
		
	}
	glog.V(0).Info ("config files :", fileList)
	for _, filename :=range fileList {
		err := gcfg.ReadFileInto(ConfigSections, filename)
		if err!=nil {
			panic ("Failed to read config file "+filename)
		}

	}
	glog.V(1).Infof ("Config information %+v", ConfigSections)

}
