package reader

import (
	"bucki/internal/config"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptrace"
	"strconv"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/client_golang/prometheus"
)

type apiResponseInfoChecks struct {
	Status string            `json:"status"`
	Name   string            `json:"name"`
	Data   map[string]string `json:"data"`
}

type apiResponseInfo struct {
	Status string `json:"status"`
	Checks []apiResponseInfoChecks
}

type foundChecksData struct {
	NumberData []string
	StringData map[string]string
}

type foundChecks struct {
	Checks map[string]foundChecksData
}

type healthPoint struct {
	Name      string
	URL       string
	Reachable bool
	FC        foundChecks
}

// HealthPoints is used for comparsion which elements set to DOWN (0)
type HealthPoints struct {
	HPS map[string]healthPoint
}

var (
	buckiHTTPSuccess = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bucki_http_success",
		Help: "Current request succeeded (1) or not (0)",
	},
		[]string{"url", "name"},
	)
	buckiHTTPduration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bucki_http_duration",
		Help: "Duration of http request in ms",
	},
		[]string{"url", "name"},
	)
	buckiHTTPresponseCode = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bucki_http_response_code",
		Help: "response code of http request",
	},
		[]string{"url", "name"},
	)
	buckiOverallState = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bucki_overall_state",
		Help: "microprofile health overall status UP (1) or DOWN (0) else (3)",
	},
		[]string{"url", "name"},
	)
	buckiCheckState = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bucki_check_state",
		Help: "microprofile health check status UP (1) or DOWN (0) else (3)",
	},
		[]string{"url", "name", "check"},
	)
	buckiCheckDataNumber = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bucki_check_data_number",
		Help: "microprofile health check data as gauge (if it is a number)",
	},
		[]string{"url", "name", "check", "data"},
	)
	buckiCheckDataString = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "bucki_check_data_string",
		Help: "microprofile health check data value as label (if it is a string, always (1))",
	},
		[]string{"url", "name", "check", "data", "value"},
	)
)

// ProceedUrls iterate over Urls from config
func ProceedUrls(cfg *config.Config, hps *HealthPoints, registry *prometheus.Registry) {
	registry.MustRegister(
		buckiHTTPSuccess,
		buckiHTTPduration,
		buckiHTTPresponseCode,
		buckiOverallState,
		buckiCheckState,
		buckiCheckDataNumber,
		buckiCheckDataString,
	)
	if len(hps.HPS) == 0 {
		hps.HPS = make(map[string]healthPoint)
	}
	checked := make(chan healthPoint)
	healthPointCounter := 0

	for _, url := range cfg.Urls {
		go readWebLink(url.Address, url.Name, cfg.ClientTimeout, checked)
	}

	var hP healthPoint
	for {
		hP = <-checked

		if _, ok := hps.HPS[hP.Name]; ok {
			if !cmp.Equal(hP, hps.HPS[hP.Name]) {
				if !hP.Reachable {

					// Schleifen konsolidieren Data in Function auslagern und delete mit go auf rufen wait? damit keine Daten inkonsistent werden
					for check, data := range hps.HPS[hP.Name].FC.Checks {
                        buckiCheckState.WithLabelValues(hP.URL, hP.Name, check).Set(0)
						for _, nd := range data.NumberData {
							buckiCheckDataNumber.DeleteLabelValues(hP.URL, hP.Name, check, nd)
						}
						for sd, v := range data.StringData {
							buckiCheckDataString.DeleteLabelValues(hP.URL, hP.Name, check, sd, v)
						}
					}
				} else {
					for check, data := range hps.HPS[hP.Name].FC.Checks {
						if _, exist := hP.FC.Checks[check]; !exist {
							buckiCheckState.DeleteLabelValues(hP.URL, hP.Name, check)
							for _, nd := range data.NumberData {
								buckiCheckDataNumber.DeleteLabelValues(hP.URL, hP.Name, check, nd)
							}
							for sd, v := range data.StringData {
								buckiCheckDataString.DeleteLabelValues(hP.URL, hP.Name, check, sd, v)
							}
						}
					}
				}
				hps.HPS[hP.Name] = hP
			}
		} else {
			hps.HPS[hP.Name] = hP
		}

		healthPointCounter++

		if healthPointCounter == len(cfg.Urls) {
			break
		}
	}
	close(checked)
}

// ReadWebLink read simple weblink
func readWebLink(url string, name string, timeout int, checkChan chan healthPoint) {

	var ari apiResponseInfo
	var dnsStart, connected time.Time
	var totalTime float64
	var responseCode float64
	var overallState float64
	var checkState float64
	var hp healthPoint

	hp.Name = name
	hp.URL = url
	hp.Reachable = true
	hp.FC.Checks = make(map[string]foundChecksData)

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
		buckiHTTPSuccess.With(prometheus.Labels{"url": url, "name": name}).Set(0)
		buckiOverallState.WithLabelValues(url, name).Set(0)
		buckiHTTPresponseCode.DeleteLabelValues(url, name)
		hp.Reachable = false
		checkChan <- hp
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	requestDone := time.Now()
	buckiHTTPSuccess.With(prometheus.Labels{"url": url, "name": name}).Set(1)
	responseCode = float64(resp.StatusCode)
    buckiHTTPresponseCode.With(prometheus.Labels{"url": url, "name": name}).Set(responseCode)


	if dnsStart.IsZero() {
		totalTime = float64(requestDone.Sub(connected) / time.Millisecond)
	} else {
		totalTime = float64(requestDone.Sub(dnsStart) / time.Millisecond)
	}

	buckiHTTPduration.With(prometheus.Labels{"url": url, "name": name}).Set(totalTime)

	if err != nil {
		log.Println(err)
		return
	}

	jsonErr := json.Unmarshal(body, &ari)
	if jsonErr != nil {
		log.Println(jsonErr)
		return
	}

	switch ari.Status {
	case "DOWN":
		overallState = 0
	case "UP":
		overallState = 1
	default:
		overallState = 3
	}
	buckiOverallState.With(prometheus.Labels{"url": url, "name": name}).Set(overallState)

	for _, check := range ari.Checks {
		var fcd foundChecksData
		fcd.StringData = make(map[string]string)

		switch check.Status {
		case "DOWN":
			checkState = 0
		case "UP":
			checkState = 1
		default:
			checkState = 3
		}

		buckiCheckState.With(prometheus.Labels{"url": url, "name": name, "check": check.Name}).Set(checkState)
// go routine ? bei vielen Checks besser data über Channel zurückgeben
		if len(check.Data) > 0 {
			for data, value := range check.Data {
				if valueFloat, errCast := strconv.ParseFloat(value, 64); errCast == nil {
					buckiCheckDataNumber.With(prometheus.Labels{"url": url, "name": name, "check": check.Name, "data": data}).Set(valueFloat)
					fcd.NumberData = append(fcd.NumberData, data)
				} else {
					buckiCheckDataString.With(prometheus.Labels{"url": url, "name": name, "check": check.Name, "data": data, "value": value}).Set(1)
					fcd.StringData[data] = value
				}
			}
		}
		hp.FC.Checks[check.Name] = fcd
	}

	checkChan <- hp
}

// func evaluateCheck
// func evaluateData
