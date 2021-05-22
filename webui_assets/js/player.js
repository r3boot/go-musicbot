"use strict";

const wsRequestNowPlaying = 0;
const wsRequestNext       = 1;
const wsRequestBoo        = 2;
const wsRequestTune       = 3;
const wsRequestQuery      = 4;
const wsRequestTrack      = 5;
const wsRequestQueue      = 6;

const wsReplyNowPlaying = 100;
const wsReplyNext       = 101;
const wsReplyBoo        = 102;
const wsReplyTune       = 103;
const wsReplySearch     = 104;
const wsReplyTrack      = 105;
const wsReplyQueue      = 106;

const msgInformation    = 0;
const msgWarning        = 1;
const msgError          = 2;

// Leave this global
var ws = null;

function filenameToTrack(fname) {
    return fname.substring(0, fname.length - 16);
}

function prettyDuration(duration) {
    var hours = ~~(duration / 3600);
    var minutes = ~~((duration % 3600) / 60);
    var seconds = Math.floor(duration % 60);

    if (hours > 0) {
        return hours + "h" + minutes + "m" + seconds + "s";
    } else if (minutes > 0) {
        return minutes + "m" + seconds + "s";
    }

    return seconds + "s";
}

function ShowNotification(type, message) {
    switch (type) {
        case msgInformation:
            console.log("INFO: " + message);
            break;
        case msgWarning:
            console.log("WARNING: " + message);
            break;
        case msgError:
            console.log("ERROR: " + message);
            break;
    }
}

function UpdateNowPlaying(data) {
    var track = filenameToTrack(data.filename);

    $("#divNowPlaying").html(track);
    $("#divRating").html(data.rating + "/10");
    $("#divDuration").html(prettyDuration(data.duration));
}

function RequestTrack(parent) {
    var yid = parent.id;

    wsSendCommandWithData(ws, wsRequestTrack, yid);

    ClearSearch();
}

function ClearSearch() {
    document.getElementById("txtQuery").value = "";
    UpdateSearchResults([]);
}

function UpdateSearchResults(data) {
    var searchResults = "";
    var yid = "";
    if ((data != null) && data.length > 0) {
        for (var i = 0; i < data.length; i++) {
            yid = data[i].substring(data[i].length - 15, data[i].length - 4);
            searchResults += "<div id=\"" + yid + "\" onClick=\"RequestTrack(this)\">Q</div>" + filenameToTrack(data[i]) + "<br/>";
        }
    }
    $("#divSearchResults").html(searchResults);
}

function UpdateQueueEntries(data) {
    var queueEntries = "";

    for (const [idx, filename] of Object.entries(data)) {
        queueEntries += idx + ") " + filenameToTrack(filename) + "<br/>";
    }
    $("#divQueueEntries").html(queueEntries);
}

function wsSendCommand(ws, commandId) {
    var request = {"o": commandId};
    ws.send(JSON.stringify(request));
}

function wsSendCommandWithData(ws, commandId, data) {
    var request = {
        "o": commandId,
        "d": data,
    };
    ws.send(JSON.stringify(request));
}

function WebSocketHandler() {
    var wsProto = "wss:";

    if (location.protocol === "http:") {
        wsProto = "ws:";
    }

    function startHandler() {
        ws = new WebSocket(wsProto + "//" + window.location.host + "/ws");

        setInterval(function() {
            wsSendCommand(ws, wsRequestNowPlaying);
            wsSendCommand(ws, wsRequestQueue);
        }, 500);

        $("#btnNext").click(function(ev) {
            ev.preventDefault();
            wsSendCommand(ws, wsRequestNext);
        });

        $("#btnBoo").click(function(ev) {
            ev.preventDefault();
            wsSendCommand(ws, wsRequestBoo);
        });

        $("#btnTune").click(function (ev) {
            ev.preventDefault();
            wsSendCommand(ws, wsRequestTune);
        });

        $("#btnSearch").click(function (ev) {
            ev.preventDefault();
            var query = document.getElementById("txtQuery").value;
            if ((query != null) && (query.length > 0)) {
                wsSendCommandWithData(ws, wsRequestQuery, query);
            } else {
                ClearSearch();
                console.log("No query")
            }
        })

        ws.onmessage = function (e) {
            var response = JSON.parse(e.data);

            switch (response.o) {
                case wsReplyNowPlaying:
                    if (response.s === true) {
                        UpdateNowPlaying(response.d);
                    } else {
                        ShowNotification(msgWarning, response.m);
                    }
                    break;
                case wsReplyNext:
                    if (response.s === true) {
                        ShowNotification(msgInformation, "Skipped to " + filenameToTrack(response.d.filename));
                    } else {
                        ShowNotification(msgWarning, response.m);
                    }
                    break;
                case wsReplyBoo:
                    if (response.s === true) {
                        ShowNotification(msgInformation, "Rating for " + filenameToTrack(response.d.filename) + " is now " + response.d.rating);
                    } else {
                        ShowNotification(msgWarning, response.m);
                    }
                    break;
                case wsReplyTune:
                    if (response.s === true) {
                        ShowNotification(msgInformation, "Rating for " + filenameToTrack(response.d.filename) + " is now " + response.d.rating);
                    } else {
                        ShowNotification(msgWarning, response.m);
                    }
                    break;
                case wsReplySearch:
                    if (response.s === true ) {
                        UpdateSearchResults(response.d);
                    } else {
                        ClearSearch();
                        ShowNotification(msgWarning, response.m);
                    }
                    break;
                case wsReplyTrack:
                    if (response.s === true ) {
                        console.log("priority: " + response.d)
                    } else {
                        ShowNotification(msgWarning, response.m);
                    }
                    break;
                case wsReplyQueue:
                    if (response.s === true ) {
                        UpdateQueueEntries(response.d);
                    } else {
                        ShowNotification(msgWarning, response.m);
                    }
                    break;
                default:
                    console.log("Received unknown websocket response: " + response.o);
            }
        };

        ws.onclose = function () {
            ShowNotification(msgError, "Websocket connection closed");
        };
    }

    function checkForDisconnect() {
        if (!ws || ws.readyState == 3) startHandler();
    }
    setInterval(checkForDisconnect, 5000);

    startHandler();

    return ws;
}

$(document).ready(function() {
    var socket = WebSocketHandler();
});