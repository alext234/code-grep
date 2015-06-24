'use strict';

// Declare app level module which depends on views, and components
angular.module('cgApp', [
 	'ngRoute',
	'ngResource',
	'ngCookies',
	'cgApp.version',
	'cgApp.viewIndex',
	'cgApp.viewProject',
	'cgApp.viewAddProject',
	'cgApp.viewContact',
  'cgApp.viewSignup',
  'cgApp.viewEditProfile',
  'cgApp.viewLogin',
  'cgApp.viewForgotPassword',
  'cgApp.viewManageProjects',
	'ui.bootstrap',
	

])
.config(['$routeProvider', '$locationProvider',function($routeProvider,$locationProvider) {
  //$routeProvider.otherwise({redirectTo: '/view-add-project'});
  $routeProvider.otherwise({redirectTo: '/view/index'});
	$locationProvider.html5Mode(true)
}])

.directive('ngEnter', function () {
    return function (scope, element, attrs) {
        element.bind("keydown keypress", function (event) {
            if(event.which === 13) {
                scope.$apply(function (){
                    scope.$eval(attrs.ngEnter);
                });

                event.preventDefault();
            }
        });
    };
})

.directive('buttonsRadio', function() {
        return {
            restrict: 'E',
            scope: { model: '=', options:'='},
            controller: function($scope){
                $scope.activate = function(option){
                    $scope.model = option;
                };
            },
            template: "<button type='button' class='btn btn-sm'"+
            						"class='btn'"+
                        "ng-class='option == model ? \"btn-primary\": \"btn-info\"'"+
                        "ng-repeat='option in options' "+
                        "ng-click='activate(option)'>{{option}} "+
                      "</button>"
        };
    })

.controller('navBarCtrl', ['$http','$rootScope','$route' ,'$scope','$modal','SharedData' ,'$location', '$window','$timeout','Settings', 'LoggedinUser', function($http,$rootScope,$route,$scope,$modal,SharedData,$location,$window, $timeout,settings, LoggedinUser) {
	
	$http.defaults.headers.common.Authorization="auth_token="+LoggedinUser.getAuthToken(); // for other $http api call
	
	$scope.$on('$locationChangeSuccess', function () {
		console.log ("navbarCtrl $locationChangeSuccess")
  	var searchItems = $location.search()
  	console.log ("navbar controller route change success ")
  	if (searchItems['search']){
  		
  		$scope.searchTerm=searchItems['search']
	  	if (searchItems['type']){
	  			var t = searchItems['type'].toLowerCase()
  				if (t=='definitions') {
  					$scope.selectedSearchOption = 'Definitions'
  				} else   				if (t=='references') {
  					$scope.selectedSearchOption = 'References'
  				}
  				else if (t=='grep') {
  					$scope.selectedSearchOption = 'Grep'
  				}

	  	} else {
	  		$scope.selectedSearchOption='Definitions'
	  	}

  	}else {
  		$scope.searchTerm=''
  		$scope.selectedSearchOption='Definitions';
  	}
		
	})
	
  $scope.$on('$routeChangeSuccess', function () {
  	

  })
	
	$scope.isActiveSearchForm = function(viewLocation) {
		
    return  $location.path().indexOf(viewLocation)==0;
	};
	///
	function setPositionOfNavBarSearch () {
		
		var elem = document.getElementById("navbarSearchForm");
		var marginLeft = $window.innerWidth / 2-430
		elem.setAttribute("style", "margin-left: " + marginLeft + "px");
		
	}
	$window.onresize = function (event) {
		setPositionOfNavBarSearch();
	}
	setPositionOfNavBarSearch();
	
	$scope.searchTerm=''
	$scope.searchEnterKey = function () {
	
		$location.$$search = {}; //clear all search
		if ($scope.searchTerm != '') {
			if ($scope.selectedSearchOption == "Definitions") {
				$location.search({
					search: $scope.searchTerm,
					type: 'definitions',
					page:0
				})
			}
			else if ($scope.selectedSearchOption == "References") {
				$location.search({
					search: $scope.searchTerm,
					type: 'references',
					page:0
				})
			}
			else if ($scope.selectedSearchOption == "Grep") {
				$location.search({
					search: $scope.searchTerm,
					type: 'grep',
					page:0
				})
	
			}
	
	
		}
	}


	
  ////////
  
	//$scope.searchOptions = ["Definitions", "References", "Grep"];
	$scope.searchOptions = ["Definitions", "References"];
	$scope.selectedSearchOption = ""
	
	// $scope.$watch('selectedSearchOption', function(v){
	//   //console.log('changed', v);
	// });

	/////////////////
	
	$scope.currentProject = SharedData.currentProject;
	
	$scope.goToView = function (viewLink) {
		
		$timeout(function () {
			$location.$$search = {}; //clear all search
			$location.hash("");
			$location.path (viewLink, false);
			$window.location.reload();

			
		}, 100)
	}
	
	
	$scope.getProfileData = LoggedinUser.getProfileData // its a function
	
	$scope.handleLogout = function () {
		LoggedinUser.setAuthToken(null)
		$window.location.reload();
		
	}
	
  $scope.openAddProjetDialog = function ($event) {
  	
  	//$event.target.parentElement.parentElement.parentElement.click();
  	//console.log ();
  	$timeout (function () {
  		// ugly workaround to close the menu when clicking https://github.com/angular-ui/bootstrap/issues/796
  		$event.target.parentElement.parentElement.parentElement.click()
  	}, 10);
  	
  	
    var modalInstance = $modal.open({
      templateUrl: 'view-add-project/add-project-modal.html',
      controller: 'addProjectCtrl',
      backdrop : 'static', // prevent user from closing when clicking mouse in the backdrop
      keyboard: false, // prevent user from closing by pressing ESC

      size: 'lg',
      resolve: {
      }
    });

    modalInstance.result.then(function () {
      
      // closed by OK
      // TODO: start processing here
      //alert (selectAddVia);
    }, function () {
      // closed by ESC, mouse clicked outside area, or mouse clicked Cancel
      //$window.history.back();
    });

  }
  
}])


