const WS_GET_PLAYLIST = 0;
const WS_GET_ARTISTS = 1;
const WS_NEXT = 2;
const WS_BOO = 3;
const WS_TUNE = 4;
const WS_NOWPLAYING = 5;
const WS_REQUEST = 6;

const RESULTS_PER_PAGE = 10;
const MAX_PAGES = 10;

var lastRequest = 0;
var Playlist = [];
var Artists = [];
var NowPlaying = "";

var ArtistsViewData = [];

var isPlaying = false;
var streamSrc = "";

var pgPage = 0;
var pgIndex = 0;

var socket = null;

function toInt(value) {
    tmp = value.toString();
    return tmp.substring(0, tmp.indexOf("."));
}

function pgFilterObjects(data) {
    return data.slice(pgIndex, pgIndex + RESULTS_PER_PAGE);
}

function pgGotoPage(page) {
    switch (page) {
        case -2:
            pgPage += 1;
            break;
        case -1:
            if (pgPage > 0) {
                pgPage -= 1;
            }
            break;
        default:
            pgPage = page;
    }

    pgIndex = pgPage * RESULTS_PER_PAGE;

    FillPlaylistResults();
}

function navItem(p) {
    var pDec = Math.floor(p);
    if (pDec === pgPage) {
        return "<li class='page-item active'><a class='page-link' href='#'>" + pDec + "</a></li>";
    } else {
        return "<li class='page-item'><a class='page-link' onclick='pgGotoPage(" + pDec + ")' href='#'>" + p + "</a></li>";
    }
}

