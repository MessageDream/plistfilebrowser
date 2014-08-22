var copyJson;
var app = angular.module('MobileAngularUiExamples', [
  "ngTouch",
  "mobile-angular-ui"
]);
    
app.config(function($interpolateProvider,$compileProvider) {
  $compileProvider.aHrefSanitizationWhitelist(/^\s*(https?|http|ftp|mailto|file|javascript|itms-services):/);
  $interpolateProvider.startSymbol('//');
  $interpolateProvider.endSymbol('//');
});

app.service('analytics', [
  '$rootScope', '$window', '$location', function($rootScope, $window, $location) {
    var send = function(evt, data) {
      ga('send', evt, data);
    }
  }
]);

app.controller('MainController', function($rootScope, $scope, analytics){

  $rootScope.$on("$routeChangeStart", function(){
    $rootScope.loading = true;
  });

  $rootScope.$on("$routeChangeSuccess", function(){
    $rootScope.loading = false;
  });

  $scope.search=function(){
    if ($scope.searchModel==""||$scope.searchModel==undefined) {
      $scope.scrollItems=json;
      return;
    };
    result=json.filter(function(file){ 
      return file.FileName.indexOf($scope.searchModel) == 0 ;
    });
    $scope.scrollItems=result;
  };

  $scope.namesort=function(){
  var result=new Array;
  if ($scope.sortname=='文件名↑') {
    $scope.sortname='文件名↓'
     result =json.sort(function(a,b){
      return a.FileName>b.FileName?-1:1;
    });

  }else{
    $scope.sortname='文件名↑'
     result =json.sort(function(a,b){
     return a.FileName>b.FileName?1:-1;
    });
  }
  $scope.scrollItems = result;
  };

  $scope.timesort=function(){
   var result=new Array;
  if ($scope.sorttime=='修改时间↑') {
    $scope.sorttime='修改时间↓'
     result =json.sort(function(a,b){
      return a.CreateTime>b.CreateTime?1:-1;
    });
  }else{
    $scope.sorttime='修改时间↑'
     result =json.sort(function(a,b){
       return a.CreateTime>b.CreateTime?-1:1;
    });
  }
   $scope.scrollItems = result;
  };

  $scope.resetsort=function(){
    $scope.scrollItems = copyJson;
  };
  copyJson=json.concat();
  $scope.scrollItems = json;

});