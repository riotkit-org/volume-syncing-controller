package http

import (
	"encoding/json"
	"fmt"
	"github.com/riotkit-org/volume-syncing-operator/pkg/client/clientset/versioned"
	"github.com/riotkit-org/volume-syncing-operator/pkg/server/cache"
	"github.com/riotkit-org/volume-syncing-operator/pkg/server/mutation"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	http2 "net/http"
)

type EndpointServingService struct {
	image string

	riotkitClient *versioned.Clientset
	client        *kubernetes.Clientset
	cache         *cache.Cache
}

func NewEndpointServingService(image string, riotkitClient *versioned.Clientset, client *kubernetes.Clientset, cache *cache.Cache) *EndpointServingService {
	return &EndpointServingService{
		image:         image,
		riotkitClient: riotkitClient,
		client:        client,
		cache:         cache,
	}
}

// ServeHealth returns 200 when things are good
func (c *EndpointServingService) ServeHealth(w http2.ResponseWriter, r *http2.Request) {
	logrus.WithField("uri", r.RequestURI).Debug("healthy")
	fmt.Fprint(w, "OK")
}

// ServeMutatePods is mutating Pods by attaching containers to them depending on the configuration stored in PodFilesystemSync
//                 A Pod could be rejected if its annotation-based configuration is invalid, or the Pod has label that Volume Syncing Operator should be used
//                 but no any definition was matched
func (c *EndpointServingService) ServeMutatePods(w http2.ResponseWriter, r *http2.Request) {
	review, parseErr := ParseAdmissionRequest(r)
	if parseErr != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, parseErr.Error())
		return
	}

	mutator := mutation.NewPodMutator(c.cache, c.riotkitClient)
	patchedPod, originalPod, mutationErr := mutator.ProcessAdmissionRequest(review, c.image)

	if mutationErr != nil {
		logrus.Errorf("Invalid admision request for Pod: %v", mutationErr.Error())
		c.sendJsonResponse(CreateReviewResponse(review.Request, false, 400, mutationErr.Error()), w)
		return
	}

	patch, patchingErr := CreateObjectPatch(originalPod, patchedPod)
	if patchingErr != nil {
		logrus.Errorf("Invalid admision request for Pod - cannot patch: %v", patchingErr.Error())
		c.sendJsonResponse(CreateReviewResponse(review.Request, false, 500, patchingErr.Error()), w)
		return
	}

	logrus.Infof("Mutating Pod in namespace '%s'", patchedPod.Namespace)
	c.sendJsonResponse(CreatePatchReviewResponse(review.Request, patch), w)
}

// ServeInformer acts as an endpoint to inform Volume Syncing Operator about new CRDs (PodFilesystemSync) and as validation endpoint
//               Invalid PodFilesystemSync requests should be blocked before stored by API server
func (c *EndpointServingService) ServeInformer(w http2.ResponseWriter, r *http2.Request) {
	review, parseErr := ParseAdmissionRequest(r)
	if parseErr != nil {
		logrus.Errorf("Invalid informer admission request - parse error: %v", parseErr.Error())
		c.sendJsonResponse(CreateReviewResponse(review.Request, false, 400, parseErr.Error()), w)
		return
	}

	logrus.Debugf("Got informer request: raw=%v, oldRaw=%v", string(review.Request.Object.Raw), string(review.Request.OldObject.Raw))

	filesystemSync, isAdded, err := mutation.ResolvePodFilesystemSync(review.Request)
	if err != nil {
		logrus.Errorf("Invalid informer admission request: %v", err.Error())
		c.sendJsonResponse(CreateReviewResponse(review.Request, false, 400, err.Error()), w)
		return
	}

	if isAdded {
		logrus.Infof("Valid informer admission request, caching PodFilesystemSync")
		c.cache.Add(filesystemSync)
	} else {
		logrus.Infof("Valid informer admission request, deleting PodFilesystemSync")
		c.cache.Delete(filesystemSync)
	}

	c.sendJsonResponse(CreateReviewResponse(review.Request, true, 200, "Valid"), w)
}

// sendJsonResponse Sends a JSON response from object of any type
func (c *EndpointServingService) sendJsonResponse(out interface{}, w http2.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	jsonOutput, err := json.Marshal(out)
	if err != nil {
		e := fmt.Sprintf("Could not parse admission response: %v", err)
		logrus.Error(e)
		http2.Error(w, e, http2.StatusInternalServerError)
		return
	}

	// writes to console
	logrus.Debug("Sending response")
	logrus.Debugf("%s", jsonOutput)

	// writes to HTTP resource
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", jsonOutput)
}
