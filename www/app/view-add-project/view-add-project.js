'use strict';




angular.module('cgApp.viewAddProject', ['ui.bootstrap','angularFileUpload'])




.controller ('addProjectCtrl', ['LoggedinUser', '$scope', '$modalInstance', '$window', 'FileUploader','Settings', '$location','$route', 'SharedData','$timeout','ProjectInfo', 'UploadLimit', function (LoggedinUser, $scope, $modalInstance, $window, FileUploader, settings, $location, $route,SharedData, $timeout, ProjectInfo,UploadLimit) {
  
	$scope.getProfileData = LoggedinUser.getProfileData // its a function


  var uploader = $scope.uploader = new FileUploader({
      url: settings.apiBase+'/upload',
      queueLimit: 1
  })

	function endsWith(str, suffix) {
	  return str.indexOf(suffix, str.length - suffix.length) !== -1;
	}

  uploader.filters.push({name:'filterBeforeAddToQueue', fn:function(item) {
    uploader.clearQueue();// only 1 item in queue
    // file extension checking here, if does not match return false
    // .tar, .tar.bz2, .tar.gz, .tar.xz
    
    if (endsWith(item.name, ".tar") || endsWith(item.name, ".tar.gz") || endsWith(item.name, ".tar.bz2")
    || endsWith(item.name, ".tar.xz")||endsWith(item.name, ".zip")) {

    	return true;
    }
  	$scope.uploadLimitCheckMessage="File extension should be .tar, .tar.xz, .tar.gz, .tar.bz2 or .zip";
		$scope.uploadLimitCheckMessageClass="text-danger";

    return false;// extension not accepted
  }});


  uploader.onAfterAddingFile = function(fileItem) {
      // here we can also check against allocated size for the user and inform
      // fileItem.file.size
      //console.log (fileItem.file.size);
      if (fileItem.file.size>$scope.uploadLimitBytes) {
      	$scope.uploadLimitCheckMessage="File size exceeds limit";
	  		$scope.uploadLimitCheckMessageClass="text-danger";
				uploader.clearQueue();
      }  else {
      	$scope.uploadLimitCheckMessage="";
	  		$scope.uploadLimitCheckMessageClass="";
      	
      }
      
      
      
  };
  $scope.isGoingToRedirect=false;
  $scope.uploadViewSuccesssMsg = "";
  uploader.onSuccessItem = function(fileItem, response, status, headers) {
    
    $scope.uploadViewSuccesssMsg = response.message; // json response from server containing project_id
    
		
   	$scope.isGoingToRedirect=true;
		
    setTimeout (function () {
    	$modalInstance.dismiss('success');
    	$scope.$apply(function () {

		    	var newUrl = "/view/project/"+response.project_id
		    	$location.url (newUrl)



  		  });
    		
    	}, 2000);
    
		
    
  };
  $scope.uploadViewErrorMsg="";
  uploader.onErrorItem = function(fileItem, response, status, headers) {
    $scope.uploadViewErrorMsg = "There was error uploading the file";
    
    fileItem.isUploaded = false;// trick  so we can upload again
  };


  $scope.fileNameChanged = function (elem) {
    if (elem.files.length==0) { // when user click to open the file selection dialog but then he click Cancel button
      //console.log("User does not select file anymore. clearqueue");
      uploader.clearQueue();
    }
  };
  
  $scope.isShowUploadView = false;
  $scope.isShowAddUrlView = false;
  
  $scope.ok = function () {
    
    
    if ($scope.selectAddVia=="upload") {
	    $scope.isShowUploadView=true;
	    setTimeout (function () {
	      
	
	      $scope.uploadViewErrorMsg="";
	      uploader.uploadAll();
	    }, 100);
    	
    } else if ($scope.selectAddVia=="url"){
    	//use ProjectInfo resource save
    	$scope.isShowAddUrlView = true;
    	$scope.addUrlViewSuccessMsg="";
    	$scope.addUrlViewErrorMsg = "";
    	
    	$scope.saveProjectInfoPromise=  $timeout (function (){
	    	 ProjectInfo.save ({fetch_url:$scope.urlInput}, function (response, responseHeaders){
	    		if (response.message) {
	    			$scope.addUrlViewSuccessMsg=response.message;
	    		}
	    		if (response.project_id) {
	    			 $scope.isGoingToRedirect=true;
		
    					$timeout (function () {
    						$modalInstance.dismiss('success');
    					
					    	var newUrl = "/view/project/"+response.project_id
					    	$location.url (newUrl)

    	
    					
    					}, 2000);
    		
    	

	    		} else {
	    			
	    			// ?? no project_id ??? server responds with something different??
	    		}
	    		
	    		
	    		
	    	}, function (httpResponse) {
	    		//console.log (httpResponse);
	    		if (httpResponse.data.errors) {
	    			$scope.addUrlViewErrorMsg=httpResponse.data.errors.message[0];
	    		} else {
	    			$scope.addUrlViewErrorMsg = "There was error communicating with the server. Please try again later";
	    		}
	    	})
    		
    	}, 600);
    	
    }
    
    
  
  };
  
  
	$scope.goToView = function (viewLink) {
		
		$timeout(function () {
		  $modalInstance.dismiss('cancel'),
			$location.$$search = {}; //clear all search
			$location.hash("");
			$location.path (viewLink, false);
			$window.location.reload();

			
		}, 100)
	}
	

  $scope.cancel = function () {
    if ($scope.isShowUploadView) {

      $scope.isShowUploadView=false;
      uploader.cancelAll();
    } else if ($scope.isShowAddUrlView) {
    	$scope.isShowAddUrlView= false;
    	$timeout.cancel ($scope.saveProjectInfoPromise); // cancel the pending query
    }
    else {
    	SharedData.addProjectDialogViaMethod = $scope.selectAddVia;
      setTimeout (
        $modalInstance.dismiss('cancel'),
        200
      );
    }
  };
  
  
  
  $scope.$on('$routeChangeStart', function () {
  		// when leaving this view, store the current selected dialog option to global data
  		SharedData.addProjectDialogViaMethod = $scope.selectAddVia;

  })
  
  $timeout (function () {
  	
  	$scope.isAddViaUploadActive = (SharedData.addProjectDialogViaMethod=="upload");
  	$scope.isAddViaUrlActive =  (SharedData.addProjectDialogViaMethod=="url");
  
  	
  },10);
  
  $scope.selectAddViaUrl = function (){
  
    $scope.selectAddVia="url";
    
  }
  
  $scope.isShowFileSelection = false;
  
  $scope.uploadLimitBytes=5000000; // some small value; will be checked everytime the upload dialog is open
  
  $scope.selectAddViaUpload = function (){
  	$scope.isShowFileSelection = false;
  	$scope.uploadLimitCheckMessage = "Checking for allowable upload size...";
 		$scope.uploadLimitCheckMessageClass="";
    $scope.selectAddVia="upload";
    $timeout (function () {
	    UploadLimit.get ({}, function (response, responseHeaders){
				
				if (response.upload_limit!=null) {
					$scope.uploadLimitBytes =response.upload_limit;
					if ( $scope.uploadLimitBytes==0) {
						$scope.uploadLimitCheckMessage="You have no space left. Please try to delete some projects to free up space."
					} else {
							$timeout (function () {
								$scope.uploadLimitCheckMessage="";
								$scope.isShowFileSelection = true;
							}, 500);

					}
					
				}
				
				
	  	}, function (httpResponse) {
	  		//console.log (httpResponse);
	    		if (httpResponse.data.errors) {
	    			$scope.uploadLimitCheckMessage=httpResponse.data.errors.message[0];
	    		} else {
	    			$scope.uploadLimitCheckMessage = "There was error communicating with the server. Please try again later";
	    		}
	  		
	  		$scope.uploadLimitCheckMessageClass="text-danger";
	  	})
    	
    }, 100);

    
  }
  
  $scope.isUrlInputValid = false;
  var validUrlExp = /(http|https|git):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?/;


	
  $scope.checkUrlInput = function (urlInput) {
  	$scope.urlInput = urlInput;
		$scope.isUrlInputValid= validUrlExp.test ($scope.urlInput);
		return true;
  }
  
}])




;
