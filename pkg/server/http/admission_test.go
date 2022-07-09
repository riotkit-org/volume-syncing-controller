package http

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	admissionv1 "k8s.io/api/admission/v1"
	v12 "k8s.io/api/authentication/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"strings"
	"testing"
)

func createFakeReview() admissionv1.AdmissionReview {
	return admissionv1.AdmissionReview{
		TypeMeta: v1.TypeMeta{},
		Request: &admissionv1.AdmissionRequest{
			UID: "161-132-2137-161",
			Kind: v1.GroupVersionKind{
				Group:   "riotkit.org",
				Version: "v1alpha1",
				Kind:    "Application",
			},
			Resource:           v1.GroupVersionResource{},
			SubResource:        "",
			RequestKind:        nil,
			RequestResource:    nil,
			RequestSubResource: "",
			Name:               "",
			Namespace:          "",
			Operation:          "",
			UserInfo:           v12.UserInfo{},
			Object:             runtime.RawExtension{},
			OldObject:          runtime.RawExtension{},
			DryRun:             nil,
			Options:            runtime.RawExtension{},
		},
		Response: nil,
	}
}

func TestParseAdmissionRequest_Success(t *testing.T) {
	fake := createFakeReview()
	fakeB, _ := json.Marshal(fake)

	r := http.Request{}
	r.Header = http.Header{}
	r.Header.Set("Content-Type", "application/json")
	r.Body = ioutil.NopCloser(strings.NewReader(string(fakeB)))

	_, parseErr := ParseAdmissionRequest(&r)
	assert.Nil(t, parseErr)
}

func TestParseAdmissionRequest_InvalidContentType(t *testing.T) {
	fake := createFakeReview()
	fakeB, _ := json.Marshal(fake)

	r := http.Request{}
	r.Header = http.Header{}
	r.Header.Set("Content-Type", "application/html")
	r.Body = ioutil.NopCloser(strings.NewReader(string(fakeB)))

	_, parseErr := ParseAdmissionRequest(&r)
	assert.NotNil(t, parseErr)
}

func TestParseAdmissionRequest_InvalidBody(t *testing.T) {
	r := http.Request{}
	r.Header = http.Header{}
	r.Header.Set("Content-Type", "application/json")

	_, parseErr := ParseAdmissionRequest(&r)
	assert.NotNil(t, parseErr)
}

func TestParseAdmissionRequest_EmptyBody(t *testing.T) {
	r := http.Request{}
	r.Header = http.Header{}
	r.Header.Set("Content-Type", "application/html")
	r.Body = ioutil.NopCloser(strings.NewReader(""))

	_, parseErr := ParseAdmissionRequest(&r)
	assert.NotNil(t, parseErr)
}

func TestParseAdmissionRequest_MisFormattedBody(t *testing.T) {
	r := http.Request{}
	r.Header = http.Header{}
	r.Header.Set("Content-Type", "application/html")
	r.Body = ioutil.NopCloser(strings.NewReader("invalid-formatting-there-not-a-json"))

	_, parseErr := ParseAdmissionRequest(&r)
	assert.NotNil(t, parseErr)
}
