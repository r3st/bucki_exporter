package exporter

import (
	"bucki/internal/config"
	"bucki/internal/reader"
	"fmt"
	"html"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	buckiScraped = promauto.NewCounter(prometheus.CounterOpts{
		Name: "bucki_scrape_count",
		Help: "Total number bucki exporter scraped",
	})
)

// Start exporter webserver
func Start(cfg *config.Config) {

	router := mux.NewRouter()
	router.HandleFunc(cfg.MetricsPath, metricsHandler(cfg)).Methods(http.MethodGet)
	router.HandleFunc("/", rootHandler(cfg)).Methods(http.MethodGet)

	http.ListenAndServe(cfg.Port, router)
}

func metricsHandler(cfg *config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buckiScraped.Inc()

		registry := prometheus.NewRegistry()
		registry.MustRegister(
			prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
			prometheus.NewGoCollector(),
			buckiScraped,
		)
		reader.ProceedUrls(cfg.Urls, registry)

		promhttp.HandlerFor(
			registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError},
		).ServeHTTP(w, r)
	})
}

func rootHandler(cfg *config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<hml>
		<head><title>Bucki Exporter</title></head>
		<body>
		<h1>Bucki Exporter</h1>
		`))
		fmt.Fprintf(w, "<p><a href=\"%s\">Metrics</a></p>", html.EscapeString(cfg.MetricsPath))
		fmt.Fprint(w, "<h2>Current checked URLs</h2>")

		for _, url := range cfg.Urls {
           fmt.Fprintf(w, "<li>%s</li>", url)
		}

		w.Write([]byte(`</body>
		</html>`))
	})
}
