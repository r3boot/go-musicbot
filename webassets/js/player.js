const WS_GET_PLAYLIST = 0;
const WS_GET_ARTISTS = 1;
const WS_NEXT = 2;
const WS_BOO = 3;
const WS_TUNE = 4;
const WS_NOWPLAYING = 5;
const WS_REQUEST = 6;

const MAX_PAGES = 10;

const MAX_NOTIF_DISPLAY_TIME = 10 * 1000;
const NOTIF_INFO = 0;
const NOTIF_WARNING = 1;
const NOTIF_ERROR = 2;

const ALERT_INFO    = "alert-info";
const ALERT_WARNING = "alert-warning";
const ALERT_ERROR   = "alert-danger";

const SORT_ARTIST = 0;
const SORT_TITLE = 1;
const SORT_DURATION = 2;
const SORT_RATING = 3;
const SORT_FILENAME = 4;

const SORT_ALPHANUM = 0;
const SORT_RANDOM = 1;

var resultsPerPage = 10;
var lastRequest = 0;
var Playlist = [];
var Artists = [];
var NowPlaying = "";
var oldNumQueued = 0;
var numQueued = 0;
var selectedArtist = "";
var sortColumn = SORT_FILENAME;
var randomMode = false;

var ArtistsViewData = [];

var isPlaying = false;
var streamSrc = "";

var pgPage = 0;
var pgIndex = 0;

var socket = null;

function calcResultsPerPage() {
    var contentHeight = window.innerHeight;
    var controlsHeight = $("#idControls").height();

    var queueHeight = $("#idQueue").height();
    var resultsHeight = (contentHeight - (controlsHeight + queueHeight));
    resultsPerPage = Math.round((resultsHeight / 60) - 1);
}

