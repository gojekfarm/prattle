package prattle

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"os"
	"time"

	"github.com/hashicorp/memberlist"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/gojekfarm/prattle/config"
)

const postfixSource = "_source"
const postfixDest = "_dest"

type BroadcastMessage struct {
	Key   string
	Value interface{}
}

type Prattle struct {
	memberlist       *memberlist.Memberlist
	broadcasts       *memberlist.TransmitLimitedQueue
	database         *db
	statsDClient     statsd.Statter
	broadcastChannel chan BroadcastMessage
	ticker           *time.Ticker
	hostname         string
	ip               string
}

func NewPrattle(consul *Client, rpcPort int, discovery config.Discovery) (*Prattle, error) {
	hostname, _ := os.Hostname()
	ips, e := net.LookupIP(hostname)
	if e != nil {
		return nil, errors.New("ip resolution failed")
	}
	var serviceID string
	statsDClient, _ := statsd.NewBufferedClient("127.0.0.1:8125", "", 5*time.Millisecond, 10)
	broadcastChannel := make(chan BroadcastMessage, 10000)

	member, err := consul.FetchHealthyNode(discovery.Name)
	if err != nil {
		return nil, err
	}
	log.Println("member: " + member)
	serviceID, err = consul.Register(discovery)
	if err != nil {
		return nil, err
	}
	d := newDb()
	transmitLimitedQueue := &memberlist.TransmitLimitedQueue{
		RetransmitMult: 3,
	}

	del := &delegate{
		getBroadcasts: func(overhead, limit int) [][]byte {
			return transmitLimitedQueue.GetBroadcasts(overhead, limit)
		},
		notifyMsg: func(b []byte) {
			pair := &BroadcastMessage{}
			json.Unmarshal(b, pair)
			statsDClient.Inc(hostname+postfixDest, int64(1), float32(1))
			d.Save(pair.Key, pair.Value)
		},
	}

	m, err := newMemberlist(rpcPort, member, del)
	if err != nil {
		return nil, err
	}
	transmitLimitedQueue.NumNodes = func() int { return m.NumMembers() }
	ticker := time.NewTicker(time.Second)
	go startPingWorker(ticker, serviceID, consul)
	startBroadcastWorker(broadcastChannel, transmitLimitedQueue)

	return &Prattle{
		memberlist:       m,
		broadcasts:       transmitLimitedQueue,
		database:         d,
		statsDClient:     statsDClient,
		broadcastChannel: broadcastChannel,
		ticker:           ticker,
		hostname:         hostname,
		ip:               ips[0].String(),
	}, nil
}

func startPingWorker(ticker *time.Ticker, serviceID string, consul *Client) {
	checkID := "service:" + serviceID
	for range ticker.C {
		consul.Ping(checkID)
	}
}

func startBroadcastWorker(broadcastChannel chan BroadcastMessage, transmitLimitedQueue *memberlist.TransmitLimitedQueue) {
	go func(transmitLimitedQueue *memberlist.TransmitLimitedQueue) {
		for {
			broadcastMessage := <-broadcastChannel
			message, err := json.Marshal(broadcastMessage)
			if err == nil {
				transmitLimitedQueue.QueueBroadcast(&broadcast{
					msg:    message,
					notify: nil,
				})
			}
		}
	}(transmitLimitedQueue)
}

func (p *Prattle) Get(k string) (interface{}, bool) {
	value, found := p.database.Get(k)
	return value, found
}

func (p *Prattle) SetViaGossip(key string, value interface{}) error {
	p.database.Save(key, value)
	p.statsDClient.Inc(p.hostname+postfixSource, int64(1), float32(1))
	p.broadcastChannel <- BroadcastMessage{
		Key:   key,
		Value: value,
	}
	return nil
}

func (p *Prattle) SetViaUDP(key string, value interface{}) error {
	p.database.Save(key, value)

	for _, node := range p.memberlist.Members() {
		if p.hasDifferentIpAs(node.Addr.String()) {
			broadcastMessage := BroadcastMessage{Key: key, Value: value}
			serialisedBroadcastMessage, _ := json.Marshal(broadcastMessage)
			p.statsDClient.Inc(p.hostname+postfixSource, int64(1), float32(1))
			p.memberlist.SendBestEffort(node, serialisedBroadcastMessage)
		}
	}
	return nil
}

func (p *Prattle) Members() []string {
	var a []string
	for _, m := range p.memberlist.Members() {
		a = append(a, m.Addr.String())
	}
	return a
}

func (p *Prattle) Shutdown() {
	p.ticker.Stop()
	close(p.broadcastChannel)
	p.memberlist.Shutdown()
}

func (p *Prattle) JoinCluster(siblingAddr string) error {
	_, err := p.memberlist.Join([]string{siblingAddr})
	if err != nil {
		log.Fatal("Could not join the cluster with sibling", siblingAddr)
		return err
	}
	return nil
}
func (prattle *Prattle) hasDifferentIpAs(ipAddress string) bool {
	return prattle.ip != ipAddress;
}