// shared REST API resource for user access
.factory("User", ['$resource', 'Settings',function($resource, settings) {
  return $resource(settings.apiBase+"/user", null,
  {
  		'update' :{  // for verifying confirm string, update password etc
  			method:'PUT'
  		}
  }
  );
}])

// api for user login or password reset
.factory("UserAPINoLogin", ['$resource', 'Settings',function($resource, settings) {
  return $resource(settings.apiBase+"/user-not-loggedin", null,
  {
  }
  );
}])



.factory("ProjectInfo", ['$resource', 'Settings',function($resource, settings) {
  return $resource(settings.apiBase+"/project-info/:projectId", null,
  {
  	'delete':
  		{
  			method: 'DELETE',

  		},
		'update' :{  // for verifying confirm string, update password etc
			method:'PUT',
			params: {
                projectId: "@id" // this will make the projectId appears in the url for PUT request
            }
		}
  		


  }
  );
  // NOTE: save --> 'POST' method : create new one
  // 			 update --> 'PUT' method : update existing
}])

.factory("ManageProjects", ['$resource', 'Settings',function($resource, settings) {
  return $resource(settings.apiBase+"/manage-projects", null,
  {


  }
  
  );
  // NOTE: save --> 'POST' method : create new one
  // 			 update --> 'PUT' method : update existing
}])

// for files/directories listing/searching
.factory("ProjectTree", ['$resource', 'Settings', function($resource, settings) {
	  return $resource(settings.apiBase+"/project-tree/:projectId", null,
  {

  }
  );
}])


// for definitions/references/grep search
.factory("ProjectDetails", ['$resource', 'Settings', function($resource, settings) {
	  return $resource(settings.apiBase+"/project-details/:projectId", null,
  {

  }
  );
}])


.factory("UploadLimit", ['$resource', 'Settings', function($resource, settings) {
  return $resource(settings.apiBase+"/upload-limit", null,
  {

  }
  );
}])


// return  authorization token and some other info such as account details
.factory ('LoggedinUser', ['$cookies', '$timeout','User',function ($cookies, $timeout, User) {

	var profileData =  null
	var authToken =null;
	
	
	if ($cookies.auth_token)  {
		if ($cookies.auth_token!=null && $cookies.auth_token!="null" && $cookies.auth_token!='undefined') {
			authToken = $cookies.auth_token
			queryUserProfile()
		}
	}
	
	function queryUserProfile () {
	    	  
    	  // fire the api to retrieve profile data
    	  $timeout (function () {
    	  
    	  	User.get ({
    	  			type: "profile_data" // what to get
    	  		},
						function(response, responseHeaders) { // no error
							if (response.data) {
								profileData = response.data
							} else {
								profileData=null
							}
							//console.log (profileData)
						},
						function (httpResponse) { // error
							profileData=null
							
						}
    	  	)
    	  }, 0);

	}
		
  return {
  	
  		getProfileData : function () {
  			if (!profileData) return null;
  			return profileData;
  		},
      getAuthToken : function () {
      	return authToken
      },
      setAuthToken: function (newAuthToken) {
      		if (newAuthToken == $cookies.auth_token)  {
      			return;
      		}
      		
      		
      	  $cookies.auth_token = newAuthToken
      	  if (newAuthToken) {
      	  	queryUserProfile();
      	  }
      }
  };
}])

// act like global settings
.factory('Settings', function() {
  return {
      apiBase : 'apiv1'
  };
})


// global data shared accross all controllers etc
.factory('SharedData', function () {
	return{
		currentProject:{
			data: null// use another data variable to point to the actual data so the binding will work when it is updated from another controller
		},
		addProjectDialogViaMethod : "url", // the dialog tab selection:  "url" or "upload",
		
	}
})


.run(['$route', '$rootScope', '$location',function ($route, $rootScope, $location) {
    var original = $location.path;
    $location.path = function (path, reload) {
        if (reload === false) {
            var lastRoute = $route.current;
            var un = $rootScope.$on('$locationChangeSuccess', function () {
                $route.current = lastRoute;
                un();
            });
        }
        return original.apply($location, [path]);
    };
}])



.factory('serverResponseInterceptor', ['$cookies','$location','$window','$q', function($cookies,$location,$window,$q) {
    var interceptor = {
        responseError : function (rejection) {
        	
        	//console.log (rejection)
        	if (rejection.status==401 && rejection.data.reason ) {
        		// when server send back like this 							server_utils.SendJsonResponseWithStatusCode (w,
								// map[string][]string{"message":[]string{"Access token expired"},
								//"reason":[]string{"token_expired"}}, http.StatusUnauthorized)

        		if (rejection.data.reason[0]=="token_expired") {
        			
        			// can't inject LoggedinUser due to circular dependecy injection
        			// so have to set cookie directly
        			$cookies.auth_token=null
    					$location.$$search = {}; //clear all search
							$location.hash("");
							$location.path ('/view/login', false); // go to login view
							$window.location.reload();
							return
        		}
        	}
  	      return $q.reject(rejection);

        }
    };
    return interceptor;
}])

.config(['$httpProvider', function($httpProvider) {
    $httpProvider.interceptors.push('serverResponseInterceptor');
}])