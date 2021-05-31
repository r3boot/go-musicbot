package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/r3boot/go-musicbot/lib/apiclient"
	"github.com/r3boot/go-musicbot/lib/apiclient/operations"
	"github.com/sirupsen/logrus"
)

const (
	defHostValue = "localhost"
	defPortValue = 8080

	helpHost = "Host to connect to"
	helpPort = "Port to connect to"

	musicbotSubmitter = "musicbot"

	envVarPrefix = "MUSICBOTCTL_"
	envVarHost   = envVarPrefix + "HOST"
	envVarPort   = envVarPrefix + "PORT"
	envVarToken  = envVarPrefix + "TOKEN"
)

func argOrEnvVar(argValue interface{}, envVarName string) (interface{}, error) {
	result := argValue
	envValue, ok := os.LookupEnv(envVarName)
	if ok {
		result = envValue
	}

	if result == "" {
		return "", fmt.Errorf("No value found")
	}

	return result, nil
}

func HandleNext(client *apiclient.Musicbot, token runtime.ClientAuthInfoWriter) error {
	resp, err := client.Operations.GetPlayerNext(operations.NewGetPlayerNextParams(), token)
	if err != nil {
		return fmt.Errorf("GetPlayerNext: %v", err)
	}

	track := resp.GetPayload()

	fmt.Printf("Skipped to %s\n", *track.Filename)
	return nil
}

func HandleIncreaseRating(client *apiclient.Musicbot, token runtime.ClientAuthInfoWriter) error {
	resp, err := client.Operations.GetRatingIncrease(operations.NewGetRatingIncreaseParams(), token)
	if err != nil {
		return fmt.Errorf("GetRatingIncrease: %v", err)
	}

	track := resp.GetPayload()

	fmt.Printf("Rating for %s is %d/10\n", *track.Filename, *track.Rating)
	return nil
}

func HandleDecreaseRating(client *apiclient.Musicbot, token runtime.ClientAuthInfoWriter) error {
	resp, err := client.Operations.GetRatingDecrease(operations.NewGetRatingDecreaseParams(), token)
	if err != nil {
		return fmt.Errorf("GetRatingDecrease: %v", err)
	}

	track := resp.GetPayload()

	fmt.Printf("Rating for %s is %d/10\n", *track.Filename, *track.Rating)
	return nil
}

func HandleAddYid(client *apiclient.Musicbot, token runtime.ClientAuthInfoWriter, yid, submitter string) error {
	params := operations.NewPostTrackDownloadParams()
	params.Body = operations.PostTrackDownloadBody{
		Yid:       &yid,
		Submitter: &submitter,
	}

	_, err := client.Operations.PostTrackDownload(params, token)
	if err != nil {
		errmsg := err.Error()
		if strings.Contains(errmsg, "postTrackDownloadRequestEntityTooLarge") {
			return fmt.Errorf("Track is too long for stream")
		} else if strings.Contains(errmsg, "postTrackDownloadConflict") {
			return fmt.Errorf("Track already downloaded")
		}
		return fmt.Errorf("PostTrackDownload: %v", err)
	}

	fmt.Printf("Track added succesfully\n")

	return nil
}

func HandleNowPlaying(client *apiclient.Musicbot, token runtime.ClientAuthInfoWriter) error {
	resp, err := client.Operations.GetPlayerNowplaying(operations.NewGetPlayerNowplayingParams(), token)
	if err != nil {
		return fmt.Errorf("GetPlayerNowPlaying: %v", err)
	}

	track := resp.GetPayload()

	fmt.Printf("> Currently playing:\n")
	fmt.Printf("File:      %s\n", *track.Filename)
	fmt.Printf("Elapsed:   %ds\n", *track.Elapsed)
	fmt.Printf("Duration:  %ds\n", *track.Duration)
	fmt.Printf("Rating     %d/10\n", *track.Rating)
	fmt.Printf("Submitter: %s\n", *track.Submitter)
	fmt.Printf("AddedOn:   %s\n", *track.Addedon)

	return nil
}

func HandleSearch(client *apiclient.Musicbot, token runtime.ClientAuthInfoWriter, query, submitter string) error {
	log := logrus.WithFields(logrus.Fields{
		"module":    "main",
		"query":     query,
		"submitter": submitter,
	})
	log.Printf("Sending search request")

	params := operations.NewPostTrackSearchParams()
	params.Request = operations.PostTrackSearchBody{
		Query:     &query,
		Submitter: &submitter,
	}

	response, err := client.Operations.PostTrackSearch(params, token)
	if err != nil {
		return fmt.Errorf("PostTrackSearch: %v", err)
	}

	for _, track := range response.Payload {
		fmt.Printf("%s\n", *track.Filename)
	}

	return nil
}

func HandleRequest(client *apiclient.Musicbot, token runtime.ClientAuthInfoWriter, query, submitter string) error {
	log := logrus.WithFields(logrus.Fields{
		"module":    "main",
		"query":     query,
		"submitter": submitter,
	})
	log.Printf("Sending request")

	params := operations.NewPostTrackRequestParams()
	params.Request = operations.PostTrackRequestBody{
		Query:     &query,
		Submitter: &submitter,
	}

	response, err := client.Operations.PostTrackRequest(params, token)
	if err != nil {
		return fmt.Errorf("PostTrackSearch: %v", err)
	}

	fmt.Printf("Your request %s is on position %d in the queue\n", *response.Payload.Track.Filename, response.Payload.Track.Priority)

	return nil
}

