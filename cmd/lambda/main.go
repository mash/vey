package main

import (
	"bytes"
	"context"
	"errors"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
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
	adapter *httpadapter.HandlerAdapter
	// injected via go build -ldflags
	Version   string
	BuildDate string
)

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
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
		log.Debug().Msg("Debug logging enabled.")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	vhttp.Log = NewLogger()

	k := vey.NewVey(vey.NewDigester(cfg.Salt), vey.NewMemCache(), vey.NewMemStore())

	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create aws session")
	}

	emailConfig, err := loadEmailConfig("email.yml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load email.yml")
	}
	log.Debug().Msgf("email config: %+v", emailConfig)

	svc := ses.New(sess)
	s := email.NewLogSender(email.NewSESSender(emailConfig, svc))
	h := vhttp.NewHandler(k, s)
	log.Info().Msg("starting")
	adapter = httpadapter.New(h)
}

type Config struct {
	Salt  []byte `yaml:"salt"`
	Debug bool   `yaml:"debug"`
}

// loadConfig loads config from file encrypted with sops.
func loadConfig(file string) (Config, error) {
	b, err := decrypt.File(file, "yaml")
	if err != nil {
		return Config{}, err
	}

	buf := bytes.NewBuffer(b)
	dec := yaml.NewDecoder(buf)
	var c Config
	if err := dec.Decode(&c); err != nil {
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
		log.Error().Err(err).Str("code", aerr.Code()).Msg(aerr.Error())
		return
	}

	log.Error().Err(err).Msg("error")
}
