package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wI2L/jsondiff"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

// ParseAdmissionRequest extracts an AdmissionReview from an http.Request if possible
func ParseAdmissionRequest(r *http.Request) (*admissionv1.AdmissionReview, error) {
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("Content-Type: %q should be %q",
			r.Header.Get("Content-Type"), "application/json")
	}

	if r.Body == nil {
		return nil, fmt.Errorf("admission request body is empty")
	}

	bodybuf := new(bytes.Buffer)
	bodybuf.ReadFrom(r.Body)
	body := bodybuf.Bytes()

	if len(body) == 0 {
		return nil, fmt.Errorf("admission request body is empty")
	}

	var a admissionv1.AdmissionReview

	if err := json.Unmarshal(body, &a); err != nil {
		return nil, fmt.Errorf("could not parse admission review request: %v", err)
	}

	if a.Request == nil {
		return nil, fmt.Errorf("admission review can't be used: Request field is nil")
	}

	return &a, nil
}

// CreateReviewResponse is creating Admission Review responses without mutating the input object - this method is used only for validation requests
func CreateReviewResponse(request *admissionv1.AdmissionRequest, allowed bool, httpCode int32, reason string) *admissionv1.AdmissionReview {
	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:     request.UID,
			Allowed: allowed,
			Result: &metav1.Status{
				Code:    httpCode,
				Message: reason,
			},
		},
	}
}

// CreateObjectPatch is creating a JSON patch from two objects
func CreateObjectPatch(source, target interface{}) ([]byte, error) {
	// generate json patch
	patch, err := jsondiff.Compare(source, target)
	if err != nil {
		return nil, err
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return nil, err
	}

	return patchBytes, nil
}

// CreatePatchReviewResponse builds an admission review with given json patch
func CreatePatchReviewResponse(request *admissionv1.AdmissionRequest, patch []byte) *admissionv1.AdmissionReview {
	patchType := admissionv1.PatchTypeJSONPatch

	return &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Response: &admissionv1.AdmissionResponse{
			UID:       request.UID,
			Allowed:   true,
			PatchType: &patchType,
			Patch:     patch,
		},
	}
}
