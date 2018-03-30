package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gojekfarm/prattle"
	"github.com/gojekfarm/prattle/config"
)

var (
	rpcPort  = flag.Int("rpc-port", 0, "RPC Port")
	httpPort = flag.Int("http-port", 0, "HTTP Port")
)

type entry struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func main() {
	flag.Parse()
	discovery := config.Discovery{
		TTL:                "10s",
		HealthEndpoint:     fmt.Sprintf("http://localhost:%d/_health", *httpPort),
		HealthPingInterval: "10s",
		Address:            "localhost",
		Name:               "Test",
		Port:               *rpcPort,
		ConsulURL:          "http://localhost:8500/",
	}
	consul, err := prattle.NewConsulClient("127.0.0.1:8500")
	if err != nil {
		log.Fatal(err)
	}
	prattle, err := prattle.NewPrattle(consul, *rpcPort, discovery)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/_health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(prattle.Members())
	})
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		val, _ := prattle.Get(r.FormValue("key"))
		json.NewEncoder(w).Encode(val)
	})
	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
		}
		var data entry
		json.Unmarshal(body, &data)
		prattle.Set(data.Key, data.Value)
		json.NewEncoder(w).Encode(data)
	})
	fmt.Printf("Listening on :%d\n", *httpPort)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), nil); err != nil {
		fmt.Println(err)
	}
}
