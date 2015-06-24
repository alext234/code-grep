'use strict';



angular.module('cgApp.viewContact', ['ui.bootstrap'])



.config(['$locationProvider','$routeProvider',
function($locationProvider,$routeProvider) {
  $routeProvider.when('/view/contact', {
    templateUrl: 'view-contact/view-contact.html',
    controller: 'ViewContactCtrl'
  });
  
 
}])

.factory("PostContactUs", ['$resource', 'Settings', function($resource, settings) {
  return $resource(settings.apiBase+"/contact-us");
}])

.controller ('ViewContactCtrl', ['$scope','$timeout', 'PostContactUs',
function ($scope,$timeout, postContactUs) {
	$scope.formInfo = {}
	$scope.isShowSuccess  = false
	$scope.isShowError  = false
	$scope.isShowSending = false
	$scope.isSubmittedSuccess = false

	$scope.closeError = function () {
		$scope.isShowError = false;
	}
	$scope.closeSuccess = function () {
		$scope.isShowSuccess = false;
	}



	$scope.submitForm= function () {
		if ($scope.contactForm.$valid) {
			$scope.isShowSending = true
			// valid form is here. Data is inside $scope.formInfo
			postContactUs.save ($scope.formInfo,
				function(data) {
					$scope.isShowSuccess = true;
					$scope.isShowError  = false;
					$scope.isShowSending = false;
					$scope.isSubmittedSuccess = true;
				},
				function (error) {
					//console.log (error.status);
					$scope.isShowError  = true;
					$scope.isShowSuccess = false;
					$scope.isShowSending = false;


				}
			); // save
			
		}
  	
  
  };
}])



;