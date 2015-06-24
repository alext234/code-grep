package main

import (

  "github.com/golang/glog"
	zmq "github.com/pebbe/zmq4"
  "flag"
  "config"
  "os"
  "os/exec"
  "encoding/json"
  "fmt"
  "common"
 // "labix.org/v2/mgo"
	// "labix.org/v2/mgo/bson"
	mgo "gopkg.in/mgo.v2"
	bson "gopkg.in/mgo.v2/bson"

	
	"time"
	"strings"
	"github.com/andelf/go-curl" 	// wrapper for libcurl
																// has to install libcurl (libcurl4-gnutls-dev..)

	"strconv"
	"bufio"
	"errors"
	"path/filepath"
	"backend_utils"
	"runtime"
	// for DEBUGGING  - realtime profiling
	_ "net/http/pprof"
	"net/http"
)


// representing an item sent from gtag via zeromq
type GtagUpdateStruct struct {
 UpdateCounter 	int64   							`json:"update_counter"`
 TagType	string											`json:"tag_type"`
 Tag	string													`json:"tag"`
 LineNumber int64											`json:"line_number"`
 LineImage	string										`json:"line_image"`
 Path	string													`json:"path"`
}





var (
	



	mgoSession    								*mgo.Session
	mgoPrjectsCollection					*mgo.Collection // collection of projects, see struct Project in data_structure
	mgoSignupUsersCollection			*mgo.Collection
	mgoProjectDefinitionsCollection *mgo.Collection
	mgoProjectReferencesCollection *mgo.Collection
	mgoProjectTagsCollection 	*mgo.Collection
	
	
	mgoProjectTreesCollection			*mgo.Collection // collection of project materialized paths, see struct TreePath
	

	ConfigSections 	config.BackendWorkderConfig
)

func init () {



}


// process project
func processProjectQueueHandler (receiver * zmq.Socket ) {

	for {
		time.Sleep (time.Second * 1)
		
    msgbytes, err := receiver.RecvBytes(0) // blocked receiving

    if err!=nil {
    	glog.Error ("Error receiving from zmq :", err)
    	continue
    }
    var project common.Project
	  err = json.Unmarshal(msgbytes, &project)
	  if err!= nil {
	    glog.Error ("Failed to unmarshal ", err)
	    continue
	  }


		var user common.SignupUser
	
		
  	err =  mgoSignupUsersCollection.FindId(project.UserId).One(&user)
		if err!=nil  {
			
			glog.Error ("error obtaining user  information :", err, " userid= ",project.UserId.Hex()) // most likely not found
			continue
		} else {
			
			remainingSpace := user.Quota - user.UsedSpace
			if remainingSpace>0 {
				
	  		go processAproject (&project, remainingSpace)
			} else {
				
				updateProjectMessageAndStatus (&project, "You have no space left for the project.","error")
			}
		}
	 



	}


}

func updateProjectMessageAndStatus(project * common.Project, newMessage string, newStatus string) (err error){
	var change mgo.Change

	// update project status
	change = mgo.Change{
    ReturnNew: true,
    Update: bson.M{
        "$set": bson.M{
         	"message": newMessage,
         	"status":newStatus,
        }}}
	if changeInfo, err := mgoPrjectsCollection.FindId(project.Id).Apply(change, project); err != nil {
    glog.Error ("error updating project message   : ", err)
    glog.Error ("changeInfo = ", changeInfo )

	}
	return err
	
}
func updateProjectMessage (project * common.Project, newMessage string) (err error) {
	var change mgo.Change

	// update project status
	change = mgo.Change{
    ReturnNew: true,
    Update: bson.M{
        "$set": bson.M{
         	"message": newMessage,
        }}}
	if changeInfo, err := mgoPrjectsCollection.FindId(project.Id).Apply(change, project); err != nil {
    glog.Error ("error updating project message   : ", err)
    glog.Error ("changeInfo = ", changeInfo )

	}
	return err

}

func updateProjectStatus (project * common.Project, newStatus string) (err error) {
	var change mgo.Change

	// update project status
	change = mgo.Change{
    ReturnNew: true,
    Update: bson.M{
        "$set": bson.M{
         	"status":newStatus,
        }}}
	if changeInfo, err := mgoPrjectsCollection.FindId(project.Id).Apply(change, project); err != nil {
    glog.Error ("error updating project status to db  : ", err)
    glog.Error ("changeInfo = ", changeInfo )

	}
	return err
}


func generateProjectWorkDirName (project * common.Project) {
	project.WorkDir = ConfigSections.Common.DirProjectsWorkDir+"/"+fmt.Sprintf ("%d", time.Now().UnixNano()) +"_"+project.Name+"_workdir"

	// also update to mongo
	var change mgo.Change

	// update project status
	change = mgo.Change{
    ReturnNew: true,
    Update: bson.M{
        "$set": bson.M{
         	"work_dir": project.WorkDir,
        }}}
	if changeInfo, err := mgoPrjectsCollection.FindId(project.Id).Apply(change, project); err != nil {
    glog.Error ("error updating project workdir   : ", err)
    glog.Error ("changeInfo = ", changeInfo )

	}
}


