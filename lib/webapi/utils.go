package webapi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func httpLog(r *http.Request, code int, size int) {
	var (
		srcip   string
		logline string
	)

	srcip = r.Header.Get("X-Forwarded-For")
	if srcip == "" {
		srcip = r.RemoteAddr
	}

	logline = srcip + " - - [" + time.Now().Format(TF_CLF) + "] "
	logline += "\"" + r.Method + " " + r.URL.Path + " " + r.Proto + "\" "
	logline += strconv.Itoa(code) + " " + strconv.Itoa(size)

	fmt.Printf("%s\n", logline)
}

func okResponse(r *http.Request, size int) {
	httpLog(r, http.StatusOK, size)
}

func errorResponse(w http.ResponseWriter, r *http.Request, errmsg string) {
	response := WebResponse{
		Status:  false,
		Message: errmsg,
	}

	data, err := response.ToJSON()
	if err != nil {
		log.Warningf("WebAPI.HomeHandler: %v", err)
		errmsg := "{\"Status\":false,\"Message\":\"json encoding failed\"}"
		http.Error(w, errmsg, http.StatusInternalServerError)
		httpLog(r, http.StatusInternalServerError, len(errmsg))
		return
	}

	http.Error(w, string(data), http.StatusInternalServerError)
	httpLog(r, http.StatusInternalServerError, len(data))
}

func wsLog(r *http.Request, code int, cmd, msg string) {
	var (
		srcip   string
		logline string
	)

	srcip = r.Header.Get("X-Forwarded-For")
	if srcip == "" {
		srcip = r.RemoteAddr
	}

	logline = srcip + " [" + time.Now().Format(TF_CLF) + "] "
	logline += strconv.Itoa(code) + " " + cmd + ": " + msg

	fmt.Printf("%s\n", logline)
}

func wsOkResponse(r *http.Request, cmd, msg string) {
	wsLog(r, http.StatusOK, cmd, msg)
}

func wsErrorResponse(r *http.Request, cmd, msg string) {
	wsLog(r, http.StatusInternalServerError, cmd, msg)
}
