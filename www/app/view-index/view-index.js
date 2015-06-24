'use strict';

angular.module('cgApp.viewIndex', ['ngRoute'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/view/index', {
    templateUrl: 'view-index/view-index.html',
    controller: 'viewIndexCtrl'
  });
}])

.controller('viewIndexCtrl', ['$scope', 'LoggedinUser', function($scope, LoggedinUser) {
		$scope.getProfileData = LoggedinUser.getProfileData // its a function

	
}])



;

