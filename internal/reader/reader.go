package reader

import (
	"bucki/internal/config"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

//type dlUrls struct {
//	URL string `json:"browser_download_url"`
//}

type apiResponseInfo struct {
	Release string `json:"tag_name"`
	//Urls []dlUrls `json:"assets"`
}

var (
	buckiHTTPSuccess = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bucki_http_success",
		Help: "Current request succeeded (1) or not (0)",
	},
		[]string{"url"},
	)
	buckiGitRelease = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bucki_git_release",
		Help: "Current release in labels, if found (1) else (0)",
	},
		[]string{"url", "release"},
	)
	buckiHTTPduration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bucki_http_duration",
		Help: "Duration of http request in ms",
	},
		[]string{"url"},
	)
	buckiHTTPresponseCode = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bucki_http_response_code",
		Help: "response code of http request",
	},
		[]string{"url"},
	)
)

// ProceedUrls iterate over Urls from config
func ProceedUrls(cfg *config.Config, registry *prometheus.Registry) {
	registry.MustRegister(
		buckiHTTPSuccess,
		buckiGitRelease,
		buckiHTTPduration,
		buckiHTTPresponseCode,
	)

	var wg sync.WaitGroup
	for _, url := range cfg.Urls {
		wg.Add(1)
		go readWebLink(url, cfg.ClientTimeout, &wg)
	}
	wg.Wait()
}

// ReadWebLink read simple weblink
func readWebLink(url string, timeout int,wg *sync.WaitGroup) {
	defer wg.Done()

	var ari apiResponseInfo
	var dnsStart, connected time.Time
	var totalTime float64
	var foundRelease float64 = 0
	var responseCode float64
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	req, _ := http.NewRequest(http.MethodGet, url, nil)

	req.Header.Set("User-Agent", "bucki-exporter")

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { dnsStart = time.Now() },

		GotConn: func(_ httptrace.GotConnInfo) { connected = time.Now() },
	}

	req = req.WithContext(httptrace.WithClientTrace(context.Background(), trace))

	resp, getErr := client.Do(req)

	if getErr != nil {
		log.Println(getErr)
		buckiHTTPSuccess.With(prometheus.Labels{"url": url}).Set(0)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	requestDone := time.Now()
	responseCode = float64(resp.StatusCode)
	buckiHTTPSuccess.With(prometheus.Labels{"url": url}).Set(1)
	buckiHTTPresponseCode.With(prometheus.Labels{"url": url}).Set(responseCode)

	if dnsStart.IsZero() {
		totalTime = float64(requestDone.Sub(connected) / time.Millisecond)
	} else {
		totalTime = float64(requestDone.Sub(dnsStart) / time.Millisecond)
	}

	buckiHTTPduration.With(prometheus.Labels{"url": url}).Set(totalTime)

	if err != nil {
		log.Println(err)
		return
	}

	jsonErr := json.Unmarshal(body, &ari)
	if jsonErr != nil {
		log.Println(jsonErr)
		return
	}

	if ari.Release != "" {
		foundRelease = 1
	}

	buckiGitRelease.With(prometheus.Labels{"url": url, "release": ari.Release}).Set(foundRelease)
}