func getBallExtension (path string)  string {
	if strings.HasSuffix(path,".tar") {
		return "tar"
	} else if strings.HasSuffix(path,".tar.gz") {
		return "tar.gz"
	} else if strings.HasSuffix(path,".tar.bz2") {
		return "tar.bz2"
	} else if strings.HasSuffix(path,".tar.xz") {
		return "tar.xz"
	} else if strings.HasSuffix(path,".zip") {
		return "zip"
	}
	return "unknown"

}

// extract from sourcePath into destDir
func doExtraction (sourcePath string, destDir string, cancelChannel chan bool) (err error) {
	glog.V(1).Info ("start extracting from ", sourcePath," into ", destDir)
	var execCmd * exec.Cmd
	extension := getBallExtension(sourcePath)


	if extension =="unknown" {
		return errors.New ("Unknown file type")
	}
	
	if extension =="zip" {
		execCmd = exec.Command("unzip","-q", sourcePath,"-d", destDir)
	} else {
		execCmd = exec.Command("tar","xf", sourcePath,"-C", destDir)
	}
	
	glog.V(1).Info ("exec :",  execCmd.Args)
	err = execCmd.Start()
	if err != nil {
		glog.Error ("Error running extract :", err)
    return errors.New ("Extraction failed")
    
	}
	
	execDoneC := make (chan error)
	go func () {
		err := execCmd.Wait()
		glog.V(1).Info("extract cmd finished : ", err)
		execDoneC <- err
	}()
	
	glog.V(1).Info ("timeout for extraction :", ConfigSections.BackendWorker.MaxExtractTimeout)

	defer func() {
		glog.V(1).Info ("return from extraction")
	}()
	
	waitTimer := time.NewTimer(time.Duration(ConfigSections.BackendWorker.MaxExtractTimeout*int64(time.Second)))
	
	for {
		select {
			case <- waitTimer.C : // someting must have gone wrong, either git takes so long or process hang somewhere
				// process has to be killed
				execCmd.Process.Kill()
				return errors.New ("Extraction took too long to complete")
				
			case err=<-execDoneC:
				return err
			default:
				runtime.Gosched()
		}
	}

	return nil
}


// check extracted size of the ball
// return -1 if cancelled or some error occurs
func getBallSize (path string, cancelChannel chan bool) int64 {
	var execCmd * exec.Cmd
	extension := getBallExtension(path)


	if extension =="unknown" {
		return 0
	}
	if extension =="zip" {
		execCmd = exec.Command("zipinfo","-h", path)
	} else {
		execCmd = exec.Command("tar","tvf", path)
	}
	glog.V(1).Info ("exec :",  execCmd.Args)
	resultC := make (chan int64)
	stopReadingStdoutC :=make (chan bool)
	stdout, err := execCmd.StdoutPipe()
	if err != nil {
		glog.Error ("Failed to connec to stdout pipe  :", err)
		return 0
	}
	
	execCmd.Start()
	// start another goroutine to processs stdout
	go func () {
		r := bufio.NewReader(stdout)
		var size int64 = 0
		for {
			select   {
				case <- stopReadingStdoutC:
					glog.V(1).Info ("stopReadingStdout")
					resultC <- 0
					return
				default:
					lineTemp, _, err := r.ReadLine();
					if err!=nil {
						glog.V(1).Info ("line = ", lineTemp)
						glog.V(1).Info ("err readline ", err)

						resultC <- size
						return
					}
					line:= strings.ToLower (string(lineTemp))
					// process line here
					switch (extension) {
						case "zip":
						// "zip file size: 12961889 bytes, number of entries: 4129"
						
						if strings.HasPrefix (line, "zip file size:") {
					
							
						
							n, err :=fmt.Sscanf (line, "zip file size: %d bytes", &size)
							if err != nil {
								glog.V(1).Info ("Error scanning for size information in :", line)
								glog.V(1).Info ("error = ", err)
								size = -1
							} else if n==0 {
								glog.V(1).Info ("Error - zero elements in scanf  :", line)
								glog.V(1).Info ("size = ", size)
								size = -1
							}
							resultC <- size
							return
						}
						
						
						default: // other types than zip
							var tmpSize int64
							var s1, s2 string
							n, err :=fmt.Sscanf (line, "%s %s %d", &s1, &s2, &tmpSize)
							if (err == nil ) && (n>0) {
									size = size + tmpSize
							}
							
						
					}
					
			}
			///???
		}
		
		
	} ()
	
	var result int64 = -1
	waitTimer := time.NewTimer(60*time.Second)
WAITING_FOR_SIZE:
	for {
		select {
			
			case <- cancelChannel :
				glog.V(1).Info ("User cancel. kill process")
				stopReadingStdoutC<-true
				execCmd.Process.Kill()
				return -2
				
			case <- waitTimer.C: // someting must have gone wrong,
				// process has to be killed
				glog.V(1).Info ("timeout waiting for size result. have to kill the process")
				execCmd.Process.Kill()
				return 0
				
	
				
				
			case result= <-resultC:
				
				break WAITING_FOR_SIZE
			default:
				runtime.Gosched()
				
		}
	}
	return result
}