function encodeString(plain) {
    var result = plain.replace(/\"/g, "|d|");
    result = result.replace(/\'/g, "|s|");
    result = result.replace(/\`/g, "|b|");
    result = encodeURI(result);
    result = btoa(result);
    return result;
}

function decodeString(encoded) {
    var result = atob(encoded);
    result = decodeURI(result);
    result = result.replace(/\|d\|/g, "\"");
    result = result.replace(/\|s\|/g, "\'");
    result = result.replace(/\|b\|/g, "\`");
    return result
}

function formatTitle(kv) {
    if ((kv.artist !== "") && (kv.title !== "")) {
        return kv.artist + " - " + kv.title;
    } else if (kv.title !== "") {
        return kv.title;
    } else {
        return kv.name;
    }
}

function randomizeArray(array) {
    var currentIndex = array.length, temporaryValue, randomIndex;

    while (0 !== currentIndex) {
        randomIndex = Math.floor(Math.random() * currentIndex);
        currentIndex -= 1;

        temporaryValue = array[currentIndex];
        array[currentIndex] = array[randomIndex];
        array[randomIndex] = temporaryValue;
    }

  return array;
}

function sortDictByArtist(playlist) {
    return playlist.sort(function compare(kv1, kv2) {
        if (kv1.artist < kv2.artist) {
            return -1;
        } else if (kv1.artist > kv2.artist) {
            return 1;
        } else {
            return 0;
        }
    });
}

function sortDictByTitle(playlist) {
    return playlist.sort(function compare(kv1, kv2) {
        if (kv1.title < kv2.title) {
            return -1;
        } else if (kv1.title > kv2.title) {
            return 1;
        } else {
            return 0;
        }
    });
}

function sortDictByDuration(playlist) {
    return playlist.sort(function compare(kv1, kv2) {
        if (kv1.duration < kv2.duration) {
            return 1;
        } else if (kv1.duration > kv2.duration) {
            return -1;
        } else {
            return 0;
        }
    });
}

function sortDictByRating(playlist) {
    return playlist.sort(function compare(kv1, kv2) {
        if (kv1.rating < kv2.rating) {
            return 1;
        } else if (kv1.rating > kv2.rating) {
            return -1;
        } else {
            return 0;
        }
    });
}
function sortDictByFilename(playlist) {
    return playlist.sort(function compare(kv1, kv2) {
       if (kv1.filename < kv2.filename) {
           return -1;
       } else if (kv1.filename > kv2.filename) {
           return 1;
       } else {
           return 0;
       }
    });
}

function SortBy(filter) {
    sortColumn = filter;
    RefreshPlaylist();
}

function pgFilterObjects(data) {
    return data.slice(pgIndex, pgIndex + resultsPerPage);
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

    pgIndex = pgPage * resultsPerPage;

    RefreshPlaylist();
}

function navItem(p) {
    var pDec = Math.floor(p);
    if (pDec === pgPage) {
        return "<li class='desktop-only page-item active'><a class='page-link' href='#'>" + pDec + "</a></li>";
    } else {
        return "<li class='desktop-only page-item'><a class='page-link' onclick='pgGotoPage(" + pDec + ")' href='#'>" + p + "</a></li>";
    }
}

function pgShowPagination(numItems) {
    var maxPages = Math.floor(numItems / resultsPerPage)-1;
    var pages = [];

    if (numItems <= resultsPerPage) {
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

    /*
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
    */

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
            if (val.toString().startsWith("\'")) {
                return true;
            }
            items.push("<li class='nav-item'><a class='nav-link' onclick='LookupTracksForArtist(\"" + encodeString(val) + "\", true)'>" + val + "</a></li>");
        });

        $("#Artists").html(items.join(""));
    }
}

function FillPlaylistResults() {
    if ($("#ArtistFilter").val() === "") {
        ArtistsViewData = Playlist.sort();
    }

    if (!randomMode) {
        switch (sortColumn) {
            case SORT_ARTIST:
                ArtistsViewData = sortDictByArtist(ArtistsViewData);
                break;
            case SORT_TITLE:
                ArtistsViewData = sortDictByTitle(ArtistsViewData);
                break;
            case SORT_DURATION:
                ArtistsViewData = sortDictByDuration(ArtistsViewData);
                break;
            case SORT_RATING:
                ArtistsViewData = sortDictByRating(ArtistsViewData);
                break;
            default:
                ArtistsViewData = sortDictByFilename(ArtistsViewData);
        }
    } else {
        ArtistsViewData = randomizeArray(Playlist);
    }

    calcResultsPerPage();

    var queryItems = pgFilterObjects(ArtistsViewData);
    var items = [];
    $.each(queryItems, function (key, val) {
        var query = val.name;

        items.push("<tr><td><span class='glyphicon glyphicon-shopping-cart' onclick='RequestTrack(\"" + encodeString(query) + "\")'></span></td><td>" + val.artist + "</td><td>" + formatTitle(val) + "</td><td>" + prettyDuration(val.duration) + "</td><td>" + val.rating + "/10</td></tr>");
    });
    $("#ArtistResults").html(items.join(""));

    pgShowPagination(ArtistsViewData.length);
}

function LookupTracksForArtist(artist, encoded) {
    var lookupArtist = artist;

    if (encoded) {
        lookupArtist = decodeString(artist).toUpperCase();
    }

    selectedArtist = lookupArtist;

    var foundArtistsData = [];
    $.each(Playlist, function (key, val) {
        if (val.artist === "") {
            if ((val.title) && (val.title !== "")) {
                if (val.title.toUpperCase().indexOf(lookupArtist) > -1) {
                    foundArtistsData.push(val);
                }
            } else {
                if (val.filename.toUpperCase().indexOf(lookupArtist) > -1) {
                    foundArtistsData.push(val);
                }
            }
        } else {
            if (val.artist.toUpperCase().indexOf(lookupArtist) > -1) {
                foundArtistsData.push(val);
            }
        }
    });

    switch (sortColumn) {
        case SORT_ARTIST:
            foundArtistsData = sortDictByArtist(foundArtistsData);
            break;
        case SORT_TITLE:
            foundArtistsData = sortDictByTitle(foundArtistsData);
            break;
        case SORT_DURATION:
            foundArtistsData = sortDictByDuration(foundArtistsData);
            break;
        case SORT_RATING:
            foundArtistsData = sortDictByRating(foundArtistsData);
            break;
        default:
            foundArtistsData = sortDictByFilename(foundArtistsData);
    }


    calcResultsPerPage();

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
        if ((val.title) && (val.title !== "")) {
            foundArtists.push("<tr><td><span class='glyphicon glyphicon-shopping-cart' onclick='RequestTrack(\"" + encodeString(query) + "\")'></span></td><td class='artist'>" + val.artist + "</td><td class='title'>" + val.title + "</td><td>" + prettyDuration(val.duration) + "</td><td>" + val.rating + "/10</td></tr>");
        } else {
            foundArtists.push("<tr><td><span class='glyphicon glyphicon-shopping-cart' onclick='RequestTrack(\"" + encodeString(query) + "\")'></span></td><td class='artist'>" + val.artist + "</td><td class='title'>" + val.filename + "</td><td>" + prettyDuration(val.duration) + "</td><td>" + val.rating + "/10</td></tr>");
        }
    });

    $("#ArtistResults").html(foundArtists.join(""));

    ArtistsViewData = foundArtistsData;

    pgShowPagination(foundArtistsData.length);
}

function RequestTrack(rawQuery) {
    var query = decodeString(rawQuery);

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
    numQueued = 0;
    queueTable.push("<h4 class='ident'>Upcoming requests</h4>");
    queueTable.push("<table class='table table-striped table-condensed'><tbody>");
    $.each(data.RequestQueue, function (key, val) {
        if (val.artist !== "") {
            queueTable.push("<tr><td>" + key + "</td><td>" + val.artist + " - " + val.title + "</td></tr>");
        } else {
            queueTable.push("<tr><td>" + key + "</td><td>" + val.title + "</td></tr>");
        }

        numQueued += 1;
    });
    queueTable.push("</tbody></table>");

    if (numQueued > 0) {
        $("#PlayQueue").html(queueTable.join(""));
    } else {
        $("#PlayQueue").html("<h4 class='ident'>Queue is empty</h4>");
    }

    if (oldNumQueued !== numQueued) {
        RefreshPlaylist();
        oldNumQueued = numQueued
    }
}

function RefreshPlaylist() {
    if (selectedArtist !== "") {
        LookupTracksForArtist(selectedArtist, false)
    } else {
        FillPlaylistResults();
    }
}

function PlaylistViewMode(mode) {
    switch (mode) {
        case SORT_ALPHANUM:
            $("#idRandomControls").addClass("hidden");
            $("#idPaginationControls").removeClass("hidden");
            randomMode = false;
            break;
        case SORT_RANDOM:
            $("#idPaginationControls").addClass("hidden");
            $("#idRandomControls").removeClass("hidden");
            randomMode = true;
            break;
    }
    RefreshPlaylist();
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
            var nowPlayingRequest = {"i": lastRequest++, "o": WS_NOWPLAYING};
            setInterval(function () {
                ws.send(JSON.stringify(nowPlayingRequest));
            }, 1000);

            var playlistRequest = {"i": lastRequest++, "o": WS_GET_PLAYLIST};
            setInterval(function () {
                ws.send(JSON.stringify(playlistRequest));
            }, 5000);

            var artistsRequest = {"i": lastRequest++, "o": WS_GET_ARTISTS};
            setInterval(function () {
                ws.send(JSON.stringify(artistsRequest));
            }, 5000);

            if ($("#AlertMessage").html() === "Websocket connection closed") {
                HideNotification();
            }
        };

        ws.onmessage = function (e) {
            var response = JSON.parse(e.data);

            switch (response.o) {
                case WS_GET_PLAYLIST:
                    if (response.s === true) {
                        var updatePlaylist = false;
                        if (Playlist.length === 0) {
                            updatePlaylist = true;
                        }

                        Playlist = randomizeArray(response.d);
                        if (updatePlaylist) {
                            RefreshPlaylist();
                        }
                    } else {
                        ShowNotification(NOTIF_WARNING, response.m);
                    }
                    break;
                case WS_GET_ARTISTS:
                    if (response.s === true) {
                        var updateArtists = false;
                        if (Artists.length === 0) {
                            updateArtists = true;
                        }
                        Artists = response.d;
                        if (updateArtists) {
                            UpdateArtists();
                        }
                    } else {
                        ShowNotification(NOTIF_WARNING, response.m);
                    }
                    break;
                case WS_NEXT:
                    if (response.s !== true) {
                        ShowNotification(NOTIF_WARNING, response.m);
                    }
                    break;
                case WS_NOWPLAYING:
                    if (response.s === true) {
                        UpdateNowPlaying(response.d);
                    } else {
                        ShowNotification(NOTIF_WARNING, response.m);
                    }
                    break;
                case WS_REQUEST:
                    if (response.s === true) {
                        var Track = response.d;

                        if (Track === null) {
                            var msg = "Got no result from server";
                            ShowNotification(NOTIF_WARNING, msg);
                        } else {
                            var title = "";
                            if ((Track.artist !== "") && (Track.title !== "")) {
                                title = Track.artist + " - " + Track.title;
                            } else if (Track.title !== "") {
                                title = Track.title;
                            } else {
                                title = Track.filename;
                            }

                            var msg = "Added " + title + " to the queue at position " + Track.prio;
                            ShowNotification(NOTIF_INFO, msg);
                        }
                    } else {
                        ShowNotification(NOTIF_WARNING, response.m);
                    }
                    break;
                default:
                    console.log("Received unknown websocket packet: " + response.o);
            }
        };

        ws.onclose = function () {
            ShowNotification(NOTIF_ERROR, "Websocket connection closed");
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
                ShowNotification(NOTIF_ERROR, "Failed to start stream: " + error);
            });
        }

        $("#idPlay").html("<span class='glyphicon glyphicon-pause'></span>");
        isPlaying = true;
    }
}

function ShowNotification(type, message) {
    if (!$("#Alert").hasClass("hidden")) {
        HideNotification();
    }

    $("#AlertMessage").html(message);

    switch (type) {
        case NOTIF_INFO:
            $("#Alert").addClass(ALERT_INFO);
            break;
        case NOTIF_WARNING:
            $("#Alert").addClass(ALERT_WARNING);
            break;
        case NOTIF_ERROR:
            $("#Alert").addClass(ALERT_ERROR);
    }

    $("#Alert").removeClass("hidden");

    if (type !== NOTIF_ERROR) {
        setTimeout(HideNotification, MAX_NOTIF_DISPLAY_TIME);
    }
}

function HideNotification() {
    $("#Alert").addClass("hidden");

    $("#AlertMessage").html("");

    var allClasses = $('#Alert').attr("class").split(' ');

    $.each(allClasses, function(idx, value) {
        switch (value) {
            case ALERT_INFO:
                $("#Alert").removeClass(ALERT_INFO);
                break;
            case ALERT_WARNING:
                $("#Alert").removeClass(ALERT_WARNING);
                break;
            case ALERT_ERROR:
                $("#Alert").removeClass(ALERT_ERROR);
                break;
        }
    });
}

function runWebsite() {
    $(document).ready(function () {
        var audioControls = document.getElementById("audioControls");
        HideNotification();
        PlaylistViewMode(SORT_ALPHANUM);

        socket = WebSocketMuxer();
        ToggleStream();

        $("#PlayerView").hide();

        ShowView("#PlayerView");

        $("#ShowPlayerView").click(function (ev) {
            ev.preventDefault();
            RefreshPlaylist();
            ShowView("#PlayerView");
        });

        $("#idPlaylistMode").on("change", function(ev) {
            ev.preventDefault();

            var mode = $("#idPlaylistMode").prop("checked");

            if (mode) {
                PlaylistViewMode(SORT_ALPHANUM);
            } else {
                PlaylistViewMode(SORT_RANDOM);
            }
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
            }
        });

        $('#volSlider').slider({
            value: (audioControls.volume * 100)
        });

        $('#volSlider').on("change", function (ev) {
            var newVolume = Math.round((ev.value.newValue / 100) * 10) / 10;
            audioControls.volume = newVolume;
        });

        $('input[type=search]').on('keyup', function() {
            FilterArtists();
        });

        $('input[type=search]').on('search', function () {
            FilterArtists();
            if ($("#ArtistFilter").val() === "") {
                selectedArtist = "";
                RefreshPlaylist();
            }
        });

        $(window).resize(function() {
            RefreshPlaylist();
        })
    });
}

runWebsite();