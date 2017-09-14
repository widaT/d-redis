package server

import (
	"log"
	"github.com/coreos/etcd/snap"
	"github.com/vmihailenco/msgpack"
	"strconv"
)

var _Storage * Storage


// a key-value store backed by raftd
type Storage struct {
	proposeC    chan<- string // channel for proposing updates
	Redis       *Memdb
	snapshotter *snap.Snapshotter
}

type kv struct {
	Method string
	Args [][]byte
	Conn string
}

func Run() {
	_Storage.snapshotter = <-snapshotterReady
	// replay log into key-value map
	_Storage.readCommits(commitC, errorC)
	// read commits from raftd into kvStore map until error
	go _Storage.readCommits(commitC, errorC)
}



func (s *Storage) Propose(kv *kv) {
	b ,err := msgpack.Marshal(kv)
	if err != nil {
		log.Fatalf("msgpack Marshal err (%v)", err)
	}
	s.proposeC <- string(b)
}

func (s *Storage) readCommits(commitC <-chan *string, errorC <-chan error) {
	for data := range commitC {
		if data == nil {
			snapshot, err := s.snapshotter.Load()
			if err == snap.ErrNoSnapshot {
				return
			}
			if err != nil && err != snap.ErrNoSnapshot {
				log.Panic(err)
			}
			log.Printf("loading snapshot at term %d and index %d", snapshot.Metadata.Term, snapshot.Metadata.Index)
			if err := s.recoverFromSnapshot(snapshot.Data); err != nil {
				log.Panic(err)
			}
			continue
		}else {
			if *data == "" {
				continue
			}
		}
		var dataKv kv
		err := msgpack.Unmarshal([]byte(*data),&dataKv)
		if err != nil {
			log.Fatalf("msgpack.Unmarshal err (%v)", err)
		}
		//log.Printf("do commit %s %s %s",dataKv.Method,dataKv.Args,dataKv.Conn)
		switch dataKv.Method {
		case "set" :
			s.Redis.Set(string(dataKv.Args[0]),dataKv.Args[1])
		case "mset":
			s.Redis.Mset(dataKv.Args...)
		case "del" :
			s.Redis.Del(dataKv.Conn,dataKv.Args...)
		case "hset":
			s.Redis.Hset(dataKv.Conn,string(dataKv.Args[0]),string(dataKv.Args[1]),dataKv.Args[2])
		case "rpush":
			s.Redis.Rpush(dataKv.Conn,dataKv.Args...)
		case "lpush":
			s.Redis.Lpush(dataKv.Conn,dataKv.Args...)
		case "lpop":
			s.Redis.Lpop(dataKv.Conn,string(dataKv.Args[0]))
		case "rpop":
			s.Redis.Rpop(dataKv.Conn,string(dataKv.Args[0]))
		case "sadd":
			s.Redis.Sadd(dataKv.Conn,string(dataKv.Args[0]),dataKv.Args[1:]...)
		case "spop":
			s.Redis.Sspop(dataKv.Conn,string(dataKv.Args[0]),dataKv.Args[1])
		case "incr":
			s.Redis.Incr(dataKv.Conn,string(dataKv.Args[0]))
		case "zadd":
			score ,_:= strconv.Atoi(string(dataKv.Args[1]))
			s.Redis.Zadd(dataKv.Conn,string(dataKv.Args[0]),score,string(dataKv.Args[2]))
		default:
			//do nothing*//*
		}
	}
	if err, ok := <-errorC; ok {
		log.Fatal(err)
	}
}

func (s *Storage) GetSnapshot()  ([]byte, error) {
	return s.Redis.getSnapshot()
}

func (s *Storage) recoverFromSnapshot(snapshot []byte) error {
	return s.Redis.recoverFromSnapshot(snapshot)
}