function pgShowPagination(numItems) {
    var maxPages = numItems / RESULTS_PER_PAGE;
    var pages = [];

    if (numItems <= RESULTS_PER_PAGE) {
        $("#ArtistPagination").html("");
        return;
    }

    if (pgPage === 0) {
        pages.push("<li class='page-item disabled'><a class='page-link' href=#' tabindex=-1'>first</a>");
        pages.push("<li class='page-item disabled'><a class='page-link' href=#' tabindex=-1'>back</a>");
    } else {
        pages.push("<li class='page-item'><a class='page-link' onclick='pgGotoPage(0)' href=#' tabindex=-1'>first</a>");
        pages.push("<li class='page-item'><a class='page-link' onclick='pgGotoPage(-1)' href=#' tabindex=-1'>back</a>");
    }

    if (maxPages > MAX_PAGES) {
        if (pgPage < 5) {
            for (p = 0; p < MAX_PAGES; p++) {
                pages.push(navItem(p));
            }
        } else if ((maxPages - pgPage) < 5) {
            for (p = maxPages - 10; p < maxPages; p++) {
                pages.push(navItem(p));
            }
        } else {
            for (p = pgPage - 5; p < pgPage + 5; p++) {
                pages.push(navItem(p));
            }
        }
    } else {
        for (p = 0; p <= maxPages; p++) {
            pages.push(navItem(p));
        }
    }

    if (pgPage == maxPages) {
        pages.push("<li class='page-item disabled'><a class='page-link' href=#' tabindex=-1'>next</a>");
        pages.push("<li class='page-item disabled'><a class='page-link' href=#' tabindex=-1>last</a>");
    } else {
        pages.push("<li class='page-item'><a class='page-link' onclick='pgGotoPage(-2)' href='#' tabindex=-1'>next</a>");
        pages.push("<li class='page-item'><a class='page-link' onclick='pgGotoPage(" + maxPages + ")' href=#' tabindex=-1'>last</a>");
    }

    $("#ArtistPagination").html(pages.join(""))
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

function UpdateArtists() {
    if ($("#ArtistFilter").val() === "") {
        var items = [];
        $.each(Artists, function (key, val) {
            items.push("<li class='nav-item'><a class='nav-link' href='#' onclick='LookupTracksForArtist(\"" + atob(val) + "\")'>" + val + "</a></li>");
        });

        $("#Artists").html(items.join(""));
    }
}

function FillPlaylistResults() {
    if ($("#ArtistFilter").val() === "") {
        ArtistsViewData = Playlist;
    }

    var queryItems = pgFilterObjects(ArtistsViewData);
    var items = [];
    $.each(queryItems, function (key, val) {
        var query = "";
        if (val.artist !== "") {
            query = val.artist + " - " + val.title;
        } else {
            query = val.title;
        }

        items.push("<tr><td>" + val.artist + "</td><td>" + val.title + "</td><td>" + prettyDuration(val.duration) + "</td><td>" + val.rating + "/10</td><td><span class='glyphicon glyphicon-shopping-cart' onclick='RequestTrack(\"" + query + "\")'></span></td></tr>");
    });
    $("#ArtistResults").html(items.join(""));

    pgShowPagination(ArtistsViewData.length);
}

function LookupTracksForArtist(artist) {
    var lookupArtist = btoa(artist).toUpperCase();

    var foundArtistsData = [];
    $.each(Playlist, function (key, val) {
        if (val.artist === "") {
            if (val.title.toUpperCase() === lookupArtist) {
                foundArtistsData.push(val);
            }
        } else if (val.artist.toUpperCase() === lookupArtist) {
            foundArtistsData.push(val);
        }
    });

    pgPage = 0;
    pgIndex = 0;

    var queryItems = pgFilterObjects(foundArtistsData);

    var foundArtists = [];
    $.each(queryItems, function (key, val) {
        var query = "";
        if (val.artist !== "") {
            query = val.artist + " - " + val.title;
        } else {
            query = val.title;
        }
        foundArtists.push("<tr><td>" + val.artist + "</td><td>" + val.title + "</td><td>" + val.rating + "/10, " + prettyDuration(val.duration) + "</td><td><span class='glyphicon glyphicon-shopping-cart' onclick='RequestTrack(\"" + query + "\")'></span></td></tr>");
    });

    $("#ArtistResults").html(foundArtists.join(""));

    ArtistsViewData = foundArtistsData;

    pgShowPagination(foundArtistsData.length);
}

function RequestTrack(query) {
    var request = {"i": lastRequest++, "o": WS_REQUEST, "d": encodeURI(query)};
    socket.send(JSON.stringify(request));
    return true;
}

function UpdateNowPlaying(data) {
    if (data.Artist) {
        NowPlaying = data.Artist + " - " + data.Title;
    } else {
        NowPlaying = data.Title;
    }

    if ($("#idAlbumArt").attr("src") !== data.ImageUrl) {
        $("#idAlbumArt").attr("src", data.ImageUrl);
    }

    $("#idNowPlaying").html(NowPlaying);
    $("#idRating").html(data.Rating + "/10");
    $("#idDuration").html(prettyDuration(data.Duration));

    var curDuration = $("#trackSlider").slider("getAttribute", "max");
    if (curDuration !== data.Duration) {
        $("#trackSlider").slider("setAttribute", "max", data.Duration)
    }
    $("#trackSlider").slider("setValue", data.Elapsed);

    var queueTable = [];
    var nItems = 0;

    queueTable.push("<h4>Upcoming requests</h4>");
    queueTable.push("<table class='table table-striped table-condensed'><tbody>");
    $.each(data.RequestQueue, function (key, val) {
        queueTable.push("<tr><td>" + key + "</td><td>" + val + "</td></tr>");
        nItems += 1;
    });
    queueTable.push("</tbody></table>");

    if (nItems > 0) {
        $("#PlayQueue").html(queueTable.join(""));
    } else {
        $("#PlayQueue").html("<h4>Queue is empty</h4>");
    }
}

function ShowView(viewId) {
    allViews = ["#PlayerView"];
    allButtons = ["#NavPlayerView"];

    for (i = 0; i < allViews.length; i++) {
        if (allViews[i] == viewId) {
            continue;
        }
        $(allViews[i]).hide();
        $(allViews[i]).removeClass("active");
    }

    for (i = 0; i < allButtons.length; i++) {
        if (allViews[i] == viewId) {
            continue;
        }
        $(allButtons[i]).removeClass("active");
    }

    $(viewId).show();
    $(viewId).addClass("active");

    switch (viewId) {
        case "#PlayerView":
            $("#NavPlayerView").addClass("active");
            break;
    }
}

function WebSocketMuxer() {
    var wsProto = "wss:";

    if (location.protocol === "http:") {
        wsProto = "ws:";
    }

    var ws = null;

    function start() {
        ws = new WebSocket(wsProto + "//" + window.location.host + "/ws");

        $("#idNext").click(function (ev) {
            var nextRequest = {"i": lastRequest++, "o": WS_NEXT};
            ws.send(JSON.stringify(nextRequest));
        });

        $("#idBoo").click(function (ev) {
            var booRequest = {"i": lastRequest++, "o": WS_BOO};
            ws.send(JSON.stringify(booRequest));
        });

        $("#idTune").click(function (ev) {
            var tuneRequest = {"i": lastRequest++, "o": WS_TUNE};
            ws.send(JSON.stringify(tuneRequest));
        });

        ws.onopen = function () {
            var playlistRequest = {"i": lastRequest++, "o": WS_GET_PLAYLIST};
            setInterval(function () {
                ws.send(JSON.stringify(playlistRequest));
            }, 1000);

            var artistsRequest = {"i": lastRequest++, "o": WS_GET_ARTISTS};
            setInterval(function () {
                ws.send(JSON.stringify(artistsRequest));
            }, 1000);

            var nowPlayingRequest = {"i": lastRequest++, "o": WS_NOWPLAYING};
            setInterval(function () {
                ws.send(JSON.stringify(nowPlayingRequest));
            }, 1000);
        };

        ws.onmessage = function (e) {
            var response = JSON.parse(e.data);

            switch (response.o) {
                case WS_GET_PLAYLIST:
                    var updatePlaylist = false;
                    if (Playlist.length === 0) {
                        updatePlaylist = true;
                    }

                    Playlist = response.d;
                    if (updatePlaylist) {
                        FillPlaylistResults();
                    }
                    break;
                case WS_GET_ARTISTS:
                    var updateArtists = false;
                    if (Artists.length === 0) {
                        updateArtists = true;
                    }
                    Artists = response.d;
                    if (updateArtists) {
                        UpdateArtists();
                    }
                    break;
                case WS_NEXT:
                    break;
                case WS_NOWPLAYING:
                    UpdateNowPlaying(response.d);
                    break;
                case WS_REQUEST:
                    break;
                default:
                    console.log("Received unknown websocket packet: " + response.o);
            }
        };

        ws.onclose = function () {
            check();
        };
    }

    function check() {
        if (!ws || ws.readyState == 3) start();
    }

    start();

    setInterval(check, 5000);

    return ws;
}

function ToggleStream() {
    var player = document.getElementById("audioControls");
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
        $("#idPlay").html("<span class='glyphicon glyphicon-play'></span>");
        isPlaying = false;
    } else {
        source.src = streamSrc;
        player.load();

        var promise = player.play();
        if (promise !== undefined) {
            promise.then(function () {
            }).catch(function (error) {
                console.log("Failed to open stream: " + error);
            });
        }

        $("#idPlay").html("<span class='glyphicon glyphicon-pause'></span>");
        isPlaying = true;
    }
}

function runWebsite() {
    $(document).ready(function () {
        var audioControls = document.getElementById("audioControls");

        socket = WebSocketMuxer();
        ToggleStream();

        $("#PlayerView").hide();

        ShowView("#PlayerView");

        $("#ShowPlayerView").click(function (ev) {
            ev.preventDefault();
            FillPlaylistResults();
            ShowView("#PlayerView");
        });

        $("#idPlay").click(function (ev) {
            ev.preventDefault();
            ToggleStream();
        });

        $('#trackSlider').slider({
            value: 0,
            enabled: false,
            formatter: function (value) {
                return prettyDuration(value);
            },
        });

        $('#volSlider').slider({
            value: (audioControls.volume * 100)
        });

        $('#volSlider').on("change", function (ev) {
            var newVolume = Math.round((ev.value.newValue / 100) * 10) / 10;
            audioControls.volume = newVolume;
        });
    });
}

runWebsite();