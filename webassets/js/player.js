
var npMsg = {"Operation":"np"};
var nextMsg = {"Operation":"next"};
var booMsg = {"Operation":"boo"};
var tuneMsg = {"Operation":"tune"};

function NewNpMsg() {
    return JSON.stringify(npMsg);
}

function StartWebSocket() {
    var ws = new WebSocket("ws://localhost:8666/ws");
    var nps = document.getElementById("nowPlaying");

    document.getElementById("idNext").onclick = function() {
        ws.send(JSON.stringify(nextMsg));
        console.log("next: " + nextMsg)
    };

    document.getElementById("idBoo").onclick = function() {
        ws.send(JSON.stringify(booMsg));
        console.log("boo");
    };

    document.getElementById("idTune").onclick = function () {
        ws.send(JSON.stringify(tuneMsg));
        console.log("tune");
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
                var update = "" + r.Title + "(" + r.Duration + ")";

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