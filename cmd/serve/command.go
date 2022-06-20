package serve

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/riotkit-org/volume-syncing-operator/pkg/client/clientset/versioned"
	"github.com/riotkit-org/volume-syncing-operator/pkg/server"
	"github.com/riotkit-org/volume-syncing-operator/pkg/server/cache"
	"github.com/riotkit-org/volume-syncing-operator/pkg/server/mutation"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"
)

type Command struct {
	LogLevel string
	TLS      bool
	LogJSON  bool
	Image    string

	TLSCrtPath string
	TLSKeyPath string

	riotkitClient *versioned.Clientset
	client        *kubernetes.Clientset
	cache         *cache.Cache
}

func (c *Command) Run() error {
	// initialize
	c.setLogger()
	c.riotkitClient, c.client = initClient()

	// populate cache to know about volume syncing configurations created before the application was started
	if err := c.cache.Populate(c.riotkitClient, c.client); err != nil {
		return errors.Wrap(err, "Cannot populate cache")
	}

	// handle our core application
	http.HandleFunc("/mutate-pods", c.serveMutatePods)
	http.HandleFunc("/health", c.serveHealth)
	http.HandleFunc("/inform", c.serveInformer)

	// start the server
	// listens to clear text http on port 8080 unless TLS env var is set to "true"
	if c.TLS {
		cert := "/etc/admission-webhook/tls/tls.crt"
		key := "/etc/admission-webhook/tls/tls.key"
		logrus.Print("Listening on port 4443...")
		return http.ListenAndServeTLS(":4443", cert, key, nil)
	} else {
		logrus.Print("Listening on port 8080...")
		return http.ListenAndServe(":8080", nil)
	}
}

// serveHealth returns 200 when things are good
func (c *Command) serveHealth(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("uri", r.RequestURI).Debug("healthy")
	fmt.Fprint(w, "OK")
}

func (c *Command) serveMutatePods(w http.ResponseWriter, r *http.Request) {
	review, parseErr := server.ParseAdmissionRequest(r)
	if parseErr != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, parseErr.Error())
		return
	}

	mutator := mutation.NewPodMutator(c.cache)
	if err := mutator.ProcessAdmissionRequest(review, c.Image); err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, err.Error())
		return
	}

	c.sendJsonResponse(review, w)
}

// serveInformer updates cache and can act as a validation
func (c *Command) serveInformer(w http.ResponseWriter, r *http.Request) {
	review, parseErr := server.ParseAdmissionRequest(r)
	if parseErr != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, parseErr.Error())
		return
	}

	filesystemSync, err := mutation.ResolvePodFilesystemSync(review.Request)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, parseErr.Error())
		return
	}

	c.cache.Add(*filesystemSync)
	w.WriteHeader(200)
}

func (c *Command) sendJsonResponse(out interface{}, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	jsonOutput, err := json.Marshal(out)
	if err != nil {
		e := fmt.Sprintf("Could not parse admission response: %v", err)
		logrus.Error(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	// writes to console
	logrus.Debug("Sending response")
	logrus.Debugf("%s", jsonOutput)

	// writes to HTTP resource
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", jsonOutput)
}

// setLogger sets the logger using env vars, it defaults to text logs on
// debug level unless otherwise specified
func (c *Command) setLogger() {
	lvl, parseErr := logrus.ParseLevel(c.LogLevel)
	if parseErr != nil {
		logrus.Fatalf("Cannot parse log level: %v", parseErr)
	}
	logrus.SetLevel(lvl)
	logrus.Printf("Setting log level=%v", lvl.String())

	if c.LogJSON {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
}

func initClient() (*versioned.Clientset, *kubernetes.Clientset) {
	kubeConfig := os.Getenv("HOME") + "/.kube/config"
	if os.Getenv("KUBECONFIG") != "" {
		kubeConfig = os.Getenv("KUBECONFIG")
	}
	if _, err := os.Stat(kubeConfig); errors.Is(err, os.ErrNotExist) {
		kubeConfig = ""
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		panic(err.Error())
	}

	riotkitClientSet, rktErr := versioned.NewForConfig(config)
	if rktErr != nil {
		panic(rktErr.Error())
	}
	kubeClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return riotkitClientSet, kubeClientSet
}
