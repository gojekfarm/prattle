package consul

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThatItRegistersSuccessfullyWhenRegistrationResponseIsOK(t *testing.T) {
	testserver := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.WriteHeader(200)
	}))
	consulURL := testserver.URL + "/"
	err := NewClient(consulURL, &http.Client{}).Register()
	assert.NoError(t, err)
}

func TestThatItDoesntRegisterTheServiceSuccessfullyWhenResponseIsNotOK(t *testing.T) {
	testserver := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.WriteHeader(400)
	}))
	consulURL := testserver.URL + "/"
	err := NewClient(consulURL, &http.Client{}).Register()
	assert.Error(t, err)
}
