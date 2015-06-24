'use strict';

angular.module('cgApp.viewManageProjects', ['ngRoute'])

.config(['$routeProvider', function($routeProvider) {
  $routeProvider.when('/view/manage-projects', {
    templateUrl: 'view-manage-projects/view-manage-projects.html',
    controller: 'viewManageProjectCtrl',
    reloadOnSearch: false,
  });
}])
.filter('format_status', function() {
  return function(input) {
    if (input=="uploaded") return "Uploaded";
    if (input=="extracting") return "Extracting";
    if (input=="url_received") return "URL received";
    if (input=="fetching") return "Fetching";
    if (input=="tree_scanning") return "Scanning";
    if (input=="analyzing") return "Analyzing";
    if (input=="ready") return "Ready";
    // not handle "error"
    return "";
    // "uploaded", "extracting"
																																			// "url_received", "fetching"
																																			// "tree_scanning"
																																			// "analyzing"
																																			// "ready"
																																			// "error"
																																			
    
  };
})

					
.controller('viewManageProjectCtrl', ['$timeout','$location','$scope', 'User','LoggedinUser','$window','ManageProjects', '$modal','ProjectInfo',function($timeout, $location, $scope, User,LoggedinUser, $window, ManageProjects,$modal, ProjectInfo) {
		
	$scope.errorMessage = null
	$scope.infoMessage = null
	

	$scope.getProfileData = LoggedinUser.getProfileData // its a function
	
	$scope.handleDelete = function (name,projectId) {
		
    var modalInstance = $modal.open({
      templateUrl: 'view-manage-projects/delete-project-confirm.html',
      controller: 'confirmDeleteProjectCtrl',
      size: 'sm',
      resolve: {
      	projectId:function () {
      		return projectId
      	},
      	name: function () {
      		return name
      	},
      	
      }
    });

    modalInstance.result.then(function () {
    	// deletion confirm, send api to delete
			ProjectInfo.delete ({projectId:projectId}, function (response, responseHeaders){

				refreshListOfProjects ()
			}, function (httpResponse) {
				//console.log (httpResponse)
			})

    	
      
    }, function () {
    	//
    	
    });

	}
	

	$scope.handleEdit = function (name,projectId) {
		
    var modalInstance = $modal.open({
      templateUrl: 'view-manage-projects/edit-project-dialog.html',
      controller: 'editProjectCtrl',
      size: 'lg',
      resolve: {
      	projectId:function () {
      		return projectId
      	},
      	name: function () {
      		return name
      	},
      	
      }
    });

    modalInstance.result.then(function () {
  		refreshListOfProjects ()
    	

    	
      
    }, function () {
    	//
    	
    });

	}
	

	
	$scope.projectList == null
	
	function refreshListOfProjects () {
		$scope.totalUsedSpace = 0
		// get list of projects
		ManageProjects.get({},
		function (response, responseHeaders) { // rest api good response
			if (response.projects) {
				var projects = response.projects
				for (var i=0; i<projects.length; i++ ){
					$scope.totalUsedSpace+= projects[i].total_size
				}
				$scope.projectList = projects
				

			}
			
			if ($scope.projectList) {
				if($scope.projectList.length ==0 ){
					$scope.infoMessage = "You have no project at the moment. "
				}
				
			}
			
		},
		function (httpResponse) { // error getting project info
				if (httpResponse.data.errors) {
			    		$scope.errorMessage = httpResponse.data.errors.message[0];
			    		$scope.processingMessage= '';
			    	} else {
			    		$scope.errorMessage = "There was an error obtaining project listings. Please try again later.";
			    		$scope.processingMessage= '';
			    	}
				}
		)
		
	}
	
	// should be called from routechangesuccess
	function handleRouteChange () {
		if (LoggedinUser.getAuthToken()==null) {
			$scope.errorMessage = "Please login to manage your projects"
			return
		}
		
		refreshListOfProjects ()
	}
	
	$scope.$on('$locationChangeSuccess', function(newState, oldState) {
		//console.log ("location change success")
		
		
	})

		
	$scope.$on('$routeChangeSuccess', function(newState, oldState) {
		console.log (" edit profile $routeChangeSuccess ")
		
		handleRouteChange()
		
	
  	

		
	})
}])

.controller('confirmDeleteProjectCtrl', ['$scope', '$modalInstance','projectId', 'name',function($scope, $modalInstance,projectId, name) {
	$scope.name  = name
	$scope.ok= function () {
		$modalInstance.close()
	}
	$scope.cancel  = function () {
		$modalInstance.dismiss('cancel')
	}

}])


.controller('editProjectCtrl', ['$scope', '$modalInstance','projectId', 'name','ProjectInfo',function($scope, $modalInstance,projectId, name, ProjectInfo) {
	
	$scope.form = {}
	$scope.name= name
	
	$scope.ok= function () {
		$modalInstance.close()
	}
	$scope.cancel  = function () {
		$modalInstance.dismiss('cancel')
	}
	
	$scope.saveDisabled = true // disable the save button
	
	// get project info
	$scope.infoMessage = "Loading project information..."
	$scope.errorMessage = null
	$scope.form_temp = {}
	$scope.form_temp.isAllowPublicView = false
	
	ProjectInfo.get ({projectId:projectId},
		function (projectInfo, responseHeaders) {
			$scope.infoMessage = null
			$scope.saveDisabled = false // enable the save button
			$scope.form = projectInfo
			if ($scope.form.view_permission=="public") {
				$scope.form_temp.isAllowPublicView = true
				
			}
			console.log ($scope.form)
			
		}, function (httpResponse) { // error
			$scope.infoMessage = null
			if (httpResponse.data.errors) {
				$scope.errorMessage = httpResponse.data.errors.message[0];
			} else {
				$scope.errorMessage = "Failed to get project information. Please try again"
			}
		}
	)
	
	$scope.ok= function () {
		
		if ($scope.form.name=="") {
			$scope.errorMessage = "Invalid name"
			
		} else {
			
			if ($scope.form_temp.isAllowPublicView ) {
				$scope.form.view_permission = "public"
			} else {
				$scope.form.view_permission = "private"
			}
			ProjectInfo.update({
				id:projectId,
				name:$scope.form.name,
				view_permission:$scope.form.view_permission,
			},
				function (projectInfo, responseHeaders) {
					$modalInstance.close()
					
				},
				function (httpResponse) {
					$scope.errorMessage ="Failed to save project. Please try again later"
				}
			)
		
		}
		
	}
	$scope.cancel  = function () {
		
		$modalInstance.dismiss('cancel')
	}


	

}])




;

