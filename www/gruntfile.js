module.exports = function(grunt) {

	require('load-grunt-tasks')(grunt);

	grunt.config.init({
		uglify:{
			options:{
				beautify:false,
				mangle:false, // there is still some error to magle names:((
				compress:{
				  drop_console:true, // this will also remove console.log debug outputs
				}, 
			},
		},
		cacheBust: {
	    options: {
	      encoding: 'utf8',
	      algorithm: 'md5',
	      length: 16
	    },
	    assets: {
	        files: [{
	            src: ['dist/index.html']
	        }]
	    }
  	},
	  useminPrepare: {
	      html: 'app/index.html',
	      options: {
	        dest: 'dist'
	      }
	  },
	  usemin:{
	  	html:['dist/index.html']
	  },
	  
		clean: {
		 	dist :{
		 		src: ['dist/']
		 	
		 	}
		},
	  
	  copy:{
	    html: {
	    	files:[

	    	{
	    		src: './app/index.html',
	    		dest: 'dist/index.html'
	    	},
	    	{
	    		expand: true,
	    		cwd: 'app/view-index/',
	    		src: ['*.html'],
	    		dest: 'dist/view-index/'
	    	},
	    	{
	    		expand: true,
	    		cwd: 'app/view-contact/',
	    		src: ['*.html'],
	    		dest: 'dist/view-contact/'
	    	},
	   	  {
	    		expand: true,
	    		cwd: 'app/view-signup/',
	    		src: ['*.html'],
	    		dest: 'dist/view-signup/'
	    	},
	  	  {
	    		expand: true,
	    		cwd: 'app/view-project/',
	    		src: ['*.html'],
	    		dest: 'dist/view-project/'
	    	},
	  	  {
	    		expand: true,
	    		cwd: 'app/view-add-project/',
	    		src: ['*.html'],
	    		dest: 'dist/view-add-project/'
	    	}
  			,
	    	
	  	  {
	    		expand: true,
	    		cwd: 'app/view-edit-profile/',
	    		src: ['*.html'],
	    		dest: 'dist/view-edit-profile/'
	    	}
  			,

	  	  {
	    		expand: true,
	    		cwd: 'app/view-forgot-password/',
	    		src: ['*.html'],
	    		dest: 'dist/view-forgot-password/'
	    	}
  			,

	  	  {
	    		expand: true,
	    		cwd: 'app/view-login/',
	    		src: ['*.html'],
	    		dest: 'dist/view-login/'
	    	}
  			,

	  	  {
	    		expand: true,
	    		cwd: 'app/view-manage-projects/',
	    		src: ['*.html'],
	    		dest: 'dist/view-manage-projects/'
	    	}
  			,

	  	  {
	    		expand: true,
	    		cwd: 'app/img/',
	    		src: ['*.*'],
	    		dest: 'dist/img/'
	    	}
  			,

    	  {
	    		expand: true,
	    		cwd:'app',
	    		src: [
	    		'busy-small.gif',
					'search-glass.png',
	    //		'view-project/view-project.js',
	    // 		'app.js',
	    // 		'view-testonly/view-testonly.js',
	    // 		'view-add-project/view-add-project.js',
	    // 		'view-contact/view-contact.js',
	    // 		'view-index/view-index.js',
	    // 		'view-signup/view-signup.js',
	    // 		'components/version/version.js',
	    // 		'components/version/version-directive.js',
					// 'components/version/interpolate-filter.js',
	    // 		'bower_components/angular/angular.min.js*',
	    // 		'bower_components/angular-resource/angular-resource.min.js*',
	    // 		'bower_components/angular-route/angular-route.min.js*',
	    // 		'bower_components/angular-utf8-base64/angular-utf8-base64.min.js',
					// 'bower_components/angular-local-storage/dist/angular-local-storage.min.js'
	    		],
	    		dest: 'dist/'
	    	}


	    	]
	    
	    }
	  }
	});

	grunt.registerTask('default',[
		'clean',
		'copy:html',
		'useminPrepare',
		'concat',
		'uglify',
    'cssmin',
		'usemin',
		'cacheBust'
    ]);
}
