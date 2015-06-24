'use strict';

angular.module('cgApp.viewProject', ['ngRoute','hljs','ab-base64'])

.directive('focusMe', function($timeout, $parse) {
  return {
    //scope: true,   // optionally create a child scope
    link: function(scope, element, attrs) {
      var model = $parse(attrs.focusMe);
      var listenerDeregister = scope.$watch(model, function(value) {
        
        if(value === true) {
        	
          $timeout(function() {
            element[0].focus();
            element[0].select();
          });
          // unwatch
          listenerDeregister();
        }
      });
      
    }
  };
})



// affix directive will make the view stick to the top
.directive( 'affix', [ '$window', '$document', function ( $window, $document ) {
  return {
    scope: { offset: '@' },
    link: function ( scope, element, attrs ) {
      var win = angular.element ( $window ),
          affixed;
      
      // Obviously, whenever a scroll occurs, we need to check and possibly
      // adjust the position of the affixed element.
      win.bind( 'scroll', checkPosition );
      
      // Less obviously, when a link is clicked (in theory changing the current
      // scroll position), we need to check and possibly adjsut the position. We,
      // however, can't do this instantly as the page may not be in the right
      // position yet.
      win.bind( 'click', function () {
        setTimeout( checkPosition, 1 );
      });

      scope.$watch('sourceCode', function() {
        checkPosition();
      });

      
      // This is mostly a direct port of the jQuery plugin. It checks to see if
      // we need to adjust the position of the affixed element and, if so, does
      // so.
      function checkPosition() {
        var scrollHeight = $document.prop( 'height' ),
            scrollTop = win.prop("pageYOffset"),
            positionTop = element.prop( 'offsetTop' ),
            reset = 'affix',
            affix;
        
        // Calculate the affix state. This is either `top` indicating we are at
        // the top and do not need to affix anything or `false` to indicate that
        // the affixing should take place.
        // TODO: support the `offset` attribute as a function
        // TODO: support affixing to bottom
        if ( scrollTop <= scope.offset ) affix = "top"
        else affix = false;
        
        // If scrolling hasn't changed the satus of our affixing, why do aything
        // else? Instead of repeating what's already been done, we just return here.
        if (affixed === affix) return;
        
        // Update the affix status to whatever we just calculated.
        affixed = affix;
        
        // First, we remove any applicable classes and then add both the `affix`
        // class as well as whatever specialized class we need, if any.
        element.removeClass( reset ).addClass( 'affix' + (affix ? '-' + affix : '') );
      }
    }
  };
}])

.config(['$routeProvider','$locationProvider', function($routeProvider,$locationProvider) {
	
	$routeProvider
	.when('/view/project/:projectId/', {
    templateUrl: 'view-project/view-project.html',
    controller: 'viewProjectCtrl',
    reloadOnSearch: false,
  })
  .when('/view/project/:projectId/:sourcePath*', {
    templateUrl: 'view-project/view-project.html',
    controller: 'viewProjectCtrl',
    reloadOnSearch: false,
  })
  ;


  
	

}])



.config(['hljsServiceProvider',function (hljsServiceProvider) {
  //console.log ("configure for hljs");
  hljsServiceProvider.setOptions({
    tabReplace:'    ',
    lineNodes: true // show line numbers
  });
}])


.factory('SourceFilesCache', [
         '$cacheFactory',
function ($cacheFactory) {
  return $cacheFactory('SourceFilesCache');
}])



