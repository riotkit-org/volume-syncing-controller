package mutation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/riotkit-org/volume-syncing-operator/pkg/apis/riotkit.org/v1alpha1"
	"github.com/riotkit-org/volume-syncing-operator/pkg/server/context"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"strings"
	"text/template"
)

func ResolvePod(a *admissionv1.AdmissionRequest) (*corev1.Pod, error) {
	if a.Kind.Kind != "Pod" {
		return nil, fmt.Errorf("only Pods are supported here, got request type: %v", a.Kind.Kind)
	}

	logrus.Debugf("Processing request: %v", string(a.Object.Raw))

	p := corev1.Pod{}
	if err := json.Unmarshal(a.Object.Raw, &p); err != nil {
		return nil, err
	}

	// fix: Missing namespace in case of scoped call by controllers like ReplicaSet/Deployment
	if p.ObjectMeta.Namespace == "" && a.Namespace != "" {
		p.ObjectMeta.Namespace = a.Namespace
	}

	if !isPodToBeProcessed(&p) {
		return nil, fmt.Errorf("only Pods labelled with '%s' can be processed", context.LabelIsEnabled)
	}
	return &p, nil
}

func ResolvePodFilesystemSync(a *admissionv1.AdmissionRequest) (*v1alpha1.PodFilesystemSync, bool, error) {
	if a.Kind.Kind != "PodFilesystemSync" {
		return nil, false, fmt.Errorf("only PodFilesystemSync definitions are supported here, got request type: %v", a.Kind.Kind)
	}

	// object could be CREATED/UPDATED or DELETED
	var objectRaw []byte
	var isAdded bool
	if a.Operation == admissionv1.Delete {
		objectRaw = a.OldObject.Raw
		isAdded = false
	} else {
		objectRaw = a.Object.Raw
		isAdded = true
	}

	p := v1alpha1.PodFilesystemSync{}
	if err := json.Unmarshal(objectRaw, &p); err != nil {
		return nil, isAdded, errors.Wrapf(err, "Cannot unmarshal request object: %v", a.Object.Raw)
	}
	if p.ObjectMeta.Namespace == "" && a.Namespace != "" {
		p.ObjectMeta.Namespace = a.Namespace
	}
	return &p, isAdded, nil
}

func isPodToBeProcessed(pod *corev1.Pod) bool {
	if val, exists := pod.Labels[context.LabelIsEnabled]; exists && strings.ToLower(strings.TrimSpace(val)) == "true" {
		return true
	}
	return false
}

// VerifySecrets is performing basic checks on `kind: Secret` - it does not check existence of a `kind: Secret` so this could be delegated to API server
func VerifySecrets(syncDefinition *v1alpha1.PodFilesystemSync, namespace string) error {
	if len(syncDefinition.Spec.EnvFromSecrets) > 0 {
		for _, secret := range syncDefinition.Spec.EnvFromSecrets {

			// [!!!] NOTICE: Pod cannot mount secrets across namespaces. Second reason is security - namespaced Pod should not be able to read secrets across cluster.
			if secret.Namespace != namespace {
				return errors.Errorf("Refernced secret '%v' is in different namespace than Pod can mount", secret.Name)
			}
		}
	}

	return nil
}

// ResolveTemplatedEnvironment is creating a map of environment variables with addition of template parsing
func ResolveTemplatedEnvironment(pod *corev1.Pod, syncDefinition *v1alpha1.PodFilesystemSync) (map[string]string, error) {
	var processed map[string]string

	for k, v := range syncDefinition.Spec.Env {
		processedVal, err := processTemplate(v, k, pod)
		if err != nil {
			return map[string]string{}, errors.Wrapf(err, "Cannot process template '%s' in environment variable '%s'", v, k)
		}
		processed[k] = processedVal
	}

	return processed, nil
}

// processTemplate is replacing {{ }} syntax with real values
func processTemplate(envString string, envName string, pod *corev1.Pod) (string, error) {
	t, parseErr := template.New(envName).Parse(envString)
	if parseErr != nil {
		return "", parseErr
	}

	// variables accessible inside a template
	insideVars := map[string]interface{}{}
	insideVars["pod"] = pod

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, insideVars); err != nil {
		return "", err
	}
	return tpl.String(), nil
}
