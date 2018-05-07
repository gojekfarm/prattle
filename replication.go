package prattle

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/hashicorp/memberlist"
)

type delegate struct {
	StatsDClient  statsd.Statter
	notifyMsg     func([]byte)
	getBroadcasts func(int, int) [][]byte
}

func (d *delegate) NodeMeta(limit int) []byte {
	return []byte{}
}

func (d *delegate) NotifyMsg(b []byte) {
	d.notifyMsg(b)
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return d.getBroadcasts(overhead, limit)
}

func (d *delegate) LocalState(join bool) []byte {
	return []byte{}
}

func (d *delegate) MergeRemoteState(buf []byte, join bool) {
}

type broadcast struct {
	msg    []byte
	notify chan<- struct{}
}

func (b *broadcast) Invalidates(other memberlist.Broadcast) bool {
	return false
}

func (b *broadcast) Message() []byte {
	return b.msg
}

func (b *broadcast) Finished() {
	if b.notify != nil {
		close(b.notify)
	}
}

func newMemberlist(rpc_port int, members string, d *delegate) (*memberlist.Memberlist, error) {
	hostname, _ := os.Hostname()
	c := memberlist.DefaultLANConfig()
	c.BindPort = rpc_port
	c.Name = fmt.Sprintf("%s-%d", hostname, rand.Intn(100))
	c.PushPullInterval = time.Second * 5 // to make sync demonstrable
	c.ProbeInterval = time.Second * 1    // to make failure demonstrable
	c.Delegate = d

	m, err := memberlist.Create(c)
	if err != nil {
		return nil, err
	}

	if len(members) > 0 {
		parts := strings.Split(members, ",")
		_, err := m.Join(parts)
		if err != nil {
			return m, err
		}
	}

	node := m.LocalNode()
	log.Println("Local member %s:%d\n", node.Addr, node.Port)
	return m, nil
}