.controller('viewProjectCtrl', ['$q','$rootScope','$route','$scope', '$http', '$window','base64','$routeParams' ,
  '$location','$anchorScroll','SourceFilesCache', 'ProjectInfo', 'ProjectTree', 'ProjectDetails', 'SharedData','$timeout','Settings','$document','$sce',
  function($q,$rootScope,$route,$scope, $http, $window,base64,$routeParams,$location,$anchorScroll,SourceFilesCache, ProjectInfo,ProjectTree,ProjectDetails,  SharedData,$timeout,settings,$document, $sce) {


	$scope.apiBase = settings.apiBase

  $scope.projectInfo = null
  
	$scope.errorMessage='';
	$scope.processingMessage = '';
	$scope.periodicQueryProjectInfo = null;
	$scope.sourceCode = null
	function obtainProjectInfo   (projectId) {
    	
    		var isChangedInStatus = true
    		$scope.projectInfoPromise  = ProjectInfo.get ({projectId:projectId},
	    		function (projectInfo, responseHeaders) {// rest api good response
	    			
			    	if (projectInfo.id!=$routeParams.projectId) {
			    		//console.log ("not the same project id anymore. better do nothing now ");

			    		return;
			    	}
			    	
			    	
			    	if ($scope.projectInfo) {
			    		if ($scope.projectInfo.status  == projectInfo.status  ) {
			    			isChangedInStatus=false
			    		}
			    	}
			    	$scope.projectInfo= projectInfo
			    	
			    	SharedData.currentProject.data = projectInfo
			    	
			    	//check for status and display message from projectInfo
			    	if (projectInfo.status=="error") {
			    		$scope.errorMessage = projectInfo.message;
			    		$scope.processingMessage= '';
			    	} else if (projectInfo.status !="ready" ) {
			    		$scope.processingMessage = projectInfo.message;
			    		
			    		
			    		// periodically check again
			    		$scope.periodicQueryProjectInfo = $timeout(function () {
			    			console.log ("obtain prject info ")
			    			obtainProjectInfo(projectId)
			    		},2000);
			    		//console.log ("should call here again");
			    	}
			    	
			    	/// only when project ok then check for the exact file or directory listing
			    	if (projectInfo.status=="ready" || projectInfo.status=="analyzing") {

			    		$scope.processingMessage='';
			    		$scope.errorMessage='';

				    	var sourcePath;
				    	
				    	
				    	if ($routeParams.sourcePath)  {
				    		
				    		sourcePath = "/"+$routeParams.sourcePath;
				    		console.log ("sourcePath = "+sourcePath)
				    	} else {
				    		sourcePath = '/';
				   
				    	}


							
							
			    		var searchItems = $location.search()
			    		if (Object.keys(searchItems).length!=0 ) {
	  			    		updateCurrentPath (sourcePath);
	  			    		if (isChangedInStatus) { // project status has changed
	  			    			console.log ("call handleSearchTerms 2")
	  			    			handleSearchTerms();
	  			    		}
			    		} else {
				    		if (isChangedInStatus) { // project status has changed
									console.log ("call obtainProjectSourcePath 5")
						    	obtainProjectSourcePath (sourcePath, false);
				    			
				    		}
					    	
			    			
			    		}
			    		
	
		
		    		}
		    		if (projectInfo.status=="analyzing"){
		    			$scope.analyzingMsg = projectInfo.message.substring (0,44)+"...";
		    			
		    		} else {
		    			$scope.analyzingMsg="";
		    		}
	
		    	},
			    function (httpResponse) { // error getting project info
	    			//console.log (httpResponse)

			    	// if exist errors array
			    	if (httpResponse.data.errors) {
			    		$scope.errorMessage = httpResponse.data.errors.message[0];
			    		$scope.processingMessage= '';
			    	} else {
			    		$scope.errorMessage = "There was an error obtaining project information. Please try again later.";
			    		$scope.processingMessage= '';
			    	}
	    	
			    })
  	
  } // obtainProjectInfo
  
  
  
  
  
  // user click 'Prev'  on search result page
  $scope.goToPrevSearchPage= function () {
  	var currentPage = parseInt($scope.searchPage)
  	if (currentPage==0) return;
  	
  	$scope.isSkipQueryingTag=true;
		$location.search({
				search: $scope.searchTerm,
				type: $scope.searchType,
				page:currentPage-1
		})
  	
  }
    
 
 
 	// user click 'Next'  on search result page
  $scope.goToNextSearchPage = function () {
  		$scope.isSkipQueryingTag=true;
  		// skip querying tag
			$location.search({
				search: $scope.searchTerm,
				type: $scope.searchType,
				page:parseInt($scope.searchPage)+1
		})

  }
  
  
  $scope.searchMessage='';
  $scope.searchResults = null
  function handleSearchTerms() {
  	var queryItems = $location.search()
  	$scope.searchTerm = ''
  	$scope.searchType = ''
  	$scope.searchPage = 0 // pagination
  	$scope.searchResults = null
  	
  	$scope.searchTerm = queryItems['search'].trim()
  	
  	if (queryItems['type']=='definitions') {
  		
  		$scope.searchType = 'definitions'
  		$scope.searchMessage = "Searching for definitions of '" + $scope.searchTerm + "'" ;
  	}
  	else if (queryItems['type']=='references') {
  		
  		$scope.searchType = 'references'
  		$scope.searchMessage = "Searching for references of '" + $scope.searchTerm + "'";
  
  	}
  	else if (queryItems['type']=='grep') {
  		
  		$scope.searchType = 'grep'
  		$scope.searchMessage = "Greping for '" + $scope.searchTerm + "'";
  
  	}
  	if (queryItems['page']) {
  		$scope.searchPage = queryItems['page']
  	}
  	if ($scope.searchType=='') return;
  	
  	// cancel any previous search if there is
  	if ($scope.searchTimer) {
  		$timeout.cancel ($scope.searchTimer);
  	}
  	
  	$scope.isShowListOfTags = false
  	// use timeout method so we can cancel http://stackoverflow.com/questions/24944772/angularjs-cancel-previous-resource-promises-if-there-is-a-new-request
  	
  	if ($scope.isSkipQueryingTag) {
  		
  		
  		$scope.isSkipQueryingTag = false
  		$scope.handleSearchWithTagId  ($scope.previousSelectedTagId,$scope.previousSelectedTag)
  		return
  	}
  	
  	
  	// search for tag first a
  	$scope.searchForTagAndPeformOp($scope.searchTerm ,
  		function (response, responseHeaders) { // search for tag success
			  		$scope.listOfTags = response.results
			  		$scope.hasMoreTags = response.hasMore
			  		
			  		if ($scope.listOfTags.length==0){
			  			$scope.searchResults=[]
			  			
	   					$scope.processingMessage = '';
	  			 		$scope.searchMessage = "";
			  			
			  		} else if ($scope.listOfTags.length==1)
			  		{
			  			$scope.handleSearchWithTagId ($scope.listOfTags[0].id, $scope.listOfTags[0].tag)
			  		} else {
	   					$scope.processingMessage = '';
	  			 		$scope.searchMessage = "";
			  				
			  			$scope.isShowListOfTags = true // user will have to select
			  			
			  		}
			  		
			  	
  			
  		},
  		function (httpResponse) { // search for tag error
			  		$scope.processingMessage = '';
			  		
			  		if (httpResponse.data.errors) {
			  			$scope.searchMessage = httpResponse.data.errors.message[0];
			  		}
			  		else {
			  			$scope.searchMessage='There was error trying to query search result from server. Please try again later';
			  			
			  		}
  			
  		}
  	)


  }
  
  $scope.searchForTagAndPeformOp = function (searchTerm, successFunc, errorFunc) {
  	$scope.searchTimer = $timeout(function (){
  			
  			
		 		ProjectDetails.get({
			  		projectId: $scope.projectInfo.id,
			  		search_term: searchTerm,
			  		search_type: 'tag', // first search for the tag
			  		search_page:0, // note: for tag search, only show first page at the moment
			  
			  	},
			  
			  	function(response, responseHeaders) { // no error
			  		
						successFunc(response, responseHeaders)
			  	
			  	
			  	},
			  
			  	function(httpResponse) { // there is error
						errorFunc(httpResponse)
			  	}
			  	)
		  	}
  	
  		
  	, 10);
  	
  }
  
  $scope.handleSearchWithTagId = function (tag_id, tag) {
  	$scope.isShowListOfTags = false
  	$scope.searchTerm=tag
  	
  	$scope.previousSelectedTagId=tag_id
  	$scope.previousSelectedTag=tag
  	
  	if ($scope.searchType=='definitions') {
  		
  		
  		$scope.searchMessage = "Searching for definitions of '" + tag + "'" ;
  	}
  	else if ($scope.searchType=='references') {
  	
  	
  		$scope.searchMessage = "Searching for references of '" + tag + "'";
  	}
  	
  		$scope.searchTimer = $timeout (function() {
  		ProjectDetails.get({
	  		projectId: $scope.projectInfo.id,
	  		tag_id: tag_id,
	  		search_term: $scope.searchTerm,
	  		search_type: $scope.searchType,
	  		search_page: $scope.searchPage
	  
	  	},
	  
	  	function(response, responseHeaders) { // no error
	  		$scope.processingMessage = '';
	  		$scope.searchMessage = "";
	  		
	  		if ($scope.searchType=='definitions') {
	  			$scope.searchResultString="Definitions of "
	  		} else if ($scope.searchType=='references') {
	  			$scope.searchResultString="References of "
	  		} else  if ($scope.searchType=='grep') {
	  			$scope.searchResultString="Grep results for "
	  		} else {
	  			$scope.searchResultString=""
	  		}
	  		$scope.searchResults =response.results
	  		if (response.hasMore){
	  			$scope.searchHasMore = true
	  		} else {
	  			$scope.searchHasMore = false
	  		}
	  		console.log ($scope.searchResults)
	  	
	  		
	  		
	  		
	  	},
	  
	  	function(httpResponse) { // there is error
	  		$scope.processingMessage = '';
	  		
	  		if (httpResponse.data.errors) {
	  			$scope.searchMessage = httpResponse.data.errors.message[0];
	  		}
	  		else {
	  			$scope.searchMessage='There was error trying to query search result from server. Please try again later';
	  			
	  		}
	  
	  	}
	  	)
  	}
		, 100)
  
  	
  	
  	//
  	
  }

  // hightlight the search term in lineimage
  $scope.decorateLineImage  = function (lineImage, wordToHighlight, maxLen) {
  	// also cap the string to a maximum length
  	return $sce.trustAsHtml(unescape(escape(lineImage.substring(0, maxLen)).replace(new RegExp(wordToHighlight, 'gi'), '<strong>$&</strong>')));

  	
  }
  
  
  var languageExtMap=[]; // extension to language mapping for highlightjs
  
  
  languageExtMap["conf"]="apache";
  languageExtMap["cnf"]="apache";
  languageExtMap["sh"]="bash";
  languageExtMap["c"]="cpp";
  languageExtMap["h"]="cpp";
  languageExtMap["cpp"]="cpp";
  languageExtMap["hpp"]="cpp";
  languageExtMap["cc"]="cpp";
  languageExtMap["c++"]="cpp";
  languageExtMap["cp"]="cpp";
  languageExtMap["cxx"]="cpp";
  languageExtMap["css"]="css";
  languageExtMap["patch"]="diff";
  languageExtMap["diff"]="diff";
  languageExtMap["ini"]="ini";
  languageExtMap["java"]="java";
  languageExtMap["js"]="javascript";
  languageExtMap["cs"]="cs";
  languageExtMap["vb"]="cs";
  languageExtMap["jsl"]="cs";
  languageExtMap["wsf"]="cs";
  languageExtMap["xml"]="xml";
  languageExtMap["m"]="objectivec";
  languageExtMap["pl"]="perl";
  languageExtMap["php"]="php";
  languageExtMap["php3"]="php";
  languageExtMap["php4"]="php";
  languageExtMap["php5"]="php";
  languageExtMap["php6"]="php";
  languageExtMap["py"]="python";
  languageExtMap["rb"]="rb";
  languageExtMap["sql"]="sql";
  languageExtMap["md"]="markdown";
  languageExtMap["go"]="go";
  
  
	  
	function endsWith(str, suffix) {
	  return str.indexOf(suffix, str.length - suffix.length) !== -1;
	}

  // return lang based on extension
  function determineLanguage(rawUrl){
  	
  	var t = rawUrl.toLowerCase()
  	
  	
  	// some special files
  	if (endsWith(t, "makefile")) {
  		
  		return "apache";
  	} else
  	if (endsWith(t, "readme")) {
  		
  		return "markdown";
  	} else   	if (endsWith(t, ".min.js")) {
  		return "cgtext"; //nohighlight
  	}
  	
  	
  	
  	///
  	var ext = (/[.]/.exec(t)) ? /[^.]+$/.exec(t) : undefined;
  	
		var lang  = languageExtMap[ext]
		if (lang==undefined) {
			lang="cgtext"; // no highlight
		}
		
  	
  	return lang
  }
  
  
	var fileDownloadCanceller = null
	
	
  function getFileContentAndDisplay (rawUrl) {
  	
  	
  	
  	
  	$scope.dirList=null
  	$scope.isLoadingFileContent = true
  	//$scope.highlightLanguage = "c" // TODO: this has to be based on file extension
  	//$scope.highlightLanguage = "cgtext"//no highlight
  	$scope.highlightLanguage = determineLanguage(rawUrl)
  	fileDownloadCanceller   = $q.defer()
  	
  	$timeout (function () {
	  	$http.get(rawUrl,
	  	{
	  	timeout: fileDownloadCanceller.promise,
	  	cache:true,
	  	transformResponse:function (data, headersGetter) { // override the default transformation, which typically try to jsonify data
	  		
	  		return data;  // avoid jsonify the data, for e.g. when the file is a .json file
	  	}
	  		
	  	}).
	  		success(function(data, status, headers, config) {
	  			
					$scope.isLoadingFileContent=false
					$scope.sourceCode = data
					
	  		}).
	  		error(function(data, status, headers, config) {
	  			$scope.isLoadingFileContent=false
	  			$scope.errorMessage = data
	  			
	  			
	    
	  		});
  		
  	},500)
  	
  }
  
  
  function updateCurrentPath (path ) {
		$scope.currentPath = path
		$scope.currentPathItems = $scope.currentPath.split('/');
		$scope.currentPathItemsHref=[]
		$scope.currentPathItems[0]= $scope.projectInfo.name;
		
		$scope.searchPath = path
		
		var newPath="/view/project/"+$scope.projectInfo.id+path
		
		console.log ('updateCurrentPath to ',newPath )
		$location.path (newPath,false) //hackaround to avoid reloading - see the override function in app.js
		//console.log ("currentPath = "+$scope.currentPath)
		updatePrevLocation();
		
		var accumulatedPath = ""
		
		for (var i=0; i<$scope.currentPathItems.length; i++) {
			if (i==0) {
				accumulatedPath=""
				$scope.currentPathItemsHref[i]="/"
			} else {
				accumulatedPath += "/" + $scope.currentPathItems[i]
				$scope.currentPathItemsHref[i]=accumulatedPath
			}

			
		}
  	
  }
  
  // this function should only be called after we have properly $scope.projectInfo
  function obtainProjectSourcePath(sourcePath, shouldClearHash) {
  	// cancel any ongoing downloading
		if (fileDownloadCanceller){
			fileDownloadCanceller.resolve()
		}

  	$scope.isLoadingPath = true // trigger spinner display
  	if (shouldClearHash){
  		$location.hash("code-grep.com")// seems to be a bug here https://github.com/angular/angular.js/issues/9635  if set to empty or null it will cause error; so set to something not empty
  	}
  	
  	console.log ("obtainProjectSourcePath")
  	
  	$timeout (function (){
	  	ProjectTree.get (
		  	{projectId: $scope.projectInfo.id,full_path:sourcePath},//request will look like apiv1/project-tree/546702e53b7d9c5f57000003?full_path=%2Fbusybox-1.22.1%2Fprocps%2Fnmeter.c
		  	
		  		function	(response, responseHeaders) {// no error
		  			$scope.isLoadingPath=false
		  			$scope.errorMessage="";
		  			
		  			
		  			updateCurrentPath (sourcePath);
		
		  			
						/// TODO: should cache all of these later to avoid lots of  call to obtainProjectSourcePath
						
						
						
						
						if (response.is_dir) {
							// it is a directory listing
							
							$scope.dirList = response.dir_list
							 window.scrollTo(0,0);
		
		
						} else {
						
							// it is a file
							var rawUrl = settings.apiBase+"/project-raw/"+$scope.projectInfo.id+sourcePath
							getFileContentAndDisplay(rawUrl);
												
		
							
						}
		
		  		},
		  		function (httpResponse) { // error obtaining sourcePath
		  			//console.log (httpResponse);
		  			$scope.isLoadingPath=false
		  			if (httpResponse.data.errors) {
		  				$scope.errorMessage = httpResponse.data.errors.message[0];
		  			} else {
		  				$scope.errorMessage = "Error obtaing path information from server. Please try again later.";
		  			}
		
		  	
		  		}
		  	)
  		
  	},
  	100)
  	$scope.sourceCode=null;
	  	
  }
  
  // store the while $location so we can handle backward and forward properly
  function updatePrevLocation (){
  	  	
  	$scope.prevLocation ={
  		$$path : $location.$$path,
  		$$hash : $location.$$hash,
  		$$search : $location.$$search,
  	}

  }
 
  $scope.$on('$locationChangeSuccess', function(newState, oldState) {

  	
  	// TODO:based on $location, have to detect whether hash change, query change, projectid change, or source path
  	
  	var prevLocation  = $scope.prevLocation
  	
		updatePrevLocation();
		
  	if (!prevLocation) {
  		console.log ("$locationChangeSuccess 1")
 			
  		return;
  	}
  	//console.log ($location.$$path);
  	//console.log (prevLocation.hash+"  "+$location.$$hash);
  	//console.log ($location.$$search);
  	
  	// TODO: should handle  $location.$path change??
  	
	  if (($location.$$search.search)&&
	  	((prevLocation.$$search.search != $location.$$search.search)||
	  (prevLocation.$$search.type != $location.$$search.type)||
	  (prevLocation.$$search.page != $location.$$search.page)
	  )
	  )
	  {
	  	console.log ("search :" +prevLocation.$$search.search+" "+$location.$$search.search);
	  	//$location.hash("");// clear the hash
	  	console.log ("call handleSearchTerms 1")
	  	handleSearchTerms();
	  	return;
	  }
  	
  	
  	if (prevLocation.$$search.search &&  !$location.$$search.search) { // when going from a search display page back to a normal listing page
  		
			if (!$scope.isLoadingPath) {
				console.log ("$locationChangeSuccess 2")
				
				//$route.reload();
				var newPath = $location.$$path
				var currentPrefix = "/view/project/"+$scope.projectInfo.id
				var newSourcePath =newPath.substr (currentPrefix.length)
			
				console.log ("sourcePath ",newSourcePath )
		  	console.log ("call obtainProjectSourcePath 6")
			
		  	$scope.searchMessage='';
		  	$scope.searchResults=null
				$scope.isShowListOfTags = false
	
				obtainProjectSourcePath(newSourcePath, false)
				
				return
			}
  		
  	}
		if (prevLocation.$$hash != $location.$$hash){
			if (prevLocation.$$path ==$location.$$path) {
				//Same file but different line number hash. gotolinenumber()
				console.log ("$locationChangeSuccess 3")
				goToLineNumber();
				
				return;
			}
			
			
		}
		
		
		// when backward or forward buttons are used only, not when user click on dir or file
		if (prevLocation.$$path !=$location.$$path){
			
			if (!$scope.projectInfo) return;
			
			var newPath = $location.$$path
			var currentPrefix = "/view/project/"+$scope.projectInfo.id
			if (newPath.indexOf (currentPrefix)==-1) {
				console.log ("$locationChangeSuccess 4")
				// maybe different project page
				$route.reload();
			} else {
				var newSourcePath =newPath.substr (currentPrefix.length)
			
		  	console.log ("call obtainProjectSourcePath 1")

				obtainProjectSourcePath(newSourcePath, false)
			}
			
		
			return;
		}
		
		
		  	
  	
  	
  	
  });
  
  $scope.$on('$routeChangeStart', function(next, current) {
  	
  	// at the moment seems we do not need to cancek as ObtainProjectInfo will just compare the project id in the url and the result and if there is a different there will be no more request
  	//$timeout.cancel ($scope.periodicQueryProjectInfo);
  	// TODO: any necessary clean up of the current project (before transitioning to a new project) can be done here
	  	console.log("routechangestart: cancel any ongoing request if possible");

	  	
	  	
	  	


 	});
 	
  $scope.$on('$routeChangeSuccess', function () {

		console.log ("viewProject routeChangeSUccess")
		if ($scope.projectInfo)
			console.log (" project info from $scope :", $scope.projectInfo.id)
		
		updatePrevLocation();
		
    var projectId = $routeParams.projectId;
    SharedData.currentProject.data  = null;

		
    if (projectId.length !=24) {
    	$scope.errorMessage = "Invalid project ID";
    } else {
    	
    	
    	$scope.processingMessage ="Please wait..."
    	
			$timeout(function () { // first query to get projectInfo
				
				obtainProjectInfo(projectId)
		},500);
    	
    }

    
    
    


			    

    

  })

  
  
  
  function goToLineNumber(){
    
		// clear highlight any previous one
		if ($scope.currentLineElem){
			
			$scope.currentLineElem.className = "hljs-linenumber";
		}
		
    //var linehash = /\#L(\d+)/g.exec ($location.url());
    var linehash = /\L(\d+)/g.exec ($location.$$hash);
    
    if (linehash !=null) {
			
			
      var linenum = linehash[1];  // actual  line number
    	console.log ("goto line number "+linenum)
      
			
      // change class if the highlighted line
      var lineElem= document.getElementById("L"+linenum);
      if (!lineElem) return;
      
      lineElem.className = "hljs-highlight hljs-linenumber";
      $scope.currentLineElem  =lineElem
      
      $anchorScroll ();
        
      //because anchorScroll will go to top of screen which is covered by nav bar,
      // we have to scroll down a little bit
      $timeout(function(){

				
         window.scrollTo(window.pageXOffset, window.pageYOffset - 200);
        
      },100);
        
    }
    
  }

	
	function highLightDoneCallBack () {
		// get position of "L1" as a reference
		var line1Elem= document.getElementById("L1");
		$scope.line1ElemPos = getPositionOfElement(line1Elem);
		//
		
		goToLineNumber();
	}
    
	$scope.highlightDone = 	highLightDoneCallBack;


	
  $scope.isExpandTopBar = false
  

  $scope.expandCollapseTopBar = function () {
    if ($scope.isExpandTopBar ==false) {
  
      $scope.isExpandTopBar = true
    } else {
  
      $scope.isExpandTopBar = false
    }
  }

	// send cancel - only when project is underprocessing (i.e. downloading or scanning etc)
	$scope.cancelProject = function () {
			ProjectInfo.delete ({projectId:$scope.projectInfo.id}, function (response, responseHeaders){
				//console.log (response)
				// TODO: should stop polling and redirect to previous page??
			}, function (httpResponse) {
				//console.log (httpResponse)
			})
	}



	var $oLay = angular.element(document.getElementById('popup-overlay'))
	var overlayDisplay = 'none'
	
	function closePopupOverlayWhenClickingOutside(event, cbFunc) {

    var clickedElement = event.target;
    if (!clickedElement) return;

		//console.log (clickedElement.id)
    var elementClasses = clickedElement.classList;
    var clickedOnPopup = clickedElement.id=='popup-overlay' ||
    clickedElement.id=='popup-overlay-child1';
    
    if (!clickedOnPopup) {
        cbFunc();
    }

	}

  // handle identifier click
  $scope.handleIdentifierClick  = function (identifier, pageX, pageY, clientX, clientY) {
  	
  	//console.log ("line 1 pos "+$scope.line1ElemPos.x+" "+$scope.line1ElemPos.y)
  	
  	var identifierLineElem = document.elementFromPoint( $scope.line1ElemPos.x, clientY);

  	var identifierLine = 0 // no line number detected
  	if (identifierLineElem) {
  		identifierLine = parseInt (identifierLineElem.id.substr(1), 10)  // TODO:  used later for API calling
  	}
  	
  	
  	

    console.log ("TODO: should cancel any ongoing query ")
    
    {
    	
    	// display popup here
			
			if (overlayDisplay=='block') {
				overlayDisplay='none'
				$window.onclick= null
			} else {
				overlayDisplay = 'block'
				$scope.popupIdentifier = identifier
			
				//console.log ("TODO: should fire query for popup here for identifier :"+identifier+" line "+identifierLine+ " and path ..."+$scope.currentPath);
			
				// query for some definitions
				$scope.popupDefinitions=null
				$scope.popupReferences=null
				$scope.popupDisplayLimit = 3 // 3 items shown
				  	// search for tag first a
  			$scope.searchForTagAndPeformOp(identifier ,
  				function (response, responseHeaders) { // search for tag success
  					
						var tagList = response.results
						var tag = null
						for (var ii=0; ii< tagList.length; ii++) {
							
							if (identifier==tagList[ii].tag) {
							
								tag = tagList[ii]
								break
							}
						}
						if (tag!== null) { // found exact match
							// proceed to  searching using the tagid
							$timeout(function(){
								ProjectDetails.get({ // quick definition
									projectId: $scope.projectInfo.id,
									search_term: identifier,
									search_type: 'quick_defs_and_refs',
									search_page: 0,
									tag_id: tag.id,
									
							
								},
							
								function(response, responseHeaders) { // no error
									
									$scope.popupDefinitions =response.result_defs;
									$scope.popupReferences =response.result_refs;
									
								},
							
								function(httpResponse) { // error
								$scope.popupDefinitions =[]
								$scope.popupReferences =[]
									
								})
							},50);
								
								} else {
						 	$scope.popupDefinitions =[]
						 	$scope.popupReferences =[]
											
						}
  				},
  				function (httpResponse) { // search for tag error
						 	$scope.popupDefinitions =[]
						 	$scope.popupReferences =[]

		  		}
  			)

	
				
				
				
				

				
				
				$timeout(function () {
					$window.onclick = function (event) { // when clicking outside of the popup
				 		
				 	  closePopupOverlayWhenClickingOutside(event, function () {
				 	  	console.log ("TODO: should cancel any ongoing query");
				 	  	overlayDisplay='none'
							$oLay.css({display:overlayDisplay})
							$window.onclick= null
				 	 	})
				   
				 }
					
				},100)
			}
			
			
			var overLayCSS = {
			        left: pageX+'px' ,
			        top: pageY+10+'px' ,
			        display: overlayDisplay
			}
			
			$oLay.css(overLayCSS)
    }
    


  }
  



  // user click on 'see all references'
  
  
  // handle directory click (at the bar)
  // if last = true, it is equivalent to the current item so do not need to do anything
  $scope.handleDirClick = function (path , last) {
  	
  	// clear all search parameter
  	$scope.searchMessage='';
  	$scope.searchResults=null
		$scope.isShowListOfTags = false
  	$location.$$search = {};
  	
		

		
  	//if ((!last && !$scope.isLoadingPath)
  	if (( !$scope.isLoadingPath)
  	) {
  		//console.log("handleDirCLick :" + path)
  		
  		// only clear hash if it is currently showing line number
  		//console.log ("hash = "+$location.$$hash)
  		
  		//console.log ("call obtainProjectSourcePath 2")
  		obtainProjectSourcePath(path, true) // true means clear the hash
  		
  		
  	}
  }

	$scope.handleDirOrFileClick = function (item ) {
		var path
		if ($scope.currentPath.slice(-1)=='/') {
			path = $scope.currentPath + item.name
		} else {
			path = $scope.currentPath+'/'+item.name
		}
		
		
		if (item.is_dir) {
			console.log ("call obtainProjectSourcePath 3")
			obtainProjectSourcePath(path, true)
		} else {
			var rawUrl = settings.apiBase+"/project-raw/"+$scope.projectInfo.id+path
			
			updateCurrentPath(path) // update url bar and also the top-level path
			getFileContentAndDisplay(rawUrl);

		}
		
		
	}


	
	
	function hidePathSearchWhenClickingOutside(event, cbFunc) {

    var clickedElement = event.target;
    if (!clickedElement) return;

		
    var elementClasses = clickedElement.classList;
    var clickedOnPathSearchgBar = clickedElement.id=='path-search-bar' ;
    
    if (!clickedOnPathSearchgBar) {
        cbFunc();
    }

	}

	$scope.isShowingPathSearchBar = false


	
	$scope.searchForPath = function (val) {
		// at the moment seems we still have problem using ng-resource for typeahead

		return $http.get(settings.apiBase + "/project-tree/"+$scope.projectInfo.id, {
			params: {
				search_for: val,
			}
		}).then(function(response) {
			return response.data.results.map(function(item){
				
        return item;
      });
		})
		
	}
	
	// select from search list
	$scope.handlePathSearchSelect = function ($item, $model, $label) {
		
		console.log ("handle Path Search Select "+$item)
		$scope.isShowingPathSearchBar = false
		$window.onclick=null
		
		$location.$$search = {}; // clear all  search options
		$scope.searchResults = null
		$scope.isShowListOfTags = false
	
		console.log ("call obtainProjectSourcePath 4")
		obtainProjectSourcePath($item, true)
		

	}
	// show the path search input, and  clicking outside will hide it away
	$scope.handleStartPathSearch = function() {

		$scope.isShowingPathSearchBar = true
		$timeout(function () {
			$window.onclick = function (event) {
		 		
		 	  hidePathSearchWhenClickingOutside(event, function () {
					
					$scope.$apply(function() {
						$scope.isShowingPathSearchBar = false
						console.log ("TODO: cancel search if necessary??")
					})
		 	  	
					
					$window.onclick= null
		 	 	})
		   
		 }
			
		},100)

		
	}


	
	// try out text selection or text click
	$scope.handleMouseUp = function ($event) {
		
		
		if ($scope.highlightLanguage=="cgtext") {
			
			return; // not highlighted
		}
		var selectedText =''
		
		
		///////
		if ($window.getSelection) {
        selectedText= $window.getSelection();
    }   else
    if ($document.selection) {
        selectedText =  $document.selection.createRange().text;
    }
    
    
    if (selectedText !=''){
    	//console.log ("TODO : maybe good idea to handle text selection ??"+selectedText)
    } else {
    	
    	
    	
    	// when click on a single location, not scanning, try to get the word
			var selection = $window.getSelection();
			
    	// first make sure it is not comment block
    	
    	parent  = selection.getRangeAt(0).startContainer.parentNode;
    	while (parent.className){
    		//console.log (parent.className)
    		if (parent.className.indexOf("hljs-")==-1) {
    			parent = parent.parentNode;
    		} else {
    			
    			if ((parent.className.indexOf("hljs-keyword")!=-1) ||
    			parent.className.indexOf("hljs-comment")!=-1 ||
    			parent.className.indexOf("hljs-linenumber")!=-1
    			) {
    				return ;// don't handle
    			}
    			break;
    		}
    		
    	}
    	

			
			var data = selection.focusNode.data;
			if (!data) return; // data is undefined somehow; for e.g. when IE is in restricted mode
			var curPos = selection.focusOffset;
			var startPos=curPos;
			var endPos = curPos;
			var patt = /[A-Za-z0-9_]/;
			
			while (startPos>0 ) {
				if (patt.test (data[startPos-1])) startPos--; else {
					//console.log ("startPos failed at :"+data[startPos-1])
					break;
				}
			}
			while (endPos<data.length ) {
				if (patt.test (data[endPos]))				endPos++; else break;
	
			}
			//console.log ("startPos :"+startPos +" endPos :"+endPos)
			if (endPos >startPos) {
				var word = data.substring(startPos, endPos);
				
				// TODO: how to get line number?
			
				$scope.handleIdentifierClick (word, $event.pageX, $event.pageY, $event.clientX, $event.clientY )
				
			}
			
			
			
			

    }
    
    
	}



}])




;


// when code is display and mouse is click
function getSelectionPosition() {
//	alert ("abc");
	// NOTE:if text selection is done (in handleMouseUp), this function should not do anything
  var selection = window.getSelection();
  console.log(selection.focusNode.data[selection.focusOffset]);

}



function getPositionOfElement(element) {
    var xPosition = 0;
    var yPosition = 0;
  
    while(element) {
        xPosition += (element.offsetLeft - element.scrollLeft + element.clientLeft);
        yPosition += (element.offsetTop - element.scrollTop + element.clientTop);
        element = element.offsetParent;
    }
    return { x: xPosition, y: yPosition };
}