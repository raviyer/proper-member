(function() {
  var app = angular.module('cluster');
  app.controller('networkController' , [ '$http', '$scope', function($http, $scope) {
     $scope.cy = null;
     $scope.selected_node = null;
     $scope.failpoints = null;
     $scope.join_host = "";
     $scope.join_port = 1235;
     $scope.fp_visible = function() {
       var b = $scope.selected_node !== null;
       console.log(b);
       return b;
     };
     $scope.mouseup = function() {
       $scope.src = null;
     }
     $scope.mousemove = function($event) {
       if ($scope.src) {
         $scope.src.offsetLeft = $event.clientX;
         $scope.src.offsetTop = $event.clientY;
       }
     }
     $scope.mousedown = function($event) {
       var src = $event.srcElement;
       if (src !== null) {
         var class_list = src.classList;
         if (class_list !== null) {
           for (var i in class_list) {
             if (class_list[i] == 'widget') {
               if ($event.clientX >= src.offsetLeft && $event.clientX <= (src.offsetLeft+src.offsetWidth)
                && $event.clientY >= src.offsetTop && event.clientY <= (src.offsetTop+10)) {
                 console.log("Border Clicked");
                 $scope.src = src;
               }
             }
           }
         }
       }
     };
     $scope.attach_host = function() {
       var n = $scope.selected_node;
       if (n !== null) {
         var uri = 'http://' + n.Ip + ':' + n.Port.toString() + '/attach?' + 'host=' + $scope.join_host
          + '&port=' + $scope.join_port.toString();
         console.log(uri);
         $http.get(uri).success(function(data) {
            $scope.refresh();
          });
       }
     };
     $scope.refresh_failpoints = function() {
       var n = $scope.selected_node;
       if (n) {
         $http.get('http://' + n.Ip + ':' + n.Port.toString() + '/failpoint').success(function(data) {
            $scope.failpoints = data;
          });
       } else {
         $scope.failpoints = null;
       }
     };
                 
     $scope.upload_failpoint = function(name) {
       var o = $scope.failpoints[name];
       var n = $scope.selected_node;
       if (n) {
         $http.post('http://' + n.Ip + ":" + n.Port.toString() + '/failpoint', o).success(function() {
            $scope.refresh_failpoints();
          });
       }
     };
     $scope.set_selection = function(evt) {
       if (evt.cyTarget === $scope.cy) {
         $scope.selected_node = null;
         $scope.failpoints = null;
         $scope.refresh_failpoints();
       } else {
         $scope.selected_node = evt.cyTarget.data().node;
         $scope.refresh_failpoints();
       }
     };
     /* make_node_element creates a cytoscape node element, from
      * the node object returned by the server
      */
     $scope.make_node_element = function(node) {
       var shape = "triangle";
       if (node.IsSeed && node.Peers) {
         shape = "circle";
       } else if (node.IsSeed) {
         shape = "rectangle";
       }
       var ne =  { data: {
           id: node.Guid,
           display: node.Name + "\n" + node.Ip + ":" + node.Port.toString(),
           shape: shape,
           node: node}};
       return ne;
     };
     /* refresh - Call to refresh the cytoscape network diagram */
     $scope.refresh = function() {
       $http.get("/members").success(function (data) {
          var nodes = [];
          var edges = [];
          for (i in data.Nodes) {
            var node = data.Nodes[i];
            var ne = $scope.make_node_element(node);
            nodes.push(ne);
            if (node.IsSeed) {
              // Draw the node group edges
              for (var s in node.Seedlings) {
                edges.push({data : {source: node.Seedlings[s],target: node.Guid}});
              }
              // Ping lines
              edges.push({data: {source: node.Guid, target: node.Seedlings[0]}, classes: "ping"});
              for (var k = 0; k < node.Seedlings.length-1; ++k) {
                var f = node.Seedlings[k];
                var t = node.Seedlings[k+1];
                edges.push({data: {source: f, target: t}, classes: "ping"});
              }
              edges.push({data: {source: node.Seedlings[node.Seedlings.length-1],
                    target: node.Guid}, classes: "ping"});
            }
          }
          var ct = document.getElementById('nw');
          console.log(ct);

          $scope.cy = cytoscape({
          layout: {
            name: 'grid',
                padding: 10
                },
              style: cytoscape.stylesheet()
              .selector('node')
              .css({
                 'shape': 'data(shape)',
                  'content': 'data(display)',
                  'text-valign': 'center',
                  'font-size' : "2"
                  })
              .selector(':selected')
              .css({
                 'border-width': 1,
                  'border-color': '#333'
                  })
              .selector('edge')
              .css({
                 'line-color': 'black',
                  'target-arrow-color': 'black',
                  'target-arrow-shape': 'triangle'
                  })
              .selector('edge.ping')
              .css({
                 'line-color': 'black',
                  'target-arrow-color': 'black',
                  'line-style': 'dotted',
                  'target-arrow-shape': 'triangle'
                  }),
              container: ct,
              elements: {nodes: nodes, edges: edges}});
          $scope.cy.on('tap', 'node', $scope.set_selection);
        });
     };

     $scope.refresh();
   }]);
  app.directive('network',  function() {
     return {restrict: 'E', templateUrl: 'network.html'};
   });
  app.directive('failpointList', function() {
     return { restrict: 'E', templateUrl: 'failpoints.html'};
     });
  app.directive('joinForm', function() {
     return { restrict: 'E', templateUrl: 'join.html'};
   });
  app.directive('dragable', ['$document', function($document) {
     return {
    restrict: 'A',
    link: function(scope, element, attr) {
         var startX = 0, startY = 0, x = 0, y = 0;

         element.css({
         position: 'relative'});

         element.on('mousedown', function(event) {
            // Prevent default dragging of selected content
            event.preventDefault();
            startX = event.pageX - x;
            startY = event.pageY - y;
            $document.on('mousemove', mousemove);
            $document.on('mouseup', mouseup);
          });

         function mousemove(event) {
           y = event.pageY - startY;
           x = event.pageX - startX;
           element.css({
           top: y + 'px',
               left:  x + 'px'
               });
         }

         function mouseup() {
           $document.off('mousemove', mousemove);
           $document.off('mouseup', mouseup);
         }
       }
     };
   }]);
   })();
