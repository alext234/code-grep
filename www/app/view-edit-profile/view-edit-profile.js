'use strict';

angular.module('cgApp.viewEditProfile', ['ngRoute'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/view/edit-profile', {
    templateUrl: 'view-edit-profile/view-edit-profile.html',
    controller: 'viewEditProfileCtrl',
    reloadOnSearch: false,
  });
}])

.controller('viewEditProfileCtrl', ['$timeout','$location','$scope', 'User','LoggedinUser','$window',
	function($timeout, $location, $scope, User,LoggedinUser, $window) {
		
	$scope.errorMessage = null
	$scope.infoMessage = null
	
	$scope.isShowChangePassword = false
	
	
	$scope.getProfileData = LoggedinUser.getProfileData // its a function
	
	$scope.form = {}
	$scope.startChangePassword = function () {
			$scope.isShowChangePassword = true
			$scope.form.password1=""
			$scope.form.password2=""
			$scope.errorMessage = null
			$scope.infoMessage = null
			
	}
	
	$scope.cancelChangePassword = function () {
			$scope.isShowChangePassword = false
			$scope.errorMessage = null
			$scope.infoMessage = null

	}
	$scope.minPasswordLength = 6
	$scope.handleSubmitPassword = function () {
		$scope.errorMessage = null
		$scope.infoMessage = null
		if ($scope.form.password1.length < $scope.minPasswordLength || $scope.form.password2.length < $scope.minPasswordLength) {
			$scope.errorMessage = "Password length must be at least "+$scope.minPasswordLength
			return
		}
		if ($scope.form.password1 != $scope.form.password2 ) {
			$scope.errorMessage = "Passwords do not match"
			return
		}
		
		// fire api to update password
	
		$timeout ( function () {
			User.update (
				{
					type: "update_password",
					password : $scope.form.password1,
					

					
				},
				function(response, responseHeaders) { // no error
					// check for status and handle properly
					$scope.infoMessage = "Password updated successfully"
					$scope.isShowChangePassword = false
					
				},
				
				function(httpResponse) { // error
				  
				  if (httpResponse.data.errors.message) {
				  	$scope.errorMessage= httpResponse.data.errors.message[0]
				  } else {
				  	$scope.errorMessage = "There was error updating your password. Please try again later"
				  }
				}
			)
		}, 50 )
			
	}
	
	// fire api to confirm new user
	function verify_user_confirm_string(confirm_user_id, confirm_string) {
		$timeout ( function () {
			User.update (
				{
					type: "verify_confirm_string",
					user_id : confirm_user_id,
					confirm_string: confirm_string,
					
				},
				function(response, responseHeaders) { // no error
					// check for status and handle properly

					if (response.errors ){
						$scope.errorMessage  =response.errors.message[0];
					} else {
					
						$scope.isShowChangePassword = true
					
					
						$scope.infoMessage = response.message + " Please set your password"
					
					
						var accessToken = response.access_token;
						// store authtoken to cookie
						LoggedinUser.setAuthToken(accessToken)
					
					}
					
					
				},
				
				function(httpResponse) { // error
				  
				  if (httpResponse.data.errors.message) {
				  	$scope.errorMessage= httpResponse.data.errors.message[0]
				  } else {
				  	$scope.errorMessage = "There was error confirming user account with the server. Please try again later."
				  }
				}
			)
		}, 50 )
	}

	// fire api to reset password
	function verify_user_reset_string(user_id, reset_string) {
		$timeout ( function () {
			User.update (
				{
					type: "verify_reset_string",
					user_id : user_id,
					reset_string: reset_string,
					
				},
				function(response, responseHeaders) { // no error
						$scope.isShowChangePassword = true
						
						
						$scope.infoMessage=response.message +" Please set your password"
						
						
						var accessToken = response.access_token;
						// store authtoken to cookie
						LoggedinUser.setAuthToken (accessToken)
						

					
					
					
				},
				
				function(httpResponse) { // error
				  
				  if (httpResponse.data.errors.message) {
				  	$scope.errorMessage= httpResponse.data.errors.message[0]
				  } else {
				  	$scope.errorMessage = "There was error contacting the server. Please try again later."
				  }
				}
			)
		}, 50 )
	}
	
	
	// should be called from routechangesuccess or locationchangesuccess
	function handleRouteChange () {
		var reset_string = $location.$$search.reset_string;
		var confirm_string = $location.$$search.confirm_string;
		var confirm_user_id = $location.$$search.user_id;
		
		// clear search information
		$location.search('user_id',null)
		$location.search('confirm_string',null)
		$location.search('reset_string',null)
		
		//console.log ("confirm_string = "+ confirm_string+"  user_id = "+confirm_user_id)
		if (confirm_string  && confirm_user_id) {
			// special handling when confirm_string and user_id exists
			verify_user_confirm_string (confirm_user_id, confirm_string);
		} else if (reset_string  && confirm_user_id) {
			// clear cookie
			LoggedinUser.setAuthToken (null)

			verify_user_reset_string (confirm_user_id, reset_string);
		}
		else { // not reset nor confirm case
			
			//not login
			if (LoggedinUser.getAuthToken()==null) {
				// have to go to login page
				$location.$$search = {};
				$location.hash("");
				$location.path ('/view/login', false); // go to login view
				$window.location.reload();

			}
			
		}
		

	}
	$scope.$on('$locationChangeSuccess', function(newState, oldState) {
		//console.log ("location change success")
		
		
	})

		
	$scope.$on('$routeChangeSuccess', function(newState, oldState) {
		console.log (" edit profile $routeChangeSuccess ")
		
		handleRouteChange()
		
	
  	

		
	})
}])



;

