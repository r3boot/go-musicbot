package webapi

import (
	"fmt"
	"net/http"
	"os"
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

	fmt.Println(logline)
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

	fmt.Println(logline)
}

func (api *WebApi) Setup() error {
	fs, err := os.Stat(api.config.Api.Assets)
	if err != nil {
		return fmt.Errorf("WebApi.Setup: unable to load media: %v", err)
	}

	if !fs.IsDir() {
		return fmt.Errorf("WebApi.Setup: %s: not a directory", fs.Name())
	}

	return nil
}

func (api *WebApi) Run() {
	url := fmt.Sprintf("%s:%s", api.config.Api.Address, api.config.Api.Port)

	http.Handle("/css/", logHandler(http.FileServer(http.Dir(api.config.Api.Assets))))
	http.Handle("/img/", logHandler(http.FileServer(http.Dir(api.config.Api.Assets))))
	http.Handle("/js/", logHandler(http.FileServer(http.Dir(api.config.Api.Assets))))

	http.HandleFunc("/playlist", api.PlaylistHandler)
	http.HandleFunc("/queue", api.PlayQueueHandler)
	http.HandleFunc("/ta", api.AutoCompleteHandler)
	http.HandleFunc("/ws", api.SocketHandler)
	http.HandleFunc("/", api.HomeHandler)

	fmt.Printf("Listening on http://%s\n", url)
	err := http.ListenAndServe(url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run server: %v", err)
	}

}
