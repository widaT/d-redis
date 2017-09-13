package server

import (
	"github.com/vmihailenco/msgpack"
	"log"
	"github.com/coreos/etcd/snap"
	"sync"
)

var _Storage * Storage
var _DBRwmu sync.RWMutex

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



func (s *Storage) Propose(opt *Opt) {
	b ,err := msgpack.Marshal(opt)
	//println(buf.String())
	if err != nil {
		log.Fatalf("msgpack Marshal err (%v)", err)
	}
	s.proposeC <- string(b)
}

func (s *Storage) readCommits(commitC <-chan *string, errorC <-chan error) {
	for data := range commitC {
		if data == nil {
			//println(" readCommits recive nil")
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
		_DBRwmu.Lock()

		/*switch dataKv.Method {
		case "set" :
			s.Redis.methodSet(dataKv.Args)
		case "del" :
			num := s.Redis.methodDel(dataKv.Args)
			if Conns.Exists(dataKv.Conn){
				respchan := Conns.Get(dataKv.Conn)
				respchan <- num
			}
		case "hset":
			num := s.Redis.methodHset(dataKv.Args)
			if Conns.Exists(dataKv.Conn){
				respchan := Conns.Get(dataKv.Conn)
				respchan <- num
			}

		case "rpush":
			num := s.Redis.methodRpush(dataKv.Args)
			if Conns.Exists(dataKv.Conn){
				respchan := Conns.Get(dataKv.Conn)
				respchan <- num
			}
		case "lpush":
			num := s.Redis.methodLpush(dataKv.Args)
			if Conns.Exists(dataKv.Conn){
				respchan := Conns.Get(dataKv.Conn)
				respchan <- num
			}
		case "lpop":
			byteArr := s.Redis.methodLpop(dataKv.Args)
			if Conns.Exists(dataKv.Conn){
				respchan := Conns.Get(dataKv.Conn)
				respchan <- byteArr
			}
		case "rpop":
			byteArr := s.Redis.methodRpop(dataKv.Args)
			if Conns.Exists(dataKv.Conn){
				respchan := Conns.Get(dataKv.Conn)
				respchan <- byteArr
			}
		case "sadd":
			num := s.Redis.methodSadd(dataKv.Args)
			if Conns.Exists(dataKv.Conn){
				respchan := Conns.Get(dataKv.Conn)
				respchan <- num
			}

		case "mset":
			s.Redis.methodMset(dataKv.Args)
		case "spop":
			s.Redis.methodSpop(dataKv.Args)
		case "incr":
			num,err := s.Redis.methodIncr(dataKv.Args)
			if Conns.Exists(dataKv.Conn){
				respchan := Conns.Get(dataKv.Conn)
				if err != nil {
					respchan <- err
				}else {
					respchan <- num
				}
			}
		default:
			//do nothing*//*
		}*/
		_DBRwmu.Unlock()
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
