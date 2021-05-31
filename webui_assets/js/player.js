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

const sourceUrl = "/stream/2600nl.mp3";

// Leave this global
var ws = null;
var playing = false;
var firstPlaying = false;

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

function TogglePlayPause() {
    var player = document.getElementById("audioControls");
    if (playing) {
        player.pause();
        player.src = "";
        player.currentTime = 0;
        playing = false;
        $("#btnPlay").html("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"24\" height=\"24\" fill=\"currentColor\" className=\"bi bi-play-fill\" viewBox=\"0 0 16 16\"><path d=\"m11.596 8.697-6.363 3.692c-.54.313-1.233-.066-1.233-.697V4.308c0-.63.692-1.01 1.233-.696l6.363 3.692a.802.802 0 0 1 0 1.393z\"/></svg>\n");
    } else {
        player.src = sourceUrl;
        player.load();
        player.play();
        playing = true;
        $("#btnPlay").html("<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"24\" height=\"24\" fill=\"currentColor\" className=\"bi bi-pause-fill\" viewBox=\"0 0 16 16\"><path d=\"M5.5 3.5A1.5 1.5 0 0 1 7 5v6a1.5 1.5 0 0 1-3 0V5a1.5 1.5 0 0 1 1.5-1.5zm5 0A1.5 1.5 0 0 1 12 5v6a1.5 1.5 0 0 1-3 0V5a1.5 1.5 0 0 1 1.5-1.5z\"/></svg>\n");
    }
}


function UpdateNowPlaying(data) {
    var track = filenameToTrack(data.Filename);

    $("#idNowPlaying").html(track);
    $("#idRating").html(data.Rating + "/10");
    $("#idDuration").html(prettyDuration(data.Duration));

    var newAlbumArt = "";
    if (data.AlbumArt == "notfound.png") {
        newAlbumArt = "/img/" + data.AlbumArt;
    } else {
        newAlbumArt = "/art/" + data.AlbumArt;
    }

    if ($("#idAlbumArt").attr("src") !== newAlbumArt) {
        $("#idAlbumArt").attr("src", newAlbumArt);
    }

    document.getElementById("idElapsedSlider").setAttribute("max", data.Duration);
    document.getElementById("idElapsedSlider").setAttribute("value", data.Elapsed);
}

function RequestTrack(parent) {
    var yid = parent.id;

    wsSendCommandWithData(ws, wsRequestTrack, yid);

    ClearSearch();
}

function ClearSearch() {
    document.getElementById("inputSearch").value = "";
    UpdateSearchResults([]);
}

function UpdateSearchResults(data) {
    var searchResults = "";
    var yid = "";
    if ((data != null) && data.length > 0) {
        for (var i = 0; i < data.length; i++) {
            yid = data[i].substring(data[i].length - 15, data[i].length - 4);
            searchResults += "<span id=\"" + yid + "\" onClick=\"RequestTrack(this)\"><svg xmlns=\"http://www.w3.org/2000/svg\" width=\"16\" height=\"16\" fill=\"currentColor\" class=\"bi bi-file-earmark-play\" viewBox=\"0 0 16 16\">\n" +
                "  <path d=\"M6 6.883v4.234a.5.5 0 0 0 .757.429l3.528-2.117a.5.5 0 0 0 0-.858L6.757 6.454a.5.5 0 0 0-.757.43z\"/>\n" +
                "  <path d=\"M14 14V4.5L9.5 0H4a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2zM9.5 3A1.5 1.5 0 0 0 11 4.5h2V14a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1h5.5v2z\"/>\n" +
                "</svg></span> " + filenameToTrack(data[i]) + "<br/>";
        }
    }
    $("#divSearchResults").html(searchResults);
}

function UpdateQueueEntries(data) {
    var queueEntries = "";

    queueEntries = "<ol>";
    for (const [idx, filename] of Object.entries(data)) {
        queueEntries += "<li>" + filenameToTrack(filename) + "</li>";
    }
    queueEntries += "</ol>";
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

        $("#inputSearch").keyup(function (ev) {
            if (ev.keyCode == 13) {
                var query = document.getElementById("inputSearch").value;
                console.log("searching for " + query)
                if ((query != null) && (query.length > 0)) {
                    wsSendCommandWithData(ws, wsRequestQuery, query);
                } else {
                    ClearSearch();
                    console.log("No query")
                }
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

    $("#btnPlay").click(function(ev) {
        ev.preventDefault();
        TogglePlayPause();
    });

    $("#idVolumeSlider").on("input", function (ev) {
        var newVolume = Math.round((ev.currentTarget.value / 100) * 10) / 10;
        $("#audioControls").prop("volume", newVolume);
    });
});
