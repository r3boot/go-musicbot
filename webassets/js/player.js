
var npMsg = {"Operation":"np"};
var nextMsg = {"Operation":"next"};
var booMsg = {"Operation":"boo"};
var tuneMsg = {"Operation":"tune"};
var playlistMsg = {"Operation":"playlist"};

function StartWebSocket() {
    var ws = new WebSocket("ws://"+window.location+":8666/ws");
    var nps = document.getElementById("nowPlaying");

    document.getElementById("idNext").onclick = function() {
        ws.send(JSON.stringify(nextMsg));
    };

    document.getElementById("idBoo").onclick = function() {
        ws.send(JSON.stringify(booMsg));
    };

    document.getElementById("idTune").onclick = function () {
        ws.send(JSON.stringify(tuneMsg));
    };

    document.getElementById("idSearch").onclick = function() {
        var query = {"Operation":"play","Query":document.getElementById("idQuery").value};
        ws.send(JSON.stringify(query));
    };

    document.getElementById("idPlaylist").onclick = function() {
        window.open("/playlist", "_blank");
    }

    ws.onopen = function() {
        var nps = document.getElementById("nowPlaying");
        nps.innerHTML = "Waiting for websocket data ...";

        ws.send(JSON.stringify(npMsg));

        setInterval(function getNowPlaying() {
            ws.send(JSON.stringify(npMsg));
        }, 1000);
    };

    ws.onmessage = function(e) {
        var r = JSON.parse(e.data);

        switch (r.Pkt) {
            case "np_r":
                var update = "" + r.Data.Title + "(" + r.Data.Duration + ")  " + r.Data.Rating + "/10";

                nps.innerHTML = update;
                break;
            default:
                console.log("Unknown message received: " + r.Pkt)
        }
    };

    ws.onclose = function() {
        nps.innerHTML = "Websocket connection closed ..."
    };
}

function main() {
    $(document).ready(function() {
        StartWebSocket();
    });
}

main();