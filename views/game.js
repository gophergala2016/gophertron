window.addEventListener('load', function () {
    toastr.options = {
	"closeButton": false,
	"debug": false,
	"newestOnTop": false,
	"progressBar": false,
	"positionClass": "toast-top-right",
	"preventDuplicates": false,
	"onclick": null,
	"showDuration": "300",
	"hideDuration": "1000",
	"timeOut": "5000",
	"extendedTimeOut": "1000",
	"showEasing": "swing",
	"hideEasing": "linear",
	"showMethod": "fadeIn",
	"hideMethod": "fadeOut"
    };
    
    var canvas = document.getElementById("myCanvas");
    var ctx = canvas.getContext("2d");
    ctx.moveTo(0,0);

    var url = 'ws://' + window.location.hostname + ':' + window.location.port + '/websocket';
    console.log("Connecting to "+url);
    var ws = new WebSocket(url);

    ws.onclose = function(close) {
	console.log(close);
    }

    ws.onerror = function(error){
	console.log(error);
    }

    ctx.fillText("Waiting for players",10,50);

     ws.onopen = function() {
	console.log("Connected!")
	ws.onmessage = function(evt) {
            if (event.data == "countdown") {
		toastr["info"]("Game starting in 5 seconds")
		return
            }
	    if (event.data == "victory") {
		toastr["success"]("You won!")
		return
	    }
	    
	    ctx.clearRect(0,0,canvas.width, canvas.height);
            ctx.beginPath();
            ctx.lineWidth=10;
            //console.log(evt.data)
	    var paths = JSON.parse(evt.data);
	    for (id in paths) {
		ctx.moveTo(paths[id].coordinate[0].X*10, paths[id].coordinate[0].Y*10);
		ctx.strokeStyle = paths[id].color;
		for (i in paths[id].coordinate) {
		    ctx.lineTo(paths[id].coordinate[i].X*10, paths[id].coordinate[i].Y*10);
		}

		ctx.stroke();
	    }
	}
	document.onkeydown = function(e) {
	    e = e || window.event;
	    var js = {"request": "move"};
	    
	    if (e.keyCode == '38') {
		console.log('up');
		js["param"] = "up";
	    }
	    else if (e.keyCode == '40') {
		console.log('down');
		js["param"] = "down";
	    }
	    else if (e.keyCode == '37') {
		console.log('left');
		js["param"] = "left";
	    }
	    else if (e.keyCode == '39') {
		console.log('right');
		js["param"] = "right";
	    }
	    else {
		return;
	    }
	    ws.send(JSON.stringify(js));
	    
	}
    }
}
		       );
