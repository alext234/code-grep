'use strict';



angular.module('cgApp.viewSignup', ['ui.bootstrap'])



.config(['$routeProvider',
function($routeProvider) {

  $routeProvider.when('/view/signup', {
    templateUrl: 'view-signup/view-signup.html',
    controller: 'ViewSignupCtrl'
  });
}])




.controller ('ViewSignupCtrl', ['$scope','$timeout', 'User','LoggedinUser',
function ($scope,$timeout, User, LoggedinUser) {

	$scope.formInfo = {}
	$scope.isShowSuccess  = false
	$scope.isShowError  = false
	$scope.isShowSending = false
	$scope.isSubmittedSuccess = false

  $scope.message = "No message received from server"
  
	$scope.closeError = function () {
		$scope.isShowError = false;
	}



	$scope.submitForm= function () {
		if ($scope.signupForm.$valid) {
			$scope.isShowSending = true
			// valid form is here. Data is inside $scope.formInfo
			User.save ($scope.formInfo,
				function(data) {
				  
					$scope.isShowSuccess = true;
					$scope.isShowError  = false;
					$scope.isShowSending = false;
					$scope.isSubmittedSuccess = true;
				  
				  $scope.message = data.message
				  
				  
				  // still loggedin to allow user to start using first
				  LoggedinUser.setAuthToken(data.access_token)
				  
				  
				  
				},
				function (error) {
					//console.log (error.status); this is http status code
					$scope.message = error.data.message[0];
					
					$scope.isShowError  = true;
					$scope.isShowSuccess = false;
					$scope.isShowSending = false;


				}
			);
			
		}
  	
  
  };
}])



;