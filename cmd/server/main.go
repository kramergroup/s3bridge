package main

import (
	"flag"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	bridge "internal/bridge"
	util "internal/util"
)

var (
	s3bridge    bridge.S3Bridge // s3 bridge stuct
	port        int             // listening port for service
	presignport int             // listening port for presign service
	loglevel    string
	LOGLEVELS   = []string{"DEBUG", "INFO", "WARN", "ERROR"}
)

func main() {

	// Read configuration from command line or environment variables
	flag.StringVar(&s3bridge.Bucket, "bucket", util.LookupEnvOrString("BUCKET", ""), "s3 source bucket")
	flag.StringVar(&s3bridge.S3AccessKey, "aws-access-key-id", util.LookupEnvOrString("AWS_ACCESS_KEY_ID", ""), "aws access key id")
	flag.StringVar(&s3bridge.S3SecretKey, "aws-secret-access-key", util.LookupEnvOrString("AWS_SECRET_ACCESS_KEY", ""), "aws secret access key")
	flag.StringVar(&s3bridge.Endpoint, "endpoint", util.LookupEnvOrString("ENDPOINT", ""), "s3 endpoint url")
	flag.StringVar(&s3bridge.Region, "aws-region", util.LookupEnvOrString("AWS_REGION", "us-west-2"), "aws region")
	flag.DurationVar(&s3bridge.ExpiryTime, "expiry-time", util.LookupEnvOrDuration("EXIRY_TIME", 2*time.Hour), "pre-signed url expiry time")
	flag.IntVar(&port, "proxy-port", util.LookupEnvOrInt("PROXY-PORT", 80), "listening port for proxy service")
	flag.IntVar(&presignport, "presign-port", util.LookupEnvOrInt("PRESIGN-PORT", 8080), "listening port for presigned service")
	flag.StringVar(&loglevel, "loglevel", util.LookupEnvOrString("LOGLEVEL", "INFO"), "log level (DEBUG, INFO, WARN, ERROR)")
	flag.Parse()

	// validate config
	if err := s3bridge.Validate(); err != nil {
		log.Fatal(err)
	}

	loglevel = strings.ToUpper(loglevel)
	if !slices.Contains(LOGLEVELS, loglevel) {
		log.Fatal("loglevel %s not known", loglevel)
	}

	// Setting loglevels
	switch loglevel {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	var wg sync.WaitGroup

	// Start proxy presigning service
	if presignport != 0 {
		wg.Add(1)
		go func() {

			defer wg.Done()
			mux := http.NewServeMux()
			// Creating the html endpoints
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

				u, err := s3bridge.GetRequestURL(r.URL.Path)
				if err != nil {
					log.Error(err)
					http.NotFound(w, r)
				} else {
					http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
				}

			})

			// Listen and serve
			log.Debug("start listening on port ", port)
			if err := http.ListenAndServe(":"+strconv.Itoa(presignport), mux); err != nil {
				log.Fatal(err)
			}

		}()

		if port != 0 {
			wg.Add(1)
			go func() {

				defer wg.Done()

				mux := http.NewServeMux()
				mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

					name := strings.TrimLeft(r.URL.Path, "/")

					// --------------------------------------------------------------
					// This implementation reads entire objects and serves them. This
					// can be quite resource intensive for large objects, because the
					// entire object is read into memory before passing on

					// ctx := r.Context()

					// buf := manager.NewWriteAtBuffer([]byte{})
					// if err := s3bridge.StreamObject(name, ctx, buf); err != nil {
					// 	http.NotFound(w, r)
					// } else {
					// 	w.Write(buf.Bytes())
					// }

					// --------------------------------------------------------------
					// This implementation exploits the range feature of HTTP2 requests, which
					// allows to read only parts of an object at a time. This should reduce
					// memory pressure substantially, because only parts of large objects
					// are held in memory at a time
					rs, err := s3bridge.ReadSeeker(name, r.Context())
					if err != nil {
						http.Error(w, "error accessing asset", http.StatusInternalServerError)
					} else {
						http.ServeContent(w, r, path.Base(name), time.Now(), rs)
					}

				})

				// Listen and serve
				log.Debug("start listening on port ", port)
				if err := http.ListenAndServe(":"+strconv.Itoa(port), mux); err != nil {
					log.Fatal(err)
				}
			}()
		}
	}

	wg.Wait()
	log.Debug("finishing")
}
