package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/divya2661/prattle/prattle"
)

var (
	members   = flag.String("members", "", "comma seperated list of members")
	rpc_port  = flag.Int("rpc_port", 0, "RPC port")
	http_port = flag.Int("http_port", 0, "http port")
	key       string
	value     string
)

type keyValue struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

func main() {
	flag.Parse()
	prattle, err := prattle.NewPrattle(*members, *rpc_port)
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/value", getValueHandler(prattle))
	http.HandleFunc("/", clusterHealthHandler(prattle))
	http.HandleFunc("/add_key", addKeyHandler(prattle))

	fmt.Printf("Listening on :%d\n", *http_port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *http_port), nil); err != nil {
		fmt.Println(err)
	}
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
