
var npMsg = {"Operation":"np"};
var queueMsg = {"Operation":"queue"};
var nextMsg = {"Operation":"next"};
var booMsg = {"Operation":"boo"};
var tuneMsg = {"Operation":"tune"};
var isPlaying = false;
var streamSrc = "";

function StartWebSocket() {
    var qInput = document.getElementById("idQuery");

    var wsProto = "wss:";

    if (location.protocol === "http:") {
        wsProto = "ws:";
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
        ws.send(JSON.stringify(query));
        qInput.value = "";
    };

    $("#idQuery").autocomplete({
        serviceUrl: '/ta',
        onSelect: function (suggestion) {
            var query = {"Operation":"request","Query":suggestion.value};
            ws.send(JSON.stringify(query));
            qInput.value = "";
        }
    });

    ws.onopen = function() {
        var nps = document.getElementById("nowPlaying");
        nps.innerHTML = "Waiting for websocket data ...";

        ws.send(JSON.stringify(npMsg));

        setInterval(function getNowPlaying() {
            ws.send(JSON.stringify(npMsg));
        }, 1000);

        setInterval(function getPlayQueue() {
            ws.send(JSON.stringify(queueMsg))
        }, 1000);
    };

    ws.onmessage = function(e) {
        var r = JSON.parse(e.data);

        switch (r.Pkt) {
            case "np_r":
                var update = "" + r.Data.Title + " (" + r.Data.Duration + ",  " + r.Data.Rating + "/10)";

                nps.innerHTML = update;
                break;
            case "queue_r":
                var pqWrapper =$("#playQueue");
                var i = 0;

                if (r.Data.Size === 0) {
                    pqWrapper.hide();
                    e.preventDefault();
                    return;
                } else {
                    pqWrapper.show();
                }

                var newList = "";

                for (i=0; i<r.Data.Size; i++) {
                    newList += "<li>" + r.Data.Entries[i] + "</li>"
                }

                var playQueue = document.getElementById("pqList");
                playQueue.innerHTML = newList;
                pqWrapper.show();
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
    var player = document.getElementById("audiocontrols");
    var source = document.getElementById("streamSrc");
    var label = document.getElementById("idPlay");

    if (streamSrc === "") {
        if (source.src === "") {
            console.log("streamSrc and source.src both not set!");
            return
        }
        streamSrc = source.src;
    }

    if (isPlaying) {
        player.pause();
        player.currentTime = 0;
        source.src = "";
        label.value = "  Play ";
        isPlaying = false;
    } else {
        source.src = streamSrc;
        player.load();

        var promise = player.play();
        if (promise !== undefined) {
            promise.then(function() {}).catch(function(error) {
                console.log("Failed to open stream: " + error);
            });
        }

        label.value = "Pause";
        isPlaying = true;
    }
}

function main() {
    $(document).ready(function() {
        var playlist      = document.getElementById("idPlaylist");
        var play          = document.getElementById("idPlay");
        var volUp         = document.getElementById("idVolUp");
        var volDown       = document.getElementById("idVolDown");
        var audioControls = document.getElementById("audiocontrols");

        $("#playQueue").hide();

        playlist.onclick = function() {
            window.open("/playlist", "_blank");
        };

        play.onclick = function() {
            ToggleStream();
        };

        volUp.onclick = function() {
            if (audioControls.volume < 1) {
                audioControls.volume += 0.1;
            }
            console.log("up: " + audioControls.volume);
        };

        volDown.onclick = function() {
            if (audioControls.volume > 0.1) {
                audioControls.volume -= 0.1;
            }
            console.log("down: " + audioControls.volume);
        };

        window.addEventListener('keydown', function (e) {
            evt = e || window.event;
            if (evt.keyCode === 32) {
                if (document.activeElement.id !== "idQuery") {
                    evt.preventDefault();
                    ToggleStream();
                }
            }
        });

        StartWebSocket();
        ToggleStream();
    });
}

main();