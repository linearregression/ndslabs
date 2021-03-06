/* global angular:false */

angular
.module('ndslabs')
/**
 * The Controller for our "Export Spec" Modal Window
 * 
 * @author lambert8
 * @see https://opensource.ncsa.illinois.edu/confluence/display/~lambert8/3.%29+Controllers%2C+Scopes%2C+and+Partial+Views
 */
.controller('ExportSpecCtrl', [ '$scope', '$log', '$uibModalInstance', '_', 'clipboard', 'Stacks', 'spec',
    function($scope, $log, $uibModalInstance, _, clipboard, Stacks, spec) {
  $scope.spec = angular.copy(spec);
  $scope.showAlert = _.find(Stacks.all, [ 'key', $scope.spec.key ]);
  
  // Remove unnecessary fields
  delete $scope.spec.$$hashKey;
  delete $scope.spec.updateTime;
  delete $scope.spec.createdTime;
  
  angular.forEach(Stacks.all, function(stack) {
    if (_.find(stack.services, [ 'service', $scope.spec.key ])) {
      $scope.showAlert = true;
    }
  });
  
  $scope.copy = function() {
    if (!clipboard.supported) {
      alert('Sorry, copy to clipboard is not supported');
      return;
    }
    
    clipboard.copyText(JSON.stringify($scope.spec, null, 4));
  };

  $scope.close = function() {
    $log.debug("Closing modal with dismissal!");
    $uibModalInstance.dismiss('cancel');
  };
}]);