
var npMsg = {"Operation":"np"};
var nextMsg = {"Operation":"next"};
var booMsg = {"Operation":"boo"};
var tuneMsg = {"Operation":"tune"};
var playlistMsg = {"Operation":"playlist"};
var isPlaying = false;

function StartWebSocket() {
    var wsProto = "ws:";

    if (location.protocol == "https") {
        wsProto = "wss:";
    }

    var ws = new WebSocket(wsProto + "//"+window.location.host+"/ws");
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
        var qInput = document.getElementById("idQuery");
        var query = {"Operation":"play","Query":qInput.value};
        ws.send(JSON.stringify(query));
        qInput.value = "";
    };

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
                var update = "" + r.Data.Title + " (" + r.Data.Duration + ",  " + r.Data.Rating + "/10)";

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

function ToggleStream() {
    var stream = document.getElementById("audiocontrols");
    var label = document.getElementById("idPlay");

    if (isPlaying) {
        stream.pause();
        label.value = "  Play ";
        isPlaying = false;
    } else {
        stream.play();
        label.value = "Pause";
        isPlaying = true;
    }
}

function main() {
    $(document).ready(function() {
        StartWebSocket();
        ToggleStream();

        document.getElementById("idPlaylist").onclick = function() {
            window.open("/playlist", "_blank");
        };

        document.getElementById("idPlay").onclick = function() {
            ToggleStream();
        };

        document.getElementById("idVolUp").onclick = function() {
            var ac = document.getElementById("audiocontrols");
            if (ac.volume < 1) {
                ac.volume += 0.1;
            }
            console.log("up: " + ac.volume);
        };

        document.getElementById("idVolDown").onclick = function() {
            var ac = document.getElementById("audiocontrols");
            if (ac.volume > 0.1) {
                ac.volume -= 0.1;
            }
            console.log("down: " + ac.volume);
        };

        window.addEventListener('keydown', function (e) {
            var ac = document.getElementById("audiocontrols");
            var qInput = document.getElementById("idQuery");

            evt = e || window.event;
            if (evt.keyCode == 32) {
                if (document.activeElement.id != "idQuery") {
                    evt.preventDefault();
                    ToggleStream();
                }
            }
        });
    });
}

main();