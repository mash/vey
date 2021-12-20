package main

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

// injected via go build -ldflags
var (
	Version   string
	BuildDate string
)

var (
	app      = kingpin.New("vey", "Vey - Email Verifying Keyserver")
	logLevel = app.Flag("loglevel", "Log level (debug, info, error)").Default("info").Envar("VEY_LOGLEVEL").Enum("debug", "info", "error")
	version  = app.Command("version", "Show version")

	serve     = app.Command("serve", "Start server")
	servePort = serve.Arg("port", "Server listens on this port").Default("8000").Envar("VEY_PORT").String()
)

func main() {
	log.Info().Str("version", Version).Str("buildDate", BuildDate)
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case version.FullCommand():
		break
	case serve.FullCommand():
		break
	}
}