// scan and update to mongodb
func scanProjectTree (project * common.Project,  cancelChannel  chan bool) {

	// call sync first before walk
	var execCmd * exec.Cmd = exec.Command("sync")
	err := execCmd.Start()
	if err!=nil {
		glog.V(1).Info ("Failed to run sync")
	}
	time.Sleep(2*time.Second)
	
	//
	project.TotalSize =0 // to store size On Disk
	glog.V(1).Info ("Start walkdir :", project.WorkDir)
	

	//itemsList  := make ([]common.TreePath, 0, 1000000)
	itemsList  := make ([]interface{}, 0, 1000000)
	
	rootLen := len (project.WorkDir)
	isCancel :=false
	err = filepath.Walk (project.WorkDir,
		func (path string, info os.FileInfo, err error) error {//callback function
			//glog.V(1).Info ("walk callback is called")
			select {
				case <-cancelChannel:
					isCancel=true
					return errors.New ("User canceled ")
					
				default:
					if err!=nil {
						glog.V(1).Info ("Tree walk err :", err)
						return err
					}
					// last element
					baseName :=info.Name()
					
					if info.IsDir() && baseName==".git" {
						return filepath.SkipDir // skip version control directories
					}
					
					endSlice := len(path)-len(baseName)
					if rootLen >endSlice { // the base directory itself
						return nil
					}
					

					// add to list  which will be later inserted into mongo
					var treePath common.TreePath
					treePath.Id =bson.NewObjectId();
					
					treePath.Name = baseName
					treePath.ProjectId = project.Id
					treePath.Path = path[rootLen:endSlice-1]
					treePath.FullPath  = treePath.Path+"/"+treePath.Name
					treePath.IsDir = info.IsDir()
					if info.IsDir() {
						treePath.Size=0
					} else {
						treePath.Size = info.Size()
					}
					
					itemsList =append (itemsList, treePath)
					//glog.V(1).Info ("add :", path, "  ", treePath.Path," ", baseName)
					//glog.V(1).Info ("walk  :"+treePath.FullPath)
					
					
					
					

					
					
					project.TotalSize = project.TotalSize+info.Size()
					
				
					
			}
			return nil
		})

	
	if err!=nil {
		if isCancel {
			updateProjectMessageAndStatus(project, "Project was cancelled", "error")
		} else {
			glog.V(1).Info ("Error scanning project dir :", err)
			updateProjectMessageAndStatus(project, "There was error scanning the source tree", "error")
		}

		return
	}
	
	// insert the whole list to mongo - bulk insert?
	glog.V(1).Info ("dirwalk completed. insert  list into db")
  

	
	err1 := mgoProjectTreesCollection.Insert (itemsList...)
	

	if err1!=nil {
		glog.Error ("Error insert list of treewalk items   into db :", err1)
		updateProjectMessageAndStatus(project, "There was error updating directory tree items to database", "error")
		return
	}


	glog.V(1).Info ("total size: ", project.TotalSize)
	
	// update project size, project status and message
	change := mgo.Change{
    ReturnNew: true,
    Update: bson.M{
        "$set": bson.M{
         	"message": "Analyzing files for definitions and references",
         	"status":"analyzing",
         	"total_size":project.TotalSize,
        }}}
	if _, err := mgoPrjectsCollection.FindId(project.Id).Apply(change, project); err != nil {
    glog.Error ("error updating project message   : ", err)
    //glog.Error ("changeInfo = ", changeInfo )
		
	}

	
	// ensureIndex on project_id and fullpath
	go func() {
		index := mgo.Index{
    	Key: []string{"project_id", "full_path"},
    	Unique: true,
    	Background: true,  // see explanation here http://godoc.org/gopkg.in/mgo.v2#Query.Apply
		}
		err := mgoProjectTreesCollection.EnsureIndex(index)
		if err!=nil {
			glog.Error ("Error ensure index for projectTrees :", err)
		}
	}()
		

	
	return
}


