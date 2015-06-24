'use strict';

angular.module('cgApp.version', [
  'cgApp.version.interpolate-filter',
  'cgApp.version.version-directive'
])

.value('version', '0.1');
