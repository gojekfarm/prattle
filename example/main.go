package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gojekfarm/prattle"
	"github.com/gojekfarm/prattle/config"
	"github.com/gojekfarm/prattle/registry/consul"
)

var (
	// TODO: find out if they can be made local.
	key   string
	value string

	rpcPort  = flag.Int("rpc-port", 0, "RPC port")
	httpPort = flag.Int("http-port", 0, "http port")
)

type keyValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func addKeyHandler(p *prattle.Prattle) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
		}
		var data keyValue
		json.Unmarshal(body, &data)
		p.Set(data.Key, data.Value)
		json.NewEncoder(w).Encode(data)
	}
}

func clusterHealthHandler(p *prattle.Prattle) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(p.Members())
	}
}

func getValueHandler(p *prattle.Prattle) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		val, _ := p.Get(r.FormValue("key"))
		json.NewEncoder(w).Encode(val)
	}
}

func main() {
	flag.Parse()

	discovery := config.Discovery{
		TTL:                "10s",
		HealthEndpoint:     "http://localhost:" + strconv.FormatInt(int64(*httpPort), 10),
		HealthPingInterval: "10s",
		Address:            "localhost",
		Name:               "Test",
		Port:               *rpcPort,
		ConsulURL:          "http://localhost:8500/",
	}
	consulClient := consul.NewClient("http://localhost:8500/", &http.Client{}, discovery)
	prattle, err := prattle.NewPrattle(consulClient, *rpcPort)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/_health", clusterHealthHandler(prattle))
	http.HandleFunc("/get", getValueHandler(prattle))
	http.HandleFunc("/set", addKeyHandler(prattle))
	fmt.Printf("Listening on :%d\n", *httpPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), nil); err != nil {
		fmt.Println(err)
	}
}
