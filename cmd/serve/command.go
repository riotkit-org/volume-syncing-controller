package serve

import (
	"github.com/pkg/errors"
	"github.com/riotkit-org/volume-syncing-controller/pkg/client/clientset/versioned"
	"github.com/riotkit-org/volume-syncing-controller/pkg/server/cache"
	"github.com/riotkit-org/volume-syncing-controller/pkg/server/http"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	netHttp "net/http"
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
	endpoints     *http.EndpointServingService
}

func (c *Command) Run() error {
	// initialize
	c.setLogger()
	c.cache = &cache.Cache{}
	c.riotkitClient, c.client = createKubernetesClients()
	c.endpoints = http.NewEndpointServingService(c.Image, c.riotkitClient, c.client, c.cache)

	// populate cache to know about volume syncing configurations created before the application was started
	if err := c.cache.Populate(c.riotkitClient, c.client); err != nil {
		return errors.Wrap(err, "Cannot populate cache")
	}

	// handle our core application
	netHttp.HandleFunc("/mutate-pods", c.endpoints.ServeMutatePods)
	netHttp.HandleFunc("/health", c.endpoints.ServeHealth)
	netHttp.HandleFunc("/inform", c.endpoints.ServeInformer)

	// start the server
	// listens to clear text http on port 8080 unless TLS env var is set to "true"
	if c.TLS {
		cert := "/etc/admission-webhook/tls/tls.crt"
		key := "/etc/admission-webhook/tls/tls.key"
		logrus.Print("Listening on port 4443...")
		return netHttp.ListenAndServeTLS(":4443", cert, key, nil)
	} else {
		logrus.Print("Listening on port 8080...")
		return netHttp.ListenAndServe(":8080", nil)
	}
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

// createKubernetesClients is creating API client libraries instances
func createKubernetesClients() (*versioned.Clientset, *kubernetes.Clientset) {
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
