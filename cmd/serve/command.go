package serve

import (
	"fmt"
	"github.com/pkg/errors"
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

	DefaultImage       string
	DefaultGitUsername string
	DefaultGitToken    string

	client *kubernetes.Clientset
}

func (c *Command) Run() error {
	c.setLogger()
	c.client = initClient()

	// handle our core application
	http.HandleFunc("/mutate-pods", c.ServeMutatePods)
	http.HandleFunc("/health", c.ServeHealth)

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

// ServeHealth returns 200 when things are good
func (c *Command) ServeHealth(w http.ResponseWriter, r *http.Request) {
	logrus.WithField("uri", r.RequestURI).Debug("healthy")
	fmt.Fprint(w, "OK")
}

func (c *Command) ServeMutatePods(w http.ResponseWriter, r *http.Request) {

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

func initClient() *kubernetes.Clientset {
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

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientSet
}
