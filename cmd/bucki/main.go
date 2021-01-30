package main

import (
	"bucki/internal/config"
	"bucki/internal/exporter"
	"bucki/internal/version"
	"log"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var cfg config.Config

	app := kingpin.New("bucki_exporter", "microprofile health exporter")
	app.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9889").StringVar(&cfg.Port)
	app.Flag("web.metrics-path", "").Default("/metrics").StringVar(&cfg.MetricsPath)
	app.Flag("client.timeout", "http client timeout to stop fetching URLs").Default("5").IntVar(&cfg.ClientTimeout)
	app.Flag("config.path", "path to config file").Default("configs/bucki.yml").StringVar(&cfg.ConfigPath)
	app.Flag("only.bucki-metrics", "export only collect microprofile health metrics").BoolVar(&cfg.BuckiMetrics)
	app.HelpFlag.Short('h')
	app.Version(version.Print("bucki_exporter"))

	_, err := app.Parse(os.Args[1:])

	if err != nil {
		log.Fatal(err)
		app.Usage(os.Args[1:])
		os.Exit(2)
	}

	config.ReadConfig(&cfg)

	exporter.Start(&cfg)
}
