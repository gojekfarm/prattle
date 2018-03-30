package prattle

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojekfarm/prattle/config"
	"github.com/gojekfarm/prattle/registry/consul"
)

var discovery = config.Discovery{
	TTL:                "10s",
	HealthEndpoint:     "http://localhost:3000/",
	HealthPingInterval: "10s",
	Address:            "http://localhost",
	Name:               "Test",
	Port:               1000,
	ConsulURL:          "http://localhost:8500",
}

func TestPrattleWithMoreThanOneNode(t *testing.T) {
	consulServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fileContents, err := ioutil.ReadFile("./fixtures/healthy_nodes_response.json")
		require.NoError(t, err)
		rw.Write(fileContents)
		rw.WriteHeader(200)
	}))

	discovery.ConsulURL = consulServer.URL + "/"
	client := consul.NewClient(discovery.ConsulURL, &http.Client{}, discovery)
	prattleOne, errOne := NewPrattle(client, 8080)
	prattleTwo, errTwo := NewPrattle(client, 8081)
	assert.Nil(t, errOne)
	assert.Nil(t, errTwo)
	assert.Equal(t, prattleOne.Members(), prattleTwo.Members())
	assert.Equal(t, int(prattleOne.members.LocalNode().Port), 8080)
	assert.Equal(t, int(prattleTwo.members.LocalNode().Port), 8081)
	assert.Equal(t, 2, prattleOne.members.NumMembers())
	assert.Equal(t, 2, prattleOne.broadcasts.NumNodes())
	assert.Equal(t, 3, prattleOne.broadcasts.RetransmitMult)
	defer prattleOne.Shutdown()
	defer prattleTwo.Shutdown()
}

func TestNewPrattleWhenMemberAddressIsNotInUse(t *testing.T) {
	consulServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fileContents, err := ioutil.ReadFile("./fixtures/healthy_nodes_response.json")
		require.NoError(t, err)
		rw.Write(fileContents)
		rw.WriteHeader(200)
	}))

	discovery.ConsulURL = consulServer.URL + "/"
	client := consul.NewClient(discovery.ConsulURL, &http.Client{}, discovery)
	prattle, err := NewPrattle(client, 8080)
	defer prattle.Shutdown()
	assert.Nil(t, err)
	assert.Equal(t, 1, prattle.members.NumMembers())
	assert.Equal(t, 1, prattle.broadcasts.NumNodes())
	assert.NotNil(t, prattle.database.connection)
}

func TestPrattleWhenMemberAddressIsAlreadyInUse(t *testing.T) {
	consulServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fileContents, err := ioutil.ReadFile("./fixtures/healthy_nodes_response.json")
		require.NoError(t, err)
		rw.Write(fileContents)
		rw.WriteHeader(200)
	}))

	discovery.ConsulURL = consulServer.URL + "/"
	client := consul.NewClient(discovery.ConsulURL, &http.Client{}, discovery)
	prattle, errOne := NewPrattle(client, 8080)
	defer prattle.Shutdown()
	_, errTwo := NewPrattle(client, 8080)
	assert.Nil(t, errOne)
	assert.NotNil(t, errTwo)
}

func TestGetWhenKeyIsNotFound(t *testing.T) {
	consulServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fileContents, err := ioutil.ReadFile("./fixtures/healthy_nodes_response.json")
		require.NoError(t, err)
		rw.Write(fileContents)
		rw.WriteHeader(200)
	}))

	discovery.ConsulURL = consulServer.URL + "/"
	client := consul.NewClient(discovery.ConsulURL, &http.Client{}, discovery)
	prattle, _ := NewPrattle(client, 8080)
	value, found := prattle.Get("ping")
	assert.False(t, found)
	assert.Equal(t, value, nil)
	defer prattle.Shutdown()
}

func TestGetWhenKeyIsFound(t *testing.T) {
	consulServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fileContents, err := ioutil.ReadFile("./fixtures/healthy_nodes_response.json")
		require.NoError(t, err)
		rw.Write(fileContents)
		rw.WriteHeader(200)
	}))

	discovery.ConsulURL = consulServer.URL + "/"
	client := consul.NewClient(discovery.ConsulURL, &http.Client{}, discovery)
	prattle, _ := NewPrattle(client, 8080)
	prattle.Set("ping", "pong")
	value, found := prattle.Get("ping")
	assert.True(t, found)
	assert.Equal(t, "pong", value)
	defer prattle.Shutdown()
}

func TestSetWhenKeyAlreadyExist(t *testing.T) {
	consulServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fileContents, err := ioutil.ReadFile("./fixtures/healthy_nodes_response.json")
		require.NoError(t, err)
		rw.Write(fileContents)
		rw.WriteHeader(200)
	}))

	discovery.ConsulURL = consulServer.URL + "/"
	client := consul.NewClient(discovery.ConsulURL, &http.Client{}, discovery)
	prattle, _ := NewPrattle(client, 8080)
	prattle.Set("ping", "pong")
	value, _ := prattle.Get("ping")
	assert.Equal(t, "pong", value)
	prattle.Set("ping", "pong2")
	newValue, _ := prattle.Get("ping")
	assert.Equal(t, "pong2", newValue)
	defer prattle.Shutdown()
}
