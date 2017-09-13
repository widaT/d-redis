package server

import (
	"flag"
	"strings"
	"fmt"
	"log"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/snap"
)
var (
	Conns = NewConnMap()
	proposeC = make(chan string)
	confChangeC = make(chan raftpb.ConfChange)
	commitC = make(chan *string)
	errorC = make(chan error)
	snapshotterReady = make(chan *snap.Snapshotter, 1)
	ID  int
)
func Main()  {
	cluster := flag.String("cluster", "http://127.0.0.1:12379", "comma separated cluster peers")
	id := flag.Int("id", 1, "node ID")
	ID = *id
	kvport := flag.Int("port", 6380, "key-value server port")
	s := flag.Uint64("s",1000000,"snapshot count")
	join := flag.Bool("join", false, "join an existing cluster")
	dataDir := flag.String("data-dir","data/","store databases")
	flag.Parse()
	defer close(proposeC)
	defer close(confChangeC)
	_Storage = &Storage{proposeC: proposeC, Redis: NewMemdb()}
	NewRaftNode(*id, strings.Split(*cluster, ","), strings.TrimRight(*dataDir,"/"),*join)
	go func() {
		c := DefaultConfig().SnapCount(*s).Laddr(fmt.Sprintf(":%d",*kvport))
		err := ListenAndServe(c,
			func(conn Conn) bool {
				return true
			},
			func(conn Conn, err error) {
			},
		)
		if err != nil {
			log.Fatal(err)
		}
	}()
	Run()
	if err, ok := <-errorC; ok {
		log.Fatalf("raft-redis: error loading wal (%v)", err)
	}
}