func HandleGetQueue(client *apiclient.Musicbot, token runtime.ClientAuthInfoWriter) error {
	log := logrus.WithFields(logrus.Fields{
		"module": "HandleGetQueue",
	})

	log.Printf("Fetching current request queue")

	response, err := client.Operations.GetPlayerQueue(operations.NewGetPlayerQueueParams(), token)
	if err != nil {
		return fmt.Errorf("GetPlayerQueue: %v", err)
	}

	queueEntries := response.Payload

	for i := 0; i < len(queueEntries); i++ {
		fmt.Printf("%d) %s\n", i, *queueEntries[i].Filename)
	}

	if len(queueEntries) == 0 {
		log.Printf("Queue is empty")
	}

	return nil
}

func main() {
	var (
		Host     = flag.String("host", defHostValue, helpHost)
		Port     = flag.Int("port", defPortValue, helpPort)
		LogLevel = flag.String("loglevel", "INFO", "Log level to use (INFO, DEBUG)")
		LogJson  = flag.Bool("json", false, "Output logging in JSON format")

		Token = flag.String("token", "", "Authentication token to use")

		GetNowPlaying  = flag.Bool("nowplaying", false, "Get currently playing track")
		Next           = flag.Bool("next", false, "Skip to the next track")
		IncreaseRating = flag.Bool("increaserating", false, "Increase the rating for the current track")
		DecreaseRating = flag.Bool("decreaserating", false, "Decrease the rating for the current track")
		AddYid         = flag.String("addyid", "", "Add a new track by Youtube ID")
		Search         = flag.String("search", "", "Search for something")
		Request        = flag.String("request", "", "Request something")
		Queue          = flag.Bool("queue", false, "Returns the current request queue")

		log *logrus.Entry
	)
	flag.Parse()

	// Configure logging
	if *LogJson {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	switch strings.ToUpper(*LogLevel) {
	case "INFO":
		{
			logrus.SetLevel(logrus.InfoLevel)
		}
	case "DEBUG":
		{
			logrus.SetLevel(logrus.DebugLevel)
		}
	}

	// Initialize logging
	log = logrus.WithFields(logrus.Fields{
		"module": "main",
	})

	host, err := argOrEnvVar(*Host, envVarHost)
	if err != nil {
		log := logrus.WithFields(logrus.Fields{
			"module": "main",
			"key":    "host",
		})
		log.Fatalf("No value set, please pass -host or set %s", envVarHost)
	}

	port, err := argOrEnvVar(*Port, envVarPort)
	if err != nil {
		log := logrus.WithFields(logrus.Fields{
			"module": "main",
			"key":    "port",
		})
		log.Fatalf("No value set, please pass -port or set %s", envVarPort)
	}

	token, err := argOrEnvVar(*Token, envVarToken)
	if err != nil {
		log := logrus.WithFields(logrus.Fields{
			"module": "main",
			"key":    "token",
		})
		log.Fatalf("No value set, please pass -token or set %s", envVarToken)
	}

	uri := fmt.Sprintf("%s:%d", host, port)

	httptransport.DefaultTimeout = 300 * time.Second
	transport := httptransport.New(uri, "/v1/", nil)

	apiToken := httptransport.APIKeyAuth("X-Access-Token", "header", token.(string))

	transport.Consumers["application/json"] = runtime.JSONConsumer()
	transport.Consumers["application/vnd.api+json"] = runtime.JSONConsumer()
	client := apiclient.New(transport, strfmt.Default)

	if *GetNowPlaying {
		err := HandleNowPlaying(client, apiToken)
		if err != nil {
			log.Fatalf("%v", err)
		}
	} else if *Next {
		err := HandleNext(client, apiToken)
		if err != nil {
			log.Fatalf("%v", err)
		}
	} else if *IncreaseRating {
		err := HandleIncreaseRating(client, apiToken)
		if err != nil {
			log.Fatalf("%v", err)
		}
	} else if *DecreaseRating {
		err := HandleDecreaseRating(client, apiToken)
		if err != nil {
			log.Fatalf("%v", err)
		}
	} else if *AddYid != "" {
		err := HandleAddYid(client, apiToken, *AddYid, musicbotSubmitter)
		if err != nil {
			log.Fatalf("%v", err)
		}
	} else if *Search != "" {
		err := HandleSearch(client, apiToken, *Search, musicbotSubmitter)
		if err != nil {
			log.Fatalf("%v", err)
		}
	} else if *Request != "" {
		err := HandleRequest(client, apiToken, *Request, musicbotSubmitter)
		if err != nil {
			log.Fatalf("%v", err)
		}
	} else if *Queue {
		err := HandleGetQueue(client, apiToken)
		if err != nil {
			log.Fatalf("%v", err)
		}
	} else {
		err := HandleNowPlaying(client, apiToken)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
}
