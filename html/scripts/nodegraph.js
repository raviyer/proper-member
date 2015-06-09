function loadUri(u) {
  var x = new XMLHttpRequest();
  x.open('GET', u, true);
  x.onload = function() {
    console.log(x.responseText);
    members = JSON.parse(x.responseText);
    var nodes = [];
    var edges = [];
    for (var i in members.Nodes) {
      var n = members.Nodes[i];
      var col = "rgb(204, 255, 153)";
      if (n.IsSeed) {
        col = "rgb(255, 204, 255)";
      }
      if (n.IsSeed && n.Peers) {
        col = "rgb(153, 204, 255)";
      }
      
      var o = {id: n.Guid,
               label: n.Name + "\n" + n.Ip + ":"
               + n.Port.toString(), shape : "box",
               color: col, title: "Focker"};


      if (n.IsSeed) {
        // Draw the node group edges
        for (var s in members.Nodes[i].Seedlings) {
          edges.push({from: n.Seedlings[s],to: o.id,
              dashes:false, arrows : 'to', color:'purple'});
        }
        // Ping lines
        edges.push({from: n.Guid, to: n.Seedlings[0], dashes: true,
            arrows: 'to', color:'orange'});
        for (var k = 0; k < n.Seedlings.length-1; ++k) {
          var f = n.Seedlings[k];
          var t = n.Seedlings[k+1];
          edges.push({from: f, to: t, dashes:true, arrows : 'to',
              color:'orange'});
        }
        edges.push({from: n.Seedlings[n.Seedlings.length-1],
            to: n.Guid, dashes:true, arrows : 'to', color:'orange'});
      }

      nodes.push(o);
    };
    console.log(nodes);
    console.log(edges);
    
    var node_data = new vis.DataSet(nodes);
    var edge_data = new vis.DataSet(edges);
    
    // Create a network
    var data = { nodes : node_data, edges : edge_data };
    var options = 
    {
      edges: {
        smooth: true
      }
    };

    var container = document.getElementById('container');
    var network = new vis.Network(container, data, options);
  }
  x.send();
}

function loadNodes() {
  // Let's first initialize sigma:

  loadUri("/members");
}