// scan files for definitions and references of tokens/symbols
func analyzeProjectTree (project * common.Project,  cancelChannel  chan bool) {



	zmqUrl :="tcp://"+ConfigSections.Common.ZmqProxyHost+":"+ConfigSections.Common.ZmqGeneralPubPort // connect string to be used by external git program
	subscriptionFilter := "\""+project.Id.Hex()+":identifier:" + "\""// also zmq_msg_prefix to be used by external program



	// // subscribe to zmq pub coming from external gtag program
	subSocket, _ := zmq.NewSocket(zmq.SUB)
	connectStr  :="tcp://"+ConfigSections.Common.ZmqProxyHost+":"+ConfigSections.Common.ZmqGeneralSubPort
	
	err:= subSocket.Connect(connectStr)
	
	if err != nil {
		glog.Error("Error connecting to zmq socket :", err)
		return
	}
	
	glog.V(1).Info ("subscription filter str: ", subscriptionFilter)
	subSocket.SetSubscribe (subscriptionFilter)

	defer subSocket.Close()
	closeSubscriber := make(chan bool)

	isErrorDuringGtag := false;
	isGtagTakesTooLong :=false
	isGtagCanceled :=false
	isGtagDone :=false // only set to true when received <<done>> message
	
	
	
	
	// this gorotuine also receive PUB from gtag program and add to
	go func(receiver * zmq.Socket) { // go routine- which is a subscriber to project commands comming from web server frontend
		statusMsg :="Analyzing files for definitions and references"
		updateProjectStatusRate := time.NewTicker(100 * time.Millisecond)



		for {

			select {
				case <- closeSubscriber: // no more checking for command
					
					return
				case <-updateProjectStatusRate.C:
					updateProjectMessage(project,statusMsg)
					
				default:
					//poll non-blocking for message from gtag program
					gtagFullMsg, _ := receiver.Recv(zmq.DONTWAIT)
					if len(gtagFullMsg)>0 {
						gtagMsgTemp :=gtagFullMsg [len(subscriptionFilter):]
						switch (gtagMsgTemp) {
							case "<<done>>":
								isGtagDone=true
								glog.V(1).Info ("gtag done successfully")
							case "<<error>>":
								
								isErrorDuringGtag= true
							default:
								statusMsg =gtagMsgTemp
								
								
								
								
					 	}

					}
					runtime.Gosched()
			}
		}
	} (subSocket)

	// call external gtag program

	 cmd := exec.Command(ConfigSections.BackendWorker.GtagPath,   "-z", zmqUrl , "-s", subscriptionFilter,"-m","mongodb://"+ConfigSections.Common.MongoDBIp, "-b", "ProjectDetails","-p", project.Id.Hex())
	 cmd.Dir = project.WorkDir //
	glog.V(1).Info ("exec :",  cmd.Args, " in directory :", cmd.Dir)
	
	err = cmd.Start()
	if err != nil {
    glog.Error ("Failed to run gtags :", err)
    updateProjectMessageAndStatus(project, "There was error while preparing to analyzing the project", "error")
  	closeSubscriber <-true

    return
	}

	execDoneC := make (chan error)
	
	// goroutine which simply waits for external gtag to terminate
	go func () {
		err := cmd.Wait() // see sample for wait with timeout https://github.com/golang-samples/exec/blob/master/wait/main.go
		glog.V(1).Info("gtags finished with error: ", err)
		execDoneC <- err
		
	}()




	glog.V(1).Info ("timeout for gtag :", ConfigSections.BackendWorker.MaxGtagTimeout)
	waitTimer := time.NewTimer(time.Duration(ConfigSections.BackendWorker.MaxGtagTimeout*int64(time.Second)))
WAITING_GTAG:
  

	for {
			select {

				case <- waitTimer.C: // someting must have gone wrong, either git takes so long or process hang somewhere
					// process has to be killed
					isGtagTakesTooLong=true

					cmd.Process.Kill()
		
					break WAITING_GTAG
			
				case <-cancelChannel:
					glog.V(1).Info ("Gtag was cancelled!")
					isGtagCanceled= true
					cmd.Process.Kill()
					break WAITING_GTAG
			
				case <-execDoneC:
					glog.V(1).Info ("Exec process exits by itself")
					break WAITING_GTAG
				default:
					runtime.Gosched()

			}
	}
	

	
	
	time.Sleep (2000*time.Millisecond) // hopefully sufficient time to receive gtagDone message
	closeSubscriber <-true
	

	if (isGtagTakesTooLong) {
		
		updateProjectMessage(project, "Project analysis takes too long to complete");
		updateProjectStatus (project,"error")
	}  else if (isGtagCanceled) {
		
		
		updateProjectMessageAndStatus(project, "Analyzing was cancelled. Cleanup in progress...", "error")
		
	//}	else if (!isErrorDuringGtag && !isGtagCanceled && isGtagDone) {
	}	else if (!isErrorDuringGtag && !isGtagCanceled ) { // TODO: temporarily skip isGtagDone at the moment because it seems we miss <<done>> under heavy load
		
		
		
		updateProjectMessageAndStatus(project, "Project processing is done. Ready for browsing!", "ready")


		
	} else {
		
		glog.V(1).Info ("isErrorDuringGtag=", isErrorDuringGtag," isGtagCanceled=", isGtagCanceled, " isGtagDone=", isGtagDone)
		updateProjectMessageAndStatus(project, "There was error analzying the project files. Clean up in progress...", "error")
		
	}
	glog.V(1).Info ("TODO: should clean up the files GPATH, GRTAGS, and GTAGS !!!!")
	




	
	return
}

// involve several steps: check for size, and then proceed to do extraction
func extractProjectBall (project * common.Project,maxDownloadSize int64,  cancelChannel  chan bool) {
	// current extensions supported .tar.xz, .tar.bz2 .tar.gz .tar .zip
	fullBallPath := ConfigSections.Common.DirProjectsUploaded+"/"+project.BallName
	
	step := "check_size" // and then "extract" and then return
	for {
		select {
			case <- cancelChannel : // user-initiated cancellation
				
				glog.V(1).Info ("user cancel uncompressing process")
				updateProjectMessageAndStatus(project, "Project was cancelled. Cleanup in progress...", "error")
				return
			default:
				time.Sleep (1000*time.Millisecond)
				switch (step) {
					case "check_size":
						glog.V(1).Info ("checking for  uncompressed size of ball : ", fullBallPath)
						ballSize := getBallSize (fullBallPath, cancelChannel)
						glog.V(1).Info ("ball size :", ballSize)
						glog.V(1).Info ("max size :", maxDownloadSize)
						if ballSize==0{
							// some error happen
							updateProjectMessageAndStatus(project, "Error occurred while extracting the project", "error")
	
							return
							
						} else if ballSize ==-2 {
							// user-initiated cancellation
							updateProjectMessageAndStatus(project, "Project was cancelled", "error")
							return
							
						}
						if (ballSize > maxDownloadSize ) {
							// TODO update status and message
							updateProjectMessageAndStatus(project, "Extraction was cancelled due to size exceeding limit", "error")
							return
						} else {
							//
							// ballSize looks good, move on to extract state
							step  = "extract"
						}
						
					case "extract":
						glog.V(1).Info ("start actual extraction here ")
						
						 
						generateProjectWorkDirName(project) // generate workdir  - for extraction
						err:=os.Mkdir (project.WorkDir, 0766)
						if err!=nil {
							glog.Error ("Failed to create workdir ",project.WorkDir)
							updateProjectMessageAndStatus(project, "Internal server error: Failed to create working directory", "error")
					
							return
						}
						
						err=doExtraction (fullBallPath, project.WorkDir, cancelChannel)
						if err!= nil {
							
							updateProjectMessageAndStatus(project, fmt.Sprintf("Error during extraction process: %s", err), "error")

							return
							
						} else {
							step  = "done"
						}
						
					case "done":
						glog.V(1).Info ("Extraction is done")
						updateProjectMessageAndStatus(project, "Scanning directories and files...", "tree_scanning")

						return
					default:
						runtime.Gosched()
				}
		
		}
	}

	



}

