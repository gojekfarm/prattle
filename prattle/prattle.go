package prattle

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/memberlist"
	"log"
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

func NewPrattle(members string, port int) (*Prattle, error) {
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

	m, err := newMemberlist(port, members, del)
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

func (p *Prattle) Set(key string, value interface{}) {
	p.database.Save(key, value)
	pair := &Pair{Key: key, Value: value}
	message, err := json.Marshal(pair)
	if err != nil {
		fmt.Println(err)
	}

	p.broadcasts.QueueBroadcast(&broadcast{
		msg:    message,
		notify: nil,
	})
}

func (p *Prattle) Members() []string {
	a := []string{}
	for _, m := range p.members.Members() {
		a = append(a, string(m.Addr))
	}
	return a
}

func (p *Prattle) Shutdown() {
	p.members.Shutdown()
}

//TODO: Add logic to recover node failures
//1. Whenever new node join the cluster, it uses sibling address, as soon as it is joined
// it will create a metadata file in system using which it can always join the cluster.
//TODO: check is memberlist already does that
func (p *Prattle) JoinCluster(siblingAddr string) error {
	_, err := p.members.Join([]string{siblingAddr})
	if (err != nil) {
		log.Fatal("Could not join the cluster with sibling", siblingAddr)
		return err
	}
	return nil
}
