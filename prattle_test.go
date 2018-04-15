package prattle

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gojekfarm/prattle/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPrattleWithSingleNode(t *testing.T) {
	consul, err := NewConsulClient("127.0.0.1:8500")
	assert.NoError(t, err)
	testService := httptest.NewServer(
		http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
			responseWriter.WriteHeader(200)
		}))
	testServiceAddress := testService.Listener.Addr().String()
	discovery := config.Discovery{
		Name:               "test-prattle-service-02",
		Address:            testServiceAddress,
		TTL:                "10s",
	}
	prattle, err := NewPrattle(consul, 9000, discovery)
	require.NoError(t, err)
	prattle.Shutdown()
}

func TestPrattleWithMoreThanOneNode(t *testing.T) {
	consul, err := NewConsulClient("127.0.0.1:8500")
	require.NoError(t, err)
	discoveryOne := config.Discovery{
		Name:               "test-service-01",
		Address:            "0.0.0.0:9000",
		TTL:                "10s",
	}
	discoveryTwo := config.Discovery{
		Name:               "test-service-01",
		Address:            "0.0.0.0:9001",
		TTL:                "10s",
	}
	prattleOne, errOne := NewPrattle(consul, 9000, discoveryOne)
	time.Sleep(1 * time.Second)
	prattleTwo, errTwo := NewPrattle(consul, 9001, discoveryTwo)
	time.Sleep(1 * time.Second)
	require.NoError(t, errOne)
	require.NoError(t, errTwo)
	defer prattleOne.Shutdown()
	defer prattleTwo.Shutdown()
	assert.Equal(t, prattleOne.Members(), prattleTwo.Members())
	assert.Equal(t, int(prattleOne.members.LocalNode().Port), 9000)
	assert.Equal(t, int(prattleTwo.members.LocalNode().Port), 9001)
	assert.Equal(t, 2, prattleOne.members.NumMembers())
	assert.Equal(t, 2, prattleOne.broadcasts.NumNodes())
}
