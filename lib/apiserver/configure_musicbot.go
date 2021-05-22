// This file is safe to edit. Once it exists it will not be overwritten

package apiserver

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/r3boot/test/lib/manager"
	"github.com/sirupsen/logrus"

	"github.com/r3boot/test/lib/config"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/r3boot/test/lib/apiserver/operations"
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

	return nil, fmt.Errorf("Authentication failed")
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

	if Config == nil {
		log.Printf("Config is nil\n")
	}

	// Setup Manager
	mgr, err := manager.NewManager(Config)
	if err != nil {
		log.Printf("NewManager: %v", err)
		return nil
	}

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()
	api.UrlformConsumer = runtime.DiscardConsumer

	api.JSONProducer = runtime.JSONProducer()

	// Applies when the "X-Access-Token" header is set
	api.AccessSecurityAuth = func(token string) (interface{}, error) {
		log := logrus.WithFields(logrus.Fields{
			"module":   "apiserver",
			"function": "AccessSecurityAuth",
		})
		principal, err := validateToken(token)
		if err != nil {
			return nil, fmt.Errorf("validateToken: %v", err)
		}
		if principal == nil {
			return nil, fmt.Errorf("validateToken: principal is nil")
		}
		log.Infof("authenticated as %s", principal.Name)
		return principal, nil
	}

	// Set your custom authorizer if needed. Default one is security.Authorized()
	// Expected interface runtime.Authorizer
	//
	// Example:
	// api.APIAuthorizer = security.Authorized()

	api.GetPlayerNextHandler = operations.GetPlayerNextHandlerFunc(func(params operations.GetPlayerNextParams, principal interface{}) middleware.Responder {
		log := logrus.WithFields(logrus.Fields{
			"module":   "apiserver",
			"function": "GetPlayerNextHandler",
		})

		err := isAuthorized(principal, "allowPlayerNext")
		if err != nil {
			log.Warnf("isAuthorized: %v", err)
			return operations.NewGetPlayerNextForbidden()
		}

		err = mgr.Next()
		if err != nil {
			log.Warnf("mgr.Next: %v", err)
			return operations.NewGetPlayerNextBadRequest()
		}

		track, err := mgr.NowPlaying()
		if err != nil {
			log.Warnf("mgr.NowPlaying: %v", err)
			return operations.NewGetPlayerNowplayingBadRequest()
		}

		addedOnTs := track.AddedOn.String()
		duration := int64(track.Duration)

		response := operations.GetPlayerNextOKBody{
			Filename:  &track.Filename,
			Addedon:   &addedOnTs,
			Duration:  &duration,
			Rating:    &track.Rating,
			Submitter: &track.Submitter,
		}

		log.Infof("Skipped to %v", track.Filename)

		return operations.NewGetPlayerNextOK().WithPayload(&response)
	})

	api.GetPlayerNowplayingHandler = operations.GetPlayerNowplayingHandlerFunc(func(params operations.GetPlayerNowplayingParams, principal interface{}) middleware.Responder {
		log := logrus.WithFields(logrus.Fields{
			"module":   "apiserver",
			"function": "GetPlayerNowplayingHandler",
		})

		err := isAuthorized(principal, "allowPlayerNowPlaying")
		if err != nil {
			log.Warnf("isAuthorized: %v", err)
			return operations.NewGetPlayerNowplayingForbidden()
		}

		track, err := mgr.NowPlaying()
		if err != nil {
			log.Warnf("mgr.NowPlaying: %v", err)
			return operations.NewGetPlayerNowplayingBadRequest()
		}

		addedOnTs := track.AddedOn.String()
		duration := int64(track.Duration)

		response := operations.GetPlayerNowplayingOKBody{
			Filename:  &track.Filename,
			Addedon:   &addedOnTs,
			Duration:  &duration,
			Rating:    &track.Rating,
			Submitter: &track.Submitter,
		}

		log.Infof("Now playing %v", track.Filename)

		return operations.NewGetPlayerNowplayingOK().WithPayload(&response)
	})

	api.GetPlayerQueueHandler = operations.GetPlayerQueueHandlerFunc(func(params operations.GetPlayerQueueParams, principal interface{}) middleware.Responder {
		log := logrus.WithFields(logrus.Fields{
			"module":   "apiserver",
			"function": "GetPlayerQueueHandler",
		})

		err := isAuthorized(principal, "allowPlayerQueue")
		if err != nil {
			log.Warnf("isAuthorized: %v", err)
			return operations.NewGetPlayerQueueForbidden()
		}

		entries, err := mgr.GetQueue()
		if err != nil {
			log.Printf("mgr.GetQueue: %v\n", err)
			return operations.NewGetPlayerNextBadRequest()
		}

		foundTracks := make([]*operations.GetPlayerQueueOKBodyItems0, 0)

		for _, track := range entries {
			filename := track.Filename
			addedOnTs := track.AddedOn.String()
			duration := int64(track.Duration)
			rating := track.Rating
			submitter := track.Submitter

			responseTrack := operations.GetPlayerQueueOKBodyItems0{
				Filename:  &filename,
				Addedon:   &addedOnTs,
				Duration:  &duration,
				Rating:    &rating,
				Submitter: &submitter,
			}

			foundTracks = append(foundTracks, &responseTrack)
		}

		log.Infof("Sending queue request")

		return operations.NewGetPlayerQueueOK().WithPayload(foundTracks)
	})

	api.GetRatingDecreaseHandler = operations.GetRatingDecreaseHandlerFunc(func(params operations.GetRatingDecreaseParams, principal interface{}) middleware.Responder {
		log := logrus.WithFields(logrus.Fields{
			"module":   "apiserver",
			"function": "GetRatingDecreaseHandler",
		})

		err := isAuthorized(principal, "allowRatingDecrease")
		if err != nil {
			log.Warnf("isAuthorized: %v", err)
			return operations.NewGetRatingDecreaseForbidden()
		}

		err = mgr.DecreaseRating()
		if err != nil {
			log.Warnf("mgr.DecreaseRating: %v\n", err)
			return operations.NewGetRatingDecreaseBadRequest()
		}

		track, err := mgr.NowPlaying()
		if err != nil {
			log.Warnf("mgr.NowPlaying: %v", err)
			return operations.NewGetRatingDecreaseBadRequest()
		}

		addedOnTs := track.AddedOn.String()
		duration := int64(track.Duration)

		response := operations.GetRatingDecreaseOKBody{
			Filename:  &track.Filename,
			Addedon:   &addedOnTs,
			Duration:  &duration,
			Rating:    &track.Rating,
			Submitter: &track.Submitter,
		}

		log.Infof("Decreased rating for %s to %d", track.Filename, track.Rating)

		return operations.NewGetRatingDecreaseOK().WithPayload(&response)
	})

	api.GetRatingIncreaseHandler = operations.GetRatingIncreaseHandlerFunc(func(params operations.GetRatingIncreaseParams, principal interface{}) middleware.Responder {
		log := logrus.WithFields(logrus.Fields{
			"module":   "apiserver",
			"function": "GetRatingIncreaseHandler",
		})

		err := isAuthorized(principal, "allowRatingIncrease")
		if err != nil {
			log.Warnf("isAuthorized: %v", err)
			return operations.NewGetRatingIncreaseForbidden()
		}

		err = mgr.IncreaseRating()
		if err != nil {
			log.Warnf("mgr.IncreaseRating: %v", err)
			return operations.NewGetRatingIncreaseBadRequest()
		}

		track, err := mgr.NowPlaying()
		if err != nil {
			log.Warnf("mgr.NowPlaying: %v", err)
			return operations.NewGetRatingIncreaseBadRequest()
		}

		addedOnTs := track.AddedOn.String()
		duration := int64(track.Duration)

		response := operations.GetRatingIncreaseOKBody{
			Filename:  &track.Filename,
			Addedon:   &addedOnTs,
			Duration:  &duration,
			Rating:    &track.Rating,
			Submitter: &track.Submitter,
		}

		log.Infof("Decreased rating for %s to %d", track.Filename, track.Rating)

		return operations.NewGetRatingIncreaseOK().WithPayload(&response)
	})

	api.PostTrackDownloadHandler = operations.PostTrackDownloadHandlerFunc(func(params operations.PostTrackDownloadParams, principal interface{}) middleware.Responder {
		log := logrus.WithFields(logrus.Fields{
			"module":   "apiserver",
			"function": "PostTrackDownloadHandler",
		})

		err := isAuthorized(principal, "allowTrackDownload")
		if err != nil {
			log.Warnf("isAuthorized: %v", err)
			return operations.NewPostTrackDownloadForbidden()
		}

		track, err := mgr.AddTrack(*params.Body.Yid, *params.Body.Submitter)
		if err != nil {
			log.Warnf("AddTrack: %v\n", err)
			return operations.NewPostTrackDownloadBadRequest()
		}

		addedOnTs := track.AddedOn.String()
		duration := int64(track.Duration)

		response := operations.PostTrackDownloadOKBody{
			Filename:  &track.Filename,
			Addedon:   &addedOnTs,
			Duration:  &duration,
			Rating:    &track.Rating,
			Submitter: &track.Submitter,
		}

		log.Infof("Downloaded %s", track.Filename)

		return operations.NewPostTrackDownloadOK().WithPayload(&response)
	})

	api.PostTrackRequestHandler = operations.PostTrackRequestHandlerFunc(func(params operations.PostTrackRequestParams, principal interface{}) middleware.Responder {
		log := logrus.WithFields(logrus.Fields{
			"module":    "apiserver",
			"function":  "PostTrackRequestHandler",
			"query":     *params.Request.Query,
			"submitter": *params.Request.Submitter,
		})

		err := isAuthorized(principal, "allowTrackRequest")
		if err != nil {
			log.Warnf("isAuthorized: %v", err)
			return operations.NewPostTrackRequestForbidden()
		}

		track, err := mgr.Request(*params.Request.Query, *params.Request.Submitter)
		if err != nil {
			log.Warnf("mgr.Search: %v", err)
			return operations.NewPostTrackRequestNotFound()
		}

		if track.Filename == "" {
			log.Warnf("mgr.Search: no results", err)
			return operations.NewPostTrackRequestNotFound()
		}

		addedOnTs := track.AddedOn.String()
		duration := int64(track.Duration)

		response := operations.PostTrackRequestOKBody{
			Submitter: params.Request.Submitter,
			Track: &operations.PostTrackRequestOKBodyTrack{
				Filename:  &track.Filename,
				Addedon:   &addedOnTs,
				Duration:  &duration,
				Rating:    &track.Rating,
				Submitter: &track.Submitter,
			},
		}

		log.Infof("Handled request")

		return operations.NewPostTrackRequestOK().WithPayload(&response)
	})

	api.PostTrackSearchHandler = operations.PostTrackSearchHandlerFunc(func(params operations.PostTrackSearchParams, principal interface{}) middleware.Responder {
		log := logrus.WithFields(logrus.Fields{
			"module":    "apiserver",
			"function":  "PostTrackSearchHandler",
			"query":     *params.Request.Query,
			"submitter": *params.Request.Submitter,
		})

		err := isAuthorized(principal, "allowTrackSearch")
		if err != nil {
			log.Warnf("isAuthorized: %v", err)
			return operations.NewPostTrackSearchForbidden()
		}

		tracks, err := mgr.Search(*params.Request.Query, *params.Request.Submitter)
		if err != nil {
			log.Warnf("mgr.Search: %v", err)
			return operations.NewPostTrackSearchNotFound()
		}

		if len(tracks) == 0 {
			log.Warnf("mgr.Search: no results", err)
			return operations.NewPostTrackSearchNotFound()
		}

		foundTracks := make([]*operations.PostTrackSearchOKBodyItems0, 0)

		for _, track := range tracks {
			filename := track.Filename
			addedOnTs := track.AddedOn.String()
			duration := int64(track.Duration)
			rating := track.Rating
			submitter := track.Submitter

			responseTrack := &operations.PostTrackSearchOKBodyItems0{
				Filename:  &filename,
				Addedon:   &addedOnTs,
				Duration:  &duration,
				Rating:    &rating,
				Submitter: &submitter,
			}

			foundTracks = append(foundTracks, responseTrack)
		}

		log.Infof("Handled search")

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
