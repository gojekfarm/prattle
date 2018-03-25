package consul

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThatItRegistersSuccessfullyWhenRegistrationRequestIsOK(t *testing.T) {
	testserver := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		responseWriter.WriteHeader(200)
	}))
	consulURL := testserver.URL
	err := NewClient(consulURL).Register()
	assert.NoError(t, err)
}
