'use strict';

angular.module('cgApp.viewLogin', ['ngRoute'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/view/login', {
    templateUrl: 'view-login/view-login.html',
    controller: 'viewLoginCtrl'
  });
}])

.controller('viewLoginCtrl', ['$timeout','$location','$scope', 'UserAPINoLogin','LoggedinUser','$window',
function($timeout, $location, $scope, UserAPINoLogin,LoggedinUser, $window) {
	
	$scope.getProfileData = LoggedinUser.getProfileData // its a function
	$scope.errorMessage = null
	$scope.form = {};

	$scope.loginInProgress=false
	$scope.loginSuccessful= false
	
	$scope.handleLogin = function () {
		$scope.errorMessage = null
		$scope.loginSuccessful= false
		
		if ($scope.loginForm.$valid) {
				// fire api to verify username and password
    	  $timeout (function () {
    	  	$scope.loginInProgress=true
    	  	UserAPINoLogin.get ({
    	  			type: "login", // action type
    	  			email:$scope.form.email,
    	  			password:$scope.form.password,
    	  		},
						function(response, responseHeaders) { // no error
							$scope.loginInProgress=false
							$scope.loginSuccessful= true
							LoggedinUser.setAuthToken (response.access_token)

							
						},
						function (httpResponse) { // error
							$scope.loginInProgress=false
							//console.log (httpResponse)
							
							if (httpResponse.data.errors ) {
								$scope.errorMessage= httpResponse.data.errors.message[0]
							} else {
								$scope.errorMessage = "There was error while performing login with the server. Please try again later."
							}
							
						}
    	  	)
    	  }, 0);
	
		}

	}
	
}])



;

