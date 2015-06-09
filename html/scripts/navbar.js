(function() {
  var app = angular.module('cluster');
  app.directive('navbar', function() {
     var o = { restrict: 'E',
               templateUrl: 'navbar.html',
               controller:
               function() {
                 this.selection = 1;
                 this.setSelection = function (sel) {
                   this.selection = sel;
                 };
               },
               controllerAs: 'navCtrl'
     };
     return o;
   });
})();