func gitCloneProject (project * common.Project,maxDownloadSize int64,  cancelChannel  <-chan bool){

	
	//////////////////////////////////////////
	
	generateProjectWorkDirName(project)
	glog.V(1).Info ("git clone dir :", project.WorkDir)
	glog.V(1).Info ("git url :", project.FetchUrl)
	glog.V(1).Info ("gi2 path : ", ConfigSections.BackendWorker.Git2Path)
	glog.V(1).Info ("maximum size allowed :", maxDownloadSize)
	


	
	
	zmqUrl :="tcp://"+ConfigSections.Common.ZmqProxyHost+":"+ConfigSections.Common.ZmqGeneralPubPort // connect string to be used by external git program
	subscriptionFilter := "\""+project.Id.Hex()+":git:" + "\""// also zmq_msg_prefix to be used by external git program



	// // subscribe to zmq pub coming from external git program
	subSocket, _ := zmq.NewSocket(zmq.SUB)
	connectStr  :="tcp://"+ConfigSections.Common.ZmqProxyHost+":"+ConfigSections.Common.ZmqGeneralSubPort
	
	err:= subSocket.Connect(connectStr)
	
	if err != nil {
		glog.Error("Error connecting to zmq socket :", err)
		return
	}
	glog.V(1).Info ("subscription filter str: ", subscriptionFilter)
	subSocket.SetSubscribe (subscriptionFilter)

	defer subSocket.Close()
	closeSubscriber := make(chan bool)

	isErrorDuringGit := false
	isGitTakesTooLong :=false
	isProjectCancel :=false
	
	go func(receiver * zmq.Socket) { // go routine- which is a subscriber to project commands comming from web server frontend
		var gitMsg string
		updateDbRate := time.NewTicker(100 * time.Millisecond)

		for {

			select {
				case <- updateDbRate.C:  // update rate to mongodb
					updateProjectMessage (project, gitMsg)
				case <- closeSubscriber: // no more checking for command
					updateDbRate.Stop()
					return
				
				default:
					//poll non-blocking
					gitFullMsg, _ := receiver.Recv(zmq.DONTWAIT)
					if len(gitFullMsg)>0 {
						gitMsgTemp :=gitFullMsg [len(subscriptionFilter):]
						switch (gitMsgTemp) {
							case "<<done>>":
								
								glog.V(1).Info ("git done successfully")
							case "<<error>>":
								
								isErrorDuringGit= true
							default:
								gitMsg = gitMsgTemp
						}

					}
			}
		}
	} (subSocket)

	// call external git program
	cmd := exec.Command(ConfigSections.BackendWorker.Git2Path,"clone", project.FetchUrl, project.WorkDir,zmqUrl , subscriptionFilter, strconv. FormatInt(maxDownloadSize, 10))
	glog.V(1).Info ("exec :",  cmd.Args)
	
	// debugging code if needing to capture stdout
	// stdout, err := cmd.StdoutPipe()
	// if err != nil {
	// 	glog.Error ("Failed to connec to stdout pipe  :", err)
	// 	return
	// }
	err = cmd.Start()
	if err != nil {
    glog.Error ("Failed to run git2 :", err)
    updateProjectMessageAndStatus(project, "There was error cloning the project. Cleanup in progress...", "error")
    
    return
	}

	execDoneC := make (chan error)
	go func () {
		err := cmd.Wait() // see sample for wait with timeout https://github.com/golang-samples/exec/blob/master/wait/main.go
		glog.V(1).Info("git2 finished with error: ", err)
		execDoneC <- err
	}()
	
	glog.V(1).Info ("timeout for git :", ConfigSections.BackendWorker.MaxGitCloneTimeout)
	
	waitTimer := time.NewTimer(time.Duration(ConfigSections.BackendWorker.MaxGitCloneTimeout*int64(time.Second)))
WAIT_GIT:
	
	for {
		select {
				case <- cancelChannel:
					glog.V(1).Info ("Project is cancelled")
					isProjectCancel=true
					break WAIT_GIT

				case <- waitTimer.C : // someting must have gone wrong, either git takes so long or process hang somewhere
					// process has to be killed
					cmd.Process.Kill()
		
					isGitTakesTooLong=true
					break WAIT_GIT
					
				case <-execDoneC:
					glog.V(1).Info ("Exec process exits by itself")
					break WAIT_GIT
				default:
					runtime.Gosched()

			}
		
	}
	
	closeSubscriber <-true
	time.Sleep (500*time.Millisecond)

	if (isProjectCancel) {
		updateProjectMessageAndStatus(project, "Project is cancelled. Cleanup in progress...", "error")
	} else if (isGitTakesTooLong) {
		
		updateProjectMessage(project, "Git cloning was stopped as it took too long to complete.");
		updateProjectStatus (project,"error")
	} else if (!isErrorDuringGit) {
		updateProjectMessageAndStatus(project, "Scanning directories and files...", "tree_scanning")
	} else {
		updateProjectStatus (project,"error")
	}
	

}

