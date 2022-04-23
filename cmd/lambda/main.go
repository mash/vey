package main

import (
	"context"
	"encoding/base64"
	"errors"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/mash/vey"
	"github.com/mash/vey/email"
	vhttp "github.com/mash/vey/http"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mozilla.org/sops/v3/decrypt"
	"gopkg.in/yaml.v2"
)

var (
	adapter *httpadapter.HandlerAdapterV2
	// injected via go build -ldflags
	Version   string
	BuildDate string
)

// API Gateway uses Payload format version v2.0.
func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return adapter.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
}

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	setup()
}

func setup() {
	cfg, err := loadConfig("vey.yml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load vey.yml")
	}

	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("debug logging enabled")
	}

	var open *url.URL
	if cfg.OpenURL != "" {
		u, err := url.Parse(cfg.OpenURL)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to parse OpenURL")
		}
		open = u
	}

	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create aws session")
	}

	svc := dynamodb.New(sess)
	cache := vey.NewDynamoDbCache(cfg.CacheTableName, svc, cfg.CacheExpiry)
	store := vey.NewDynamoDbStore(cfg.StoreTableName, svc)

	salt, err := base64.StdEncoding.DecodeString(cfg.Salt)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to decode salt")
	}
	k := vey.NewVey(vey.NewDigester(salt), cache, store)

	emailConfig, err := loadEmailConfig("email.yml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load email.yml")
	}

	sender := email.NewSESSender(emailConfig, ses.New(sess))
	if cfg.Debug {
		sender = email.NewLogSender(sender)
	}
	h := vhttp.NewHandler(k, sender, open)

	vhttp.Log = NewLogger()

	log.Info().Msg("starting")
	adapter = httpadapter.NewV2(h)
}

type Config struct {
	// base64 encoded
	Salt           string        `yaml:"salt"`
	Debug          bool          `yaml:"debug"`
	StoreTableName string        `yaml:"store_table_name"`
	CacheTableName string        `yaml:"cache_table_name"`
	CacheExpiry    time.Duration `yaml:"cache_expiry"`
	OpenURL        string        `yaml:"open_url"`
}

// loadConfig loads config from file encrypted with sops.
func loadConfig(file string) (Config, error) {
	b, err := decrypt.File(file, "yaml")
	if err != nil {
		return Config{}, err
	}

	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return c, err
	}
	return c, nil
}

func loadEmailConfig(file string) (email.SESConfig, error) {
	f, err := os.Open(file)
	if err != nil {
		return email.SESConfig{}, err
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	var emailConfig email.SESConfig
	if err := dec.Decode(&emailConfig); err != nil {
		return email.SESConfig{}, err
	}
	return emailConfig, nil
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
		log.Error().Err(err).Str("awscode", aerr.Code()).Msg(aerr.Error())
		return
	}

	log.Error().Err(err).Msg("error")
}
