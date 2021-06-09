// This file is safe to edit. Once it exists it will not be overwritten

package apiserver

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/r3boot/go-musicbot/lib/config"
	"github.com/r3boot/go-musicbot/lib/log"
	"github.com/r3boot/go-musicbot/lib/manager"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/r3boot/go-musicbot/lib/apiserver/operations"
)

//go:generate swagger generate server --target ../../../test --name Musicbot --spec ../../swagger.yaml --server-package ./lib/apiserver --principal interface{} --exclude-main

var (
	Config *config.Config
	mgr    *manager.Manager
)

func configureFlags(api *operations.MusicbotAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func validateToken(token string) (*config.ApiUser, error) {
	for _, userInfo := range Config.WebApi.Users {
		if userInfo.Token != token {
			continue
		}
		return &userInfo, nil
	}

	return nil, fmt.Errorf("authentication failed")
}

func isAuthorized(principal interface{}, role string) error {
	for _, authorization := range principal.(*config.ApiUser).Authorizations {
		if authorization == role {
			return nil
		}
	}
	return fmt.Errorf("%s is not authorized for %s", principal.(*config.ApiUser).Name, role)
}

func configureAPI(api *operations.MusicbotAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.Logger = func(msg string, args ...interface{}) {
		log.Infof(log.Fields{
			"package": "apiserver",
		}, msg, args...)
	}

	if Config == nil {
		log.Fatalf(log.Fields{
			"package":  "apiserver",
			"function": "configureAPI",
		}, "no configuration found")
	}

	// Setup Manager
	mgr, err := manager.NewManager(Config)
	if err != nil {
		log.Fatalf(log.Fields{
			"package":  "apiserver",
			"function": "configureAPI",
			"call":     "manager.NewManager",
		}, err.Error())
	}

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()
	api.UrlformConsumer = runtime.DiscardConsumer

	api.JSONProducer = runtime.JSONProducer()

	// Applies when the "X-Access-Token" header is set
	api.AccessSecurityAuth = func(token string) (interface{}, error) {
		principal, err := validateToken(token)
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "apiserver",
				"function": "AccessSecurityAuth",
				"call":     "validateToken",
				"token":    token,
			}, err.Error())
			return nil, fmt.Errorf("authentication failed")
		}
		if principal == nil {
			log.Warningf(log.Fields{
				"package":  "apiserver",
				"function": "AccessSecurityAuth",
			}, "principal == nil")
			return nil, fmt.Errorf("authentication failed")
		}
		return principal, nil
	}

	// Set your custom authorizer if needed. Default one is security.Authorized()
	// Expected interface runtime.Authorizer
	//
	// Example:
	// api.APIAuthorizer = security.Authorized()

	api.GetPlayerNextHandler = operations.GetPlayerNextHandlerFunc(func(params operations.GetPlayerNextParams, principal interface{}) middleware.Responder {
		err := isAuthorized(principal, "allowPlayerNext")
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "apiserver",
				"function": "GetPlayerNextHandler",
				"call":     "isAuthorized",
			}, err.Error())
			return operations.NewGetPlayerNextForbidden()
		}

		err = mgr.Next()
		if err != nil {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "GetPlayerNextHandler",
				"call":      "mgr.Next",
				"principal": principal.(*config.ApiUser).Name,
			}, err.Error())
			return operations.NewGetPlayerNextBadRequest()
		}

		// Wait for liquidsoap to load the next track
		time.Sleep(500 * time.Millisecond)

		track, err := mgr.NowPlaying()
		if err != nil {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "GetPlayerNextHandler",
				"call":      "mgr.NowPlaying",
				"principal": principal.(*config.ApiUser).Name,
			}, err.Error())
			return operations.NewGetPlayerNowplayingBadRequest()
		}

		addedOnTs := track.AddedOn.String()
		duration := int64(track.Duration)

		response := operations.GetPlayerNextOKBody{
			Yid:       &track.Yid,
			Filename:  &track.Filename,
			Addedon:   &addedOnTs,
			Duration:  &duration,
			Rating:    &track.Rating,
			Submitter: &track.Submitter,
		}

		log.Debugf(log.Fields{
			"package":   "apiserver",
			"function":  "GetPlayerNextHandler",
			"principal": principal.(*config.ApiUser).Name,
			"filename":  track.Filename,
		}, "skipped to track")

		return operations.NewGetPlayerNextOK().WithPayload(&response)
	})

	api.GetPlayerNowplayingHandler = operations.GetPlayerNowplayingHandlerFunc(func(params operations.GetPlayerNowplayingParams, principal interface{}) middleware.Responder {
		err := isAuthorized(principal, "allowPlayerNowPlaying")
		if err != nil {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "GetPlayerNowplayingHandler",
				"call":      "isAuthorized",
				"principal": principal.(*config.ApiUser).Name,
			}, err.Error())
			return operations.NewGetPlayerNowplayingForbidden()
		}

		track, err := mgr.NowPlaying()
		if err != nil {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "GetPlayerNowplayingHandler",
				"call":      "mgr.NowPlaying",
				"principal": principal.(*config.ApiUser).Name,
			}, err.Error())
			return operations.NewGetPlayerNowplayingBadRequest()
		}

		addedOnTs := track.AddedOn.String()
		duration := int64(track.Duration)
		elapsed := int64(track.Elapsed / time.Second)

		response := operations.GetPlayerNowplayingOKBody{
			Yid:       &track.Yid,
			Filename:  &track.Filename,
			Addedon:   &addedOnTs,
			Duration:  &duration,
			Elapsed:   &elapsed,
			Rating:    &track.Rating,
			Submitter: &track.Submitter,
		}

		return operations.NewGetPlayerNowplayingOK().WithPayload(&response)
	})

	api.GetPlayerQueueHandler = operations.GetPlayerQueueHandlerFunc(func(params operations.GetPlayerQueueParams, principal interface{}) middleware.Responder {
		err := isAuthorized(principal, "allowPlayerQueue")
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "apiserver",
				"function": "GetPlayerQueueHandler",
				"call":     "isAuthorized",
			}, err.Error())
			return operations.NewGetPlayerQueueForbidden()
		}

		entries, err := mgr.GetQueue()
		if err != nil {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "GetPlayerQueueHandler",
				"call":      "mgr.GetQueue",
				"principal": principal.(*config.ApiUser).Name,
			}, err.Error())
			return operations.NewGetPlayerNextBadRequest()
		}

		foundTracks := make([]*operations.GetPlayerQueueOKBodyItems0, 0)

		for _, track := range entries {
			yid := track.Yid
			filename := track.Filename
			addedOnTs := track.AddedOn.String()
			duration := int64(track.Duration)
			rating := track.Rating
			submitter := track.Submitter

			responseTrack := operations.GetPlayerQueueOKBodyItems0{
				Yid:       &yid,
				Filename:  &filename,
				Addedon:   &addedOnTs,
				Duration:  &duration,
				Rating:    &rating,
				Submitter: &submitter,
			}

			foundTracks = append(foundTracks, &responseTrack)
		}

		return operations.NewGetPlayerQueueOK().WithPayload(foundTracks)
	})

	api.GetRatingDecreaseHandler = operations.GetRatingDecreaseHandlerFunc(func(params operations.GetRatingDecreaseParams, principal interface{}) middleware.Responder {
		err := isAuthorized(principal, "allowRatingDecrease")
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "apiserver",
				"function": "GetRatingDecreaseHandler",
				"call":     "isAuthorized",
			}, err.Error())
			return operations.NewGetRatingDecreaseForbidden()
		}

		err = mgr.DecreaseRating()
		if err != nil {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "GetRatingDecreaseHandler",
				"call":      "mgr.DecreaseRating",
				"principal": principal.(*config.ApiUser).Name,
			}, err.Error())
			return operations.NewGetRatingDecreaseBadRequest()
		}

		track, err := mgr.NowPlaying()
		if err != nil {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "GetRatingDecreaseHandler",
				"call":      "mgr.NowPlaying",
				"principal": principal.(*config.ApiUser).Name,
			}, err.Error())
			return operations.NewGetRatingDecreaseBadRequest()
		}

		addedOnTs := track.AddedOn.String()
		duration := int64(track.Duration)

		response := operations.GetRatingDecreaseOKBody{
			Yid:       &track.Yid,
			Filename:  &track.Filename,
			Addedon:   &addedOnTs,
			Duration:  &duration,
			Rating:    &track.Rating,
			Submitter: &track.Submitter,
		}

		log.Debugf(log.Fields{
			"package":   "apiserver",
			"function":  "GetRatingDecreaseHandler",
			"principal": principal.(*config.ApiUser).Name,
			"filename":  track.Filename,
			"rating":    track.Rating,
		}, "rating decreased")

		return operations.NewGetRatingDecreaseOK().WithPayload(&response)
	})

	api.GetRatingIncreaseHandler = operations.GetRatingIncreaseHandlerFunc(func(params operations.GetRatingIncreaseParams, principal interface{}) middleware.Responder {
		err := isAuthorized(principal, "allowRatingIncrease")
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "apiserver",
				"function": "GetRatingIncreaseHandler",
				"call":     "isAuthorized",
			}, err.Error())
			return operations.NewGetRatingIncreaseForbidden()
		}

		err = mgr.IncreaseRating()
		if err != nil {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "GetRatingIncreaseHandler",
				"call":      "mgr.IncreaseRating",
				"principal": principal.(*config.ApiUser).Name,
			}, err.Error())
			return operations.NewGetRatingIncreaseBadRequest()
		}

		track, err := mgr.NowPlaying()
		if err != nil {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "GetRatingIncreaseHandler",
				"call":      "mgr.NowPlaying",
				"principal": principal.(*config.ApiUser).Name,
			}, err.Error())
			return operations.NewGetRatingIncreaseBadRequest()
		}

		addedOnTs := track.AddedOn.String()
		duration := int64(track.Duration)

		response := operations.GetRatingIncreaseOKBody{
			Yid:       &track.Yid,
			Filename:  &track.Filename,
			Addedon:   &addedOnTs,
			Duration:  &duration,
			Rating:    &track.Rating,
			Submitter: &track.Submitter,
		}

		log.Debugf(log.Fields{
			"package":   "apiserver",
			"function":  "GetRatingIncreaseHandler",
			"principal": principal.(*config.ApiUser).Name,
			"filename":  track.Filename,
			"rating":    track.Rating,
		}, "rating increased")

		return operations.NewGetRatingIncreaseOK().WithPayload(&response)
	})

	api.PostTrackHasHandler = operations.PostTrackHasHandlerFunc(func(params operations.PostTrackHasParams, principal interface{}) middleware.Responder {
		if mgr.HasYid(*params.Body.Yid) {
			log.Debugf(log.Fields{
				"package":   "apiserver",
				"function":  "PostTrackHasHandler",
				"principal": principal.(*config.ApiUser).Name,
				"yid":       *params.Body.Yid,
				"submitter": *params.Body.Submitter,
			}, "track found in database")
			return operations.NewPostTrackHasNoContent()
		}

		log.Debugf(log.Fields{
			"package":   "apiserver",
			"function":  "PostTrackHasHandler",
			"principal": principal.(*config.ApiUser).Name,
			"yid":       *params.Body.Yid,
			"submitter": *params.Body.Submitter,
		}, "track not found in database")
		return operations.NewPostTrackHasNotFound()
	})

	api.PostTrackLengthHandler = operations.PostTrackLengthHandlerFunc(func(params operations.PostTrackLengthParams, principal interface{}) middleware.Responder {
		if mgr.IsAllowedLength(*params.Body.Yid) {
			log.Debugf(log.Fields{
				"package":   "apiserver",
				"function":  "PostTrackHasHandler",
				"call":      "mgr.IsAllowedLength",
				"principal": principal.(*config.ApiUser).Name,
				"yid":       *params.Body.Yid,
				"submitter": *params.Body.Submitter,
			}, "song is short enough for stream")
			return operations.NewPostTrackLengthNoContent()
		}

		log.Debugf(log.Fields{
			"package":   "apiserver",
			"function":  "PostTrackLengthHandler",
			"principal": principal.(*config.ApiUser).Name,
			"yid":       *params.Body.Yid,
			"submitter": *params.Body.Submitter,
		}, "song is too long for stream")
		return operations.NewPostTrackLengthNotFound()
	})

	api.PostTrackTitleHandler = operations.PostTrackTitleHandlerFunc(func(params operations.PostTrackTitleParams, principal interface{}) middleware.Responder {
		title, err := mgr.GetTitle(*params.Body.Yid)
		if err != nil {
			return operations.NewPostTrackTitleNotFound()
		}

		log.Debugf(log.Fields{
			"package":   "apiserver",
			"function":  "PostTrackTitleHandler",
			"principal": principal.(*config.ApiUser).Name,
			"yid":       *params.Body.Yid,
			"submitter": *params.Body.Submitter,
			"title":     title,
		}, "fetched title for track")

		return operations.NewPostTrackTitleOK().WithPayload(title)
	})

	api.PostTrackDownloadHandler = operations.PostTrackDownloadHandlerFunc(func(params operations.PostTrackDownloadParams, principal interface{}) middleware.Responder {
		err := isAuthorized(principal, "allowTrackDownload")
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "apiserver",
				"function": "PostTrackDownloadHandler",
				"call":     "isAuthorized",
			}, err.Error())
			return operations.NewPostTrackDownloadForbidden()
		}

		if mgr.HasYid(*params.Body.Yid) {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "PostTrackDownloadHandler",
				"call":      "mgr.HasYid",
				"principal": principal.(*config.ApiUser).Name,
				"yid":       *params.Body.Yid,
				"submitter": *params.Body.Submitter,
			}, "yid already downloaded")
			return operations.NewPostTrackDownloadConflict()
		}

		if !mgr.IsAllowedLength(*params.Body.Yid) {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "PostTrackDownloadHandler",
				"call":      "mgr.IsAllowedLength",
				"principal": principal.(*config.ApiUser).Name,
				"yid":       *params.Body.Yid,
				"submitter": *params.Body.Submitter,
			}, "song is too long for stream")
			return operations.NewPostTrackDownloadRequestEntityTooLarge()
		}

		track, err := mgr.AddTrack(*params.Body.Yid, *params.Body.Submitter)
		if err != nil {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "PostTrackDownloadHandler",
				"call":      "mgr.AddTrack",
				"principal": principal.(*config.ApiUser).Name,
				"yid":       *params.Body.Yid,
				"submitter": *params.Body.Submitter,
			}, err.Error())
			return operations.NewPostTrackDownloadBadRequest()
		}

		addedOnTs := track.AddedOn.String()
		duration := int64(track.Duration)

		response := operations.PostTrackDownloadOKBody{
			Yid:       &track.Yid,
			Filename:  &track.Filename,
			Addedon:   &addedOnTs,
			Duration:  &duration,
			Rating:    &track.Rating,
			Submitter: &track.Submitter,
		}

		log.Debugf(log.Fields{
			"package":   "apiserver",
			"function":  "PostTrackDownloadHandler",
			"principal": principal.(*config.ApiUser).Name,
			"yid":       *params.Body.Yid,
			"submitter": *params.Body.Submitter,
			"filename":  track.Filename,
		}, "track added succesfully")

		return operations.NewPostTrackDownloadOK().WithPayload(&response)
	})

	api.PostTrackRequestHandler = operations.PostTrackRequestHandlerFunc(func(params operations.PostTrackRequestParams, principal interface{}) middleware.Responder {
		err := isAuthorized(principal, "allowTrackRequest")
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "apiserver",
				"function": "PostTrackRequestHandler",
				"call":     "isAuthorized",
			}, err.Error())
			return operations.NewPostTrackRequestForbidden()
		}

		track, err := mgr.Request(*params.Request.Query, *params.Request.Submitter)
		if err != nil {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "PostTrackRequestHandler",
				"call":      "mgr.Request",
				"principal": principal.(*config.ApiUser).Name,
				"query":     *params.Request.Query,
				"submitter": *params.Request.Submitter,
			}, err.Error())
			return operations.NewPostTrackRequestNotFound()
		}

		if track.Filename == "" {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "PostTrackRequestHandler",
				"call":      "mgr.Request",
				"principal": principal.(*config.ApiUser).Name,
				"query":     *params.Request.Query,
				"submitter": *params.Request.Submitter,
			}, "no results found")
			return operations.NewPostTrackRequestNotFound()
		}

		addedOnTs := track.AddedOn.String()
		duration := int64(track.Duration)

		response := operations.PostTrackRequestOKBody{
			Submitter: params.Request.Submitter,
			Track: &operations.PostTrackRequestOKBodyTrack{
				Yid:       &track.Yid,
				Filename:  &track.Filename,
				Addedon:   &addedOnTs,
				Duration:  &duration,
				Rating:    &track.Rating,
				Submitter: &track.Submitter,
			},
		}

		log.Debugf(log.Fields{
			"package":   "apiserver",
			"function":  "PostTrackRequestHandler",
			"call":      "mgr.Request",
			"principal": principal.(*config.ApiUser).Name,
			"query":     *params.Request.Query,
			"submitter": *params.Request.Submitter,
			"filename":  *response.Track.Filename,
		}, "track queued")

		return operations.NewPostTrackRequestOK().WithPayload(&response)
	})

	api.PostTrackSearchHandler = operations.PostTrackSearchHandlerFunc(func(params operations.PostTrackSearchParams, principal interface{}) middleware.Responder {
		err := isAuthorized(principal, "allowTrackSearch")
		if err != nil {
			log.Warningf(log.Fields{
				"package":  "apiserver",
				"function": "PostTrackSearchHandler",
				"call":     "isAuthorized",
			}, err.Error())
			return operations.NewPostTrackSearchForbidden()
		}

		tracks, err := mgr.Search(*params.Request.Query, *params.Request.Submitter)
		if err != nil {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "PostTrackSearchHandler",
				"call":      "mgr.Search",
				"principal": principal.(*config.ApiUser).Name,
				"query":     *params.Request.Query,
				"submitter": *params.Request.Submitter,
			}, err.Error())
			return operations.NewPostTrackSearchNotFound()
		}

		if len(tracks) == 0 {
			log.Warningf(log.Fields{
				"package":   "apiserver",
				"function":  "PostTrackSearchHandler",
				"call":      "mgr.Search",
				"principal": principal.(*config.ApiUser).Name,
				"query":     *params.Request.Query,
				"submitter": *params.Request.Submitter,
			}, "no results found")
			return operations.NewPostTrackSearchNotFound()
		}

		foundTracks := make([]*operations.PostTrackSearchOKBodyItems0, 0)

		for _, track := range tracks {
			yid := track.Yid
			filename := track.Filename
			addedOnTs := track.AddedOn.String()
			duration := int64(track.Duration)
			rating := track.Rating
			submitter := track.Submitter

			responseTrack := &operations.PostTrackSearchOKBodyItems0{
				Yid:       &yid,
				Filename:  &filename,
				Addedon:   &addedOnTs,
				Duration:  &duration,
				Rating:    &rating,
				Submitter: &submitter,
			}

			foundTracks = append(foundTracks, responseTrack)
		}

		log.Debugf(log.Fields{
			"package":   "apiserver",
			"function":  "PostTrackSearchHandler",
			"call":      "mgr.Search",
			"principal": principal.(*config.ApiUser).Name,
			"query":     *params.Request.Query,
			"submitter": *params.Request.Submitter,
			"num_found": len(foundTracks),
		}, "returning results")

		return operations.NewPostTrackSearchOK().WithPayload(foundTracks)
	})

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