func downloadProjectViaCurl (project * common.Project,maxDownloadSize int64,  cancelChannel  <-chan bool) {
	fullSavedPath := ConfigSections.Common.DirProjectsUploaded+"/"+project.BallName
	glog.V(1).Info ("savedPath  =", fullSavedPath)
	fp, err := os.OpenFile(fullSavedPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err!=nil {
		glog.Error ("failed to open file to write :", fullSavedPath)
		return
	}
	defer fp.Close() // defer close


	curl.GlobalInit(curl.GLOBAL_ALL)

	// init the curl session
	easy := curl.EasyInit()
	defer easy.Cleanup()

	easy.Setopt(curl.OPT_URL, project.FetchUrl)
	easy.Setopt(curl.OPT_WRITEDATA, fp)



	easy.Setopt(curl.OPT_WRITEFUNCTION, func (ptr []byte, userdata interface{}) bool {
			file := userdata.(*os.File)
			if _, err := file.Write(ptr); err != nil {
				glog.Error ("failed to write to file :", err)
				return false
			}
			return true
			
		
			
	})


	easy.Setopt(curl.OPT_LOW_SPEED_LIMIT, 500); // if speed is only 500Bytes/sec for 30 seconds
	easy.Setopt(curl.OPT_LOW_SPEED_TIME, 30); // then should abort it
	
	easy.Setopt(curl.OPT_MAX_RECV_SPEED_LARGE, 2000000); // limit download speed
	//easy.Setopt(curl.OPT_MAX_RECV_SPEED_LARGE, 20000); // DEBUGGING: limit to very low to test out cancel function

	easy.Setopt(curl.OPT_NOPROGRESS, false)
	easy.Setopt(curl.OPT_TIMEOUT,ConfigSections.BackendWorker.MaxDownloadTimeout);
	var sizeExceeded bool = false
	var dlPercent int64=0
	var isProjectCancel bool  = false
	easy.Setopt(curl.OPT_PROGRESSFUNCTION, func(dltotal, dlnow, ultotal, ulnow float64, userdata interface{}) bool {
		
		select {
			case <- cancelChannel:
				isProjectCancel=true
				return false
			default:


				break
		}
		
		
		if (int64(dltotal) >=maxDownloadSize) {
			sizeExceeded = true
			return false// exceed size allowed
		}
		
		if (dltotal>0.0) {
			newPercent  := int64 (dlnow/dltotal*100)
			if (newPercent > dlPercent) {
				dlPercent = newPercent
				updateProjectMessage (project, fmt.Sprintf("Downloaded %d%%", dlPercent))
				

			}
			
		}
		
		return true
	})

	if err := easy.Perform(); err != nil {
		if (sizeExceeded) {
			updateProjectMessage (project, fmt.Sprintf("Downloading was cancelled due to size exceeding limit"))
			updateProjectStatus (project,"error")
		} else if (isProjectCancel) {

			updateProjectMessage (project, fmt.Sprintf("Project was cancelled. Clean up in progress.."))
			updateProjectStatus (project,"error")

		} else {
			
			glog.Errorf("Error downloading :", err)
			updateProjectMessage (project, fmt.Sprintf("There was error downloading the file"))
			updateProjectStatus (project,"error")

		}
	} else {
		
		updateProjectMessage (project, fmt.Sprintf("Start extracting..."))
		time.Sleep(1* time.Second)
		updateProjectStatus (project,"extracting")
		time.Sleep(1* time.Second)
	}

}
////////////////
// a worker processing a project
func processAproject (project * common.Project,  maxDownloadSize int64) {

	// // subscribe to project commands that might come from frontend web server
	subSocket, _ := zmq.NewSocket(zmq.SUB)
	connectStr  :="tcp://"+ConfigSections.Common.ZmqProxyHost+":"+ConfigSections.Common.ZmqGeneralSubPort
	
	glog.V(1).Info ("Subscribe to project  :", connectStr)
	err:= subSocket.Connect(connectStr)
	
	if err != nil {
		glog.Error("Error connecting to zmq socket :", err)
		return
	}
	subscriptionFilter := project.Id.Hex()+":command:"
	glog.V(1).Info ("subscription filter str: ", subscriptionFilter)
	subSocket.SetSubscribe (subscriptionFilter)

	
	closeSubscriber := make(chan bool)
	waitForSubScriberGoRoutineExit :=make (chan bool)
	projectCancel := make(chan bool)
	go func(receiver * zmq.Socket) { // go routine- which is a subscriber to project commands comming from web server frontend
		for {

			select {
				case <- closeSubscriber: // no more checking for command
					glog.V(1).Info ("closeSubscriber")
					waitForSubScriberGoRoutineExit <- true
					return
				default:
					//poll non-blocking
					cmdMsg, _ := receiver.Recv(zmq.DONTWAIT)
					if len(cmdMsg)>0 {
						cmd :=cmdMsg [len(subscriptionFilter):]
						glog.V(1).Info("Receive cmd :", cmd)
						if cmd=="cancel" {
							projectCancel<-true
						}
					}
			}
		}
	} (subSocket)
	defer func () {
		glog.V(1).Info ("quit go routine subcribing to project commands")
		
		close(closeSubscriber)
		<-waitForSubScriberGoRoutineExit
		
		close(projectCancel)
		
		subSocket.Close()
		glog.V(1).Info ("return from project processing")
		
	}()
	////////////////////////////////////////////
	

	glog.V(1).Info ("process project :", project.Id.Hex())
	for {
		select {
			case <-projectCancel:
				glog.V(1).Info ("project processing is cancelled")
				updateProjectMessage (project, fmt.Sprintf("Project processing is cancelled. Cleanup in progress..."))
				updateProjectStatus (project,"error")

				
				return
			default:
				switch (project.Status) {
					case "error":
						glog.V(1).Info ("Project is in error state, return ")
						return
					case "ready":
						glog.V(1).Info ("Project is in ready state, return ")
						return
					case "tree_scanning":
						glog.V(1).Info("Start scanning source tree", project.WorkDir)
						scanProjectTree (project, projectCancel) // scan and update to mongodb

					case "analyzing":
						glog.V(1).Info ("start analyzing for definitions and references")
						analyzeProjectTree (project, projectCancel) // parse files for definitions and references
						
					case "extracting":
						glog.V(1).Info ("extracting :",project.BallName)
						extractProjectBall (project,maxDownloadSize, projectCancel)
						
					case "uploaded":
						if err:=updateProjectStatus(project, "extracting"); err!=nil {
							return // some error occurs while updating project id
						}
			
			   

					case "url_received":
						if err:=updateProjectStatus(project, "fetching"); err!=nil {
							
							return // some error occurs while updating project id
						}
						
			
					
						
						glog.V(1).Info ("start fetching :", project.Id," ", project.FetchUrl)
						if strings.HasSuffix(project.FetchUrl,".git") {
							
							gitCloneProject (project, maxDownloadSize, projectCancel)
						} else {
			
							// download normal url using libcurl
							downloadProjectViaCurl (project, maxDownloadSize , projectCancel)
			
			
						}
					
				}
		}

		
	}
	

}


///
func cleanupDeletedProjects()  {
	for {
			time.Sleep (1*time.Minute)
			
			select {

				default:
				  //glog.V(1).Info ("look for  project to cleanup")
					// look for the next "error" project - deleted or
					var project common.Project
					query :=  mgoPrjectsCollection.Find(bson.M{"status":"error"})
					query.Limit(1)// one at a time
					err:=query.One (&project)
					
					if err!=nil {
						continue
					}
					glog.V(1).Info ("Start cleanup ", project.Id)
				
					
			

					// remove the project
					if  err :=mgoPrjectsCollection.RemoveId(project.Id); err != nil {
						
						glog.Error ("failed to remove project from project collections")
						continue // maybe another worker already removed it
					}
					
					// remove from the user project list
					change := mgo.Change{
					    ReturnNew: true,
				  	  Update: bson.M{
				        "$pull": bson.M{
				         	"project_list": project.Id,
				        }}}
					var user common.SignupUser
					_, err =  mgoSignupUsersCollection.FindId(project.UserId).Apply(change,&user)
					if err!=nil {
						glog.Error ("failed to remove project from user project list")
					}

					
					// remove materialized path from db
					err = mgoProjectTreesCollection.Remove(bson.M{"project_id":project.Id})
					
					// iterate through tags and delete references/definitions
					
					tagIter := mgoProjectTagsCollection.Find (bson.M{"project_id":project.Id}).Iter()
					var item common.Tag
					for tagIter.Next(&item) {
						// remove references from db
						err = mgoProjectReferencesCollection.Remove(bson.M{"tag_id":item.Id})
						
						// remove definitions from db
						err = mgoProjectDefinitionsCollection.Remove(bson.M{"tag_id":item.Id})
						
						
					}
					
					// remove  the tag itself
					mgoProjectTagsCollection.Remove (bson.M{"project_id":project.Id})
					
					// remove ballpath - if there is
					
					if project.BallName!="" {
						fullBallPath := ConfigSections.Common.DirProjectsUploaded+"/"+project.BallName
						glog.V(1).Info ("Remove :" , fullBallPath)
						err = os.RemoveAll(fullBallPath)
						if err!=nil {
							glog.Error ("Error RemoveAll  :", err)
						}
					}
					
					// remove work directory
					
					if strings.HasPrefix (project.WorkDir, ConfigSections.Common.DirProjectsWorkDir+"/") {
						glog.V(1).Info ("Remove :" , project.WorkDir)
						err = os.RemoveAll(project.WorkDir)
						if err!=nil {
							glog.Error ("Error RemoveAll  :", err)
						}
					}
					
					
					glog.V(1).Info ("completed cleanup of project ",project.Id)
					// ensureindex of project trees and references and definitions
					go func() {
						index := mgo.Index{
				    	Key: []string{"project_id", "full_path"},
				    	Unique: true,
				    	Background: true,  // see explanation here http://godoc.org/gopkg.in/mgo.v2#Query.Apply
						}
						err := mgoProjectTreesCollection.EnsureIndex(index)
						if err!=nil {
							glog.Error ("Error ensure index for projectTrees :", err)
						}
							
							
						
						
					}()


					
					
				
			}
	}

}


func initializeDbCollections(){
	mgoPrjectsCollection			= mgoSession.DB("Users").C("projects") // project collections

	mgoProjectTreesCollection  = mgoSession.DB("ProjectDetails").C("trees") // materialized paths collection
	
	mgoProjectDefinitionsCollection = mgoSession.DB("ProjectDetails").C("definitions")
	mgoProjectReferencesCollection = mgoSession.DB("ProjectDetails").C("references")
	mgoProjectTagsCollection= mgoSession.DB("ProjectDetails").C("tags")
	
	mgoSignupUsersCollection	= mgoSession.DB("Users").C("sign_up")	   // those who sign up with email address
	
}
////////////////////////////////////////////////////////
////////////////////////////////////////////////////////
// initialize and start workers in seperate go routines
func main () {
  hostname,err :=os.Hostname()
  if err!=nil {
    glog.Error ("Failed to get hostname")
    return
  }

	var configFiles string
  flag.StringVar(&configFiles, "config", "backend.conf,common.conf", "configuration files (comma separated)")
	////
	flag.Parse() // needed for glog
  glog.Info ("back end worker started. Hostname =  ", hostname)
	
	config.ParseConfigFiles (configFiles, &ConfigSections)
	glog.V(1).Infof ("Config information %+v", ConfigSections)

	backend_utils.ConfigSections =&ConfigSections
	/////



	// database connections
	glog.V(0).Info("starting mongo db session")
	mgoSession, err = mgo.Dial(ConfigSections.Common.MongoDBIp)
	if err != nil {
		panic(err)
	}
	defer mgoSession.Close()
	mgoSession.SetSocketTimeout(1 * time.Hour) //
	mgoSession.SetMode(mgo.Strong, true)

	
	initializeDbCollections()
	// go func () { // goroutine to periodically ping db and reconnect if necessary
	// 	for {
	// 		time.Sleep (1*time.Second)

	// 		select {
				
	// 			default:
	// 				err := mgoSession.Ping()
	// 				if err!=nil {
	// 					glog.V(1).Info ("Database ping failed. reconnect")
						
	// 					for { // keep trying to connect
	// 						mgoSession, err = mgo.Dial(ConfigSections.Common.MongoDBIp)
							
	// 						if err==nil { // connect ok
	// 							mgoSession.SetMode(mgo.Strong, true)
	// 							mgoSession.SetSocketTimeout(1 * time.Hour) //
	// 							initializeDbCollections()
		
								
	// 							break
	// 						}
	// 						// keep trying to connect again
	// 						glog.V(1).Info ("Reconnect DB failed. try again")
		
	// 						time.Sleep (1*time.Second)
	// 					}
					
	// 				}
				
	// 		}

			
	// 	}
	// } ()
	

	////////////
  //
  sendConfirmEmailQueueReceiver, _ := zmq.NewSocket(zmq.PULL)
  defer sendConfirmEmailQueueReceiver.Close()
  
	emailQueueConnectStr :="tcp://"+ConfigSections.Common.ZmqProxyHost+":"+ConfigSections.Common.ZmqConfirmEmailPullPort
	glog.V(1).Info ("Email queue zmq :" +emailQueueConnectStr)
  sendConfirmEmailQueueReceiver.Connect(emailQueueConnectStr)
  
	
	/////
  sendResetEmailQueueReceiver, _ := zmq.NewSocket(zmq.PULL)
  defer sendResetEmailQueueReceiver.Close()
  
	resetQueueConnectStr :="tcp://"+ConfigSections.Common.ZmqProxyHost+":"+ConfigSections.Common.ZmqResetEmailPullPort
	glog.V(1).Info ("Email queue zmq :" +resetQueueConnectStr)
  sendResetEmailQueueReceiver.Connect(resetQueueConnectStr)

	
	////
  projectToBeProcessedQueueReceiver, _ := zmq.NewSocket(zmq.PULL)
  defer projectToBeProcessedQueueReceiver.Close()
  
  projectQueueConnectStr :="tcp://"+ConfigSections.Common.ZmqProxyHost+":"+ConfigSections.Common.ZmqProjectToBeProcessedPullPort
  glog.V(1).Info ("Project queue zmq :" +projectQueueConnectStr)
  projectToBeProcessedQueueReceiver.Connect(projectQueueConnectStr)

  



  // start all go routines to process work queues
  ///////////////
  
  go backend_utils.SendConfirmEmailQueueHandler(sendConfirmEmailQueueReceiver)
  go backend_utils.SendResetEmailQueueHandler(sendResetEmailQueueReceiver)
  go processProjectQueueHandler (projectToBeProcessedQueueReceiver)
	
	// the clean up worker - go and clean deleted projects
	go  cleanupDeletedProjects()

	// realtime profiling for DEBUGGING
	go func() {
		http.ListenAndServe(":6060", nil)
	}()
  select{
  }
}
