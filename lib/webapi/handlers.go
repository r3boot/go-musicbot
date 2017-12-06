package webapi

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"net/http"
)

func logHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpLog(r, http.StatusOK, 0)
		h.ServeHTTP(w, r)
	})
}

func (a *WebAPI) HomeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		{
			content, err := ioutil.ReadFile(a.assets + "/templates/player.html")
			if err != nil {
				log.Warningf("WebAPI.HomeHandler ioutil.ReadFile: %v", err)
				errorResponse(w, r, "Failed to read template")
				return
			}

			t := template.New("index")

			_, err = t.Parse(string(content))
			if err != nil {
				log.Warningf("WebAPI.HomeHandler t.Parse: %v", err)
				errorResponse(w, r, "Failed to parse template")
				return
			}

			data := IndexTemplateData{
				Title:     "2600nl radio",
				Mp3Stream: a.Config.Api.Mp3StreamURL,
				OggStream: a.Config.Api.OggStreamURL,
			}

			output := bytes.Buffer{}

			err = t.Execute(&output, data)
			if err != nil {
				log.Warningf("WebAPI.HomeHandler t.Execute: %v", err)
				errmsg := "Failed to execute template"
				http.Error(w, errmsg, http.StatusInternalServerError)
				httpLog(r, http.StatusInternalServerError, len(errmsg))
				return
			}

			w.Write(output.Bytes())
			httpLog(r, http.StatusOK, output.Len())
		}
	default:
		{
			errorResponse(w, r, "Unsupported method")
		}
	}
}

func (a *WebAPI) PlaylistHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		{
			if a.Playlist == nil {
				log.Warningf("WebAPI.PlaylistHandler: Playlist is empty")
				errorResponse(w, r, "Playlist is empty")
				return
			}

			response := WebResponse{
				Status: true,
				Data:   a.Playlist,
			}

			data, err := response.ToJSON()
			if err != nil {
				log.Warningf("WebAPI.PlaylistHandler: %v", err)
				errorResponse(w, r, "json encoding failed")
				return
			}

			w.Write(data)
			okResponse(r, len(data))
		}
	default:
		{
			errorResponse(w, r, "Unsupported method")
		}
	}
}

func (a *WebAPI) ArtistHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		{
			if a.Artists == nil {
				log.Warningf("WebAPI.ArtistHandler: Artists not loaded yet")
				errorResponse(w, r, "Artists is empty")
				return
			}

			response := WebResponse{
				Status: true,
				Data:   a.Artists,
			}

			data, err := response.ToJSON()
			if err != nil {
				log.Warningf("WebAPI.ArtistHandler: %v", err)
				errorResponse(w, r, "json encoding failed")
				return
			}

			w.Write(data)
			okResponse(r, len(data))
		}
	case http.MethodPost:
		{
			r.ParseForm()

			for key, value := range r.Form {
				if key != "q" {
					continue
				}

				q := value[0]

				entries, err := a.mpdClient.TracksForArtist(q)
				if err != nil {
					log.Warningf("WebAPI.ArtistHandler: %v", err)
					errorResponse(w, r, "failed to lookup artists")
					return
				}

				response := WebResponse{
					Status: true,
					Data:   entries,
				}

				data, err := response.ToJSON()
				if err != nil {
					log.Warningf("WebAPI.ArtistHandler: %v", err)
					errorResponse(w, r, "json encoding failed")
					return
				}

				w.Write(data)
				okResponse(r, len(data))
				return
			}
			errorResponse(w, r, "received invalid form")
		}
	default:
		{
			errorResponse(w, r, "Unsupported method")
		}
	}
}
