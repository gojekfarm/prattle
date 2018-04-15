package prattle

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/memberlist"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/gojekfarm/prattle/config"
)

type BroadcastMessage struct {
	Key   string
	Value interface{}
}

type Prattle struct {
	members          *memberlist.Memberlist
	broadcasts       *memberlist.TransmitLimitedQueue
	database         *db
	statsDClient     statsd.Statter
	broadcastChannel chan BroadcastMessage
}

func NewPrattle(consul *Client, rpcPort int, discovery config.Discovery) (*Prattle, error) {
	var serviceID string
	statsDClient, _ := statsd.NewBufferedClient("127.0.0.1:8127", "", 5*time.Millisecond, 10)
	broadcastChannel := make(chan BroadcastMessage, 10000)

	member, err := consul.FetchHealthyNode(discovery.Name)
	if err != nil {
		return nil, err
	}
	fmt.Println("member: " + member)
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
			if _, ok := d.Get(pair.Key); ok == false {
				fmt.Println("saving")
				statsDClient.Inc("bla", int64(1), float32(1))
				d.Save(pair.Key, pair.Value)
			}
		},
	}

	m, err := newMemberlist(rpcPort, member, del)
	if err != nil {
		return nil, err
	}
	transmitLimitedQueue.NumNodes = func() int { return m.NumMembers() }
	ticker := time.NewTicker(time.Second)
	go startPingWorker(ticker, serviceID, consul);
	startBroadcastWorker(broadcastChannel, transmitLimitedQueue)
	return &Prattle{
		members:          m,
		broadcasts:       transmitLimitedQueue,
		database:         d,
		statsDClient:     statsDClient,
		broadcastChannel: broadcastChannel,
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

func (p *Prattle) Set(key string, value interface{}) error {
	p.database.Save(key, value)
	p.statsDClient.Inc("source", int64(1), float32(1))
	p.broadcastChannel <- BroadcastMessage{
		Key:   key,
		Value: value,
	}
	return nil
}

func (p *Prattle) Members() []string {
	var a []string
	for _, m := range p.members.Members() {
		a = append(a, m.Addr.String())
	}
	return a
}

func (p *Prattle) Shutdown() {
	p.members.Shutdown()
	close(p.broadcastChannel)
}

func (p *Prattle) JoinCluster(siblingAddr string) error {
	_, err := p.members.Join([]string{siblingAddr})
	if err != nil {
		log.Fatal("Could not join the cluster with sibling", siblingAddr)
		return err
	}
	return nil
}
