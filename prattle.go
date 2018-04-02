package prattle

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/memberlist"

	"github.com/gojekfarm/prattle/config"
)

type Pair struct {
	Key   string
	Value interface{}
}

type Prattle struct {
	members    *memberlist.Memberlist
	broadcasts *memberlist.TransmitLimitedQueue
	database   *db
}

func NewPrattle(consul *Client, rpcPort int, discovery config.Discovery) (*Prattle, error) {
	member, err := consul.FetchHealthyNode(discovery.Name)
	if err != nil {
		return nil, err
	}
	fmt.Println("member: " + member)
	_, err = consul.Register(discovery)
	if err != nil {
		return nil, err
	}
	d := newDb()
	b := &memberlist.TransmitLimitedQueue{
		RetransmitMult: 3,
	}
	del := &delegate{
		getBroadcasts: func(overhead, limit int) [][]byte {
			return b.GetBroadcasts(overhead, limit)
		},
		notifyMsg: func(b []byte) {
			pair := &Pair{}
			json.Unmarshal(b, pair)
			d.Save(pair.Key, pair.Value)
		},
	}

	m, err := newMemberlist(rpcPort, member, del)
	if err != nil {
		return nil, err
	}
	b.NumNodes = func() int {
		return m.NumMembers()
	}
	return &Prattle{
		members:    m,
		broadcasts: b,
		database:   d,
	}, nil
}

func (p *Prattle) Get(k string) (interface{}, bool) {
	value, found := p.database.Get(k)
	return value, found
}

func (p *Prattle) Set(key string, value interface{}) error {
	p.database.Save(key, value)
	pair := &Pair{
		Key:   key,
		Value: value,
	}
	message, err := json.Marshal(pair)
	if err != nil {
		return err
	}
	go func() {
		p.broadcasts.QueueBroadcast(&broadcast{
			msg:    message,
			notify: nil,
		})
	}()
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
}

func (p *Prattle) JoinCluster(siblingAddr string) error {
	_, err := p.members.Join([]string{siblingAddr})
	if err != nil {
		log.Fatal("Could not join the cluster with sibling", siblingAddr)
		return err
	}
	return nil
}
