package main

import (
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/mash/vey"
	"github.com/mash/vey/email"
	vhttp "github.com/mash/vey/http"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/alecthomas/kingpin.v2"
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

	serve     = app.Command("serve", "Start server")
	servePort = serve.Arg("port", "Server listens on this port").Default("8000").Envar("VEY_PORT").String()
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Info().Str("version", Version).Str("buildDate", BuildDate)
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case version.FullCommand():
		break
	case serve.FullCommand():
		salt := []byte("salt")
		k := vey.NewVey(vey.NewDigester(salt), vey.NewMemCache(), vey.NewMemStore())
		s := email.NewLogSender()
		h := vhttp.NewHandler(k, s)
		log.Info().Msg("listening on port " + *servePort)
		http.ListenAndServe(":"+*servePort, h)
	}
}
