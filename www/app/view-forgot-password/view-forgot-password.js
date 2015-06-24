'use strict';

angular.module('cgApp.viewForgotPassword', ['ngRoute'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/view/forgot-password', {
    templateUrl: 'view-forgot-password/view-forgot-password.html',
    controller: 'viewForgotPassword'
  });
}])

.controller('viewForgotPassword', ['$timeout','$location','$scope', 'UserAPINoLogin','LoggedinUser','$window',
function($timeout, $location, $scope, UserAPINoLogin,LoggedinUser, $window) {
	
	$scope.infoMessage= null
	$scope.errorMessage= null
	$scope.submitResetSuccessful= false
	$scope.submitResetInProgress=false
	$scope.getProfileData = LoggedinUser.getProfileData
	
	$scope.form={}
	$scope.handleResetPassword = function () {
		$scope.infoMessage = null
		$scope.errorMessage= null
		$scope.submitResetSuccessful= false
		$scope.submitResetInProgress=false
		
		if ($scope.loginForm.$valid) {
				// fire api to submit password reset to server
				
    	  $timeout (function () {
    	  	$scope.submitResetInProgress=true
    	  	UserAPINoLogin.get ({
    	  			type: "password_reset", // action type
    	  			email:$scope.form.email,
    	  			
    	  		},
						function(response, responseHeaders) { // no error
							$scope.submitResetInProgress=false
							if (response.errors ) {
								$scope.errorMessage= response.errors.message[0]
							} else {
								$scope.submitResetSuccessful= true
								$scope.infoMessage = response.message

								
							}
						},
						function (httpResponse) { // error
							$scope.submitResetInProgress=false
							$scope.errorMessage = "There was error while performing password reset with the server. Please try again later."
							
						}
    	  	)
    	  }, 0);
	
		}

	}
	
}])



;

