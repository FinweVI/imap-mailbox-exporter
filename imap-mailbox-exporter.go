package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/mxk/go-imap/imap"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

type imapState struct {
	msgCount int
	unseen   int
	up       int
}

type imapExporter struct {
	mailserver       string
	username         string
	password         string
	mailbox          string
	minQueryInterval time.Duration
	lastQuery        time.Time
	lastState        imapState
	mutex            sync.Mutex

	up       *prometheus.Desc
	msgCount prometheus.Gauge
	unseen   prometheus.Gauge
}

func newExporter(mailserver, username, password string, mailbox string, minQueryInterval time.Duration) *imapExporter {
	return &imapExporter{
		mailserver:       mailserver,
		username:         username,
		password:         password,
		mailbox:          mailbox,
		minQueryInterval: minQueryInterval,

		up: prometheus.NewDesc(
			prometheus.BuildFQName("imap", "", "up"),
			"Could the IMAP server be reached",
			nil,
			nil),
		msgCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "imap",
			Name:      "total_messages_count",
			Help:      "Current number of messages in mailbox",
		}),
		unseen: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "imap",
			Name:      "unseen_message",
			Help:      "Sequence number of the first unseen message",
		}),
	}
}

func (exp *imapExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- exp.up
	exp.msgCount.Describe(ch)
	exp.unseen.Describe(ch)
}

func (exp *imapExporter) queryImapServer() imapState {
	state := exp.lastState
	exp.lastQuery = time.Now()

	var (
		client *imap.Client
		err    error
	)

	// Connect to the server
	client, err = imap.Dial(exp.mailserver)
	if err != nil {
		state.up = 0
		return state
	}

	// Remember to log out and close the connection when finished
	defer client.Logout(30 * time.Second)

	// Enable encryption, if supported by the server
	if client.Caps["STARTTLS"] {
		client.StartTLS(nil)
	} else {
		log.Fatal("IMAP server does not support encryption!")
	}

	// Authenticate
	if client.State() != imap.Login {
		log.Fatal("IMAP server in wrong state for Login!")
	}
	_, err = client.Login(exp.username, exp.password)
	if err != nil {
		log.Fatal(err)
	}

	// Open a mailbox read-only (synchronous command - no need for imap.Wait)
	client.Select(exp.mailbox, true)

	state.up = 1
	state.msgCount = int(client.Mailbox.Messages)
	state.unseen = int(client.Mailbox.Unseen)

	return state
}

func (exp *imapExporter) collect(ch chan<- prometheus.Metric) error {

	state := exp.lastState
	if time.Since(exp.lastQuery) >= exp.minQueryInterval {
		state = exp.queryImapServer()
		exp.lastState = state
	}

	exp.msgCount.Set(float64(state.msgCount))
	exp.msgCount.Collect(ch)

	exp.unseen.Set(float64(state.unseen))
	exp.unseen.Collect(ch)

	ch <- prometheus.MustNewConstMetric(exp.up, prometheus.GaugeValue, float64(state.up))

	return nil
}

func (exp *imapExporter) Collect(ch chan<- prometheus.Metric) {
	exp.mutex.Lock() // To protect metrics from concurrent collects.
	defer exp.mutex.Unlock()
	if err := exp.collect(ch); err != nil {
		log.Fatal("Scraping failure!")
	}
	return
}

var (
	imapServer   = flag.String("imap.server", os.Getenv("IMAP_SERVER"), "IMAP server to query")
	imapUsername = flag.String("imap.username", os.Getenv("IMAP_USERNAME"), "IMAP username for login")
	imapPassword = flag.String("imap.password", os.Getenv("IMAP_PASSWORD"), "IMAP password for login")
	imapMailbox  = flag.String("imap.mailbox", os.Getenv("IMAP_MAILBOX"), "IMAP mailbox to query")
	imapInterval = flag.String("imap.query.interval", os.Getenv("IMAP_QUERY_INTERVAL"), "Minimum interval between queries to IMAP server in seconds")

	listenAddress   = flag.String("listen.address", os.Getenv("LISTEN_ADDRESS"), "Address to listen on for web interface and telemetry")
	metricsEndpoint = flag.String("metrics.endpoint", os.Getenv("METRICS_ENDPOINT"), "Path under which to expose metrics")
)

func main() {
	flag.Parse()

	if *imapServer == "" {
		log.Fatal("Missing IMAP server configuration")
	}
	if *imapUsername == "" {
		log.Fatal("Missing IMAP username configuration")
	}
	if *imapPassword == "" {
		log.Fatal("Missing IMAP password configuration")
	}

	if *imapMailbox == "" {
		*imapMailbox = "INBOX"
	}
	if *imapInterval == "" {
		*imapInterval = "120"
	}
	if *listenAddress == "" {
		*listenAddress = "127.0.0.1:9117"
	}
	if *metricsEndpoint == "" {
		*metricsEndpoint = "/metrics"
	}

	imapIntervali, err := strconv.Atoi(*imapInterval)
	if err != nil {
		log.Fatalf("Invalid query interval: %s", *imapInterval)
	}
	imapIntervald := time.Duration(imapIntervali) * time.Second

	exporter := newExporter(*imapServer, *imapUsername, *imapPassword, *imapMailbox, imapIntervald)
	prometheus.MustRegister(exporter)

	http.Handle(*metricsEndpoint, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>IMAP Mailbox Exporter</title></head>
             <body>
             <h1>IMAP Mailbox Exporter</h1>
             <p><a href='` + *metricsEndpoint + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	log.Infof("Exporter listening on %s", *listenAddress)

	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
