package main

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/mash/vey"
	"github.com/mash/vey/email"
	vhttp "github.com/mash/vey/http"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

// injected via go build -ldflags
var (
	Version   string
	BuildDate string
)

var (
	app     = kingpin.New("vey", "Vey - Email Verifying Keyserver")
	debug   = app.Flag("debug", "Debug level logging turns on.").Bool()
	version = app.Command("version", "Show version")

	serve               = app.Command("serve", "Start server")
	servePort           = serve.Flag("port", "Server listens on this port").Default("8000").Envar("VEY_PORT").String()
	serveEmailConfig    = serve.Flag("emailConfig", "Email configuration file").Default("email.yml").Envar("VEY_EMAIL_CONFIG").String()
	serveStore          = serve.Flag("store", "Store implementation. Can be \"dynamodb\" or \"memory\".").Default("memory").String()
	serveStoreDynDBName = serve.Flag("store-dyndb-name", "DynamoDB table name used to implement Store interface").Default("veystore").String()
	serveCache          = serve.Flag("cache", "Cache implementation").Default("memory").String()
	serveCacheDynDBName = serve.Flag("cache-dyndb-name", "DynamoDB table name used to implement Cache interface").Default("veycache").String()
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	if *debug {
		log.Debug().Msg("Debug logging enabled.")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	vhttp.Log = NewLogger()

	switch cmd {
	case version.FullCommand():
		log.Info().Str("buildDate", BuildDate).Str("version", Version).Msg("")

	case serve.FullCommand():
		salt := []byte("salt")

		sess, err := session.NewSession(&aws.Config{})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to create aws session")
		}

		var store vey.Store
		if *serveStore == "dynamodb" {
			log.Debug().Msgf("using dynamodb store: %s", *serveStoreDynDBName)
			svc := dynamodb.New(sess)
			store = vey.NewDynamoDbStore(*serveStoreDynDBName, svc)
		} else {
			log.Debug().Msg("using memory store")
			store = vey.NewMemStore()
		}

		var cache vey.Cache
		if *serveCache == "dynamodb" {
			log.Debug().Msgf("using dynamodb cache: %s", *serveCacheDynDBName)
			svc := dynamodb.New(sess)
			cache = vey.NewDynamoDbCache(*serveCacheDynDBName, svc, 15*time.Minute)
		} else {
			log.Debug().Msg("using memory cache")
			cache = vey.NewMemCache(15 * time.Minute)
		}

		k := vey.NewVey(vey.NewDigester(salt), cache, store)

		f, err := os.Open(*serveEmailConfig)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to open email config file: " + *serveEmailConfig)
		}
		dec := yaml.NewDecoder(f)
		var emailConfig email.SESConfig
		if err := dec.Decode(&emailConfig); err != nil {
			log.Fatal().Err(err).Msg("failed to decode email config file: " + *serveEmailConfig)
		}
		log.Debug().Str("email config file", *serveEmailConfig).Msgf("config: %+v", emailConfig)

		svc := ses.New(sess)
		s := email.NewLogSender(email.NewSESSender(emailConfig, svc))
		h := vhttp.NewHandler(k, s, nil)
		log.Info().Msg("listening on port " + *servePort)
		http.ListenAndServe(":"+*servePort, h)
	}
}

type logger struct{}

// NewLogger returns a new default Logger that logs to stderr.
func NewLogger() vhttp.Logger {
	return logger{}
}

func (l logger) Error(err error) {
	var er vhttp.Error
	if errors.As(err, &er) {
		if e := er.Unwrap(); e != nil {
			log.Error().Err(e).Int("code", er.Code).Msg(er.Msg)
			return
		}
	}

	if aerr, ok := err.(awserr.Error); ok {
		log.Error().Err(err).Str("code", aerr.Code()).Msg(aerr.Error())
		return
	}

	log.Error().Err(err).Msg("error")
}
