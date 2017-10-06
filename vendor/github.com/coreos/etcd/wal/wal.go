// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wal

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"
	"sync"
	"time"

	"github.com/coreos/etcd/pkg/pbutil"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/coreos/etcd/wal/walpb"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/coreos/pkg/capnslog"
	"github.com/coreos/etcd/pkg/fileutil"
	"strings"
	"strconv"
)

const (
	metadataType int64 = iota + 1
	entryType
	stateType
	crcType
	snapshotType

	// warnSyncDuration is the amount of time allotted to an fsync before
	// logging a warning
	warnSyncDuration = time.Second
	LOGKEYPREFIX = "LOG_%016x"
)

var (
	// SegmentSizeBytes is the preallocated size of each wal segment file.
	// The actual size might be larger than this. In general, the default
	// value should be used, but this is defined as an exported variable
	// so that tests can set a different segment size.
	SegmentSizeBytes int64 = 64 * 1000 * 1000 // 64MB

	plog = capnslog.NewPackageLogger("github.com/coreos/etcd", "wal")

	ErrMetadataConflict = errors.New("wal: conflicting metadata found")
	ErrFileNotFound     = errors.New("wal: file not found")
	ErrCRCMismatch      = errors.New("wal: crc mismatch")
	ErrSnapshotMismatch = errors.New("wal: snapshot mismatch")
	ErrSnapshotNotFound = errors.New("wal: snapshot not found")
	crcTable            = crc32.MakeTable(crc32.Castagnoli)
)

// WAL is a logical representation of the stable storage.
// WAL is either in read mode or append mode but not both.
// A newly created WAL is in append mode, and ready for appending records.
// A just opened WAL is in read mode, and ready for reading records.
// The WAL will be ready for appending after reading out all the previous records.
type WAL struct {
	dir string // the living directory of the underlay files
	// dirFile is a fd for the wal directory for syncing on Rename
	metadata []byte           // metadata recorded at the head of each WAL
	state    raftpb.HardState // hardstate recorded at the head of WAL
	start     walpb.Snapshot // snapshot to start reading

	mu      sync.Mutex
	enti    uint64   // index of the last entry saved to the wal


	Db  * leveldb.DB
}


var record_index uint64 = 0
func (w *WAL)save(log []byte)  error {
	record_index ++
	return w.Db.Put([]byte(fmt.Sprintf(LOGKEYPREFIX,record_index)), log, nil)
}
// Create creates a WAL ready for appending records. The given metadata is
// recorded at the head of each WAL file, and can be retrieved with ReadAll.
func Create(dirpath string, metadata []byte) (*WAL, error) {

	w := &WAL{
		dir:      dirpath,
		metadata: metadata,
	}
	w.Db,_ = leveldb.OpenFile(w.dir,nil)
	if err := w.SaveSnapshot(walpb.Snapshot{}); err != nil {
		return nil, err
	}

	return w, nil
}


func Open(dirpath string, snap walpb.Snapshot) (*WAL, error) {
	w := &WAL{
		dir:      dirpath,
	}
	w.Db,_ = leveldb.OpenFile(w.dir,nil)
	return w, nil
}

func Exist(dirpath string) bool {
	names, err := fileutil.ReadDir(dirpath)
	if err != nil {
		return false
	}
	return len(names) != 0
}



func (w *WAL) ReadAll(snap *raftpb.Snapshot) (metadata []byte, state raftpb.HardState, ents []raftpb.Entry, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	iter := w.Db.NewIterator(nil, nil)
	if snap != nil {
		w.start = walpb.Snapshot{Index:snap.Metadata.Index,Term:snap.Metadata.Term}
		iter.Seek([]byte(fmt.Sprintf(LOGKEYPREFIX,snap.Metadata.Index)))
	}

	var match bool
	for iter.Next() {
		rec := walpb.Record{}
		rec.Unmarshal(iter.Value())

		record_index,_ = strconv.ParseUint( strings.TrimLeft(string(iter.Key()),"LOG_"),10,64)
		switch rec.Type {
		case entryType:
			e := mustUnmarshalEntry(rec.Data)
			if e.Index > w.start.Index {
				ents = append(ents[:e.Index-w.start.Index-1], e)
			}
			w.enti = e.Index
		case stateType:
			state = mustUnmarshalState(rec.Data)
		case metadataType:
			if metadata != nil && !bytes.Equal(metadata, rec.Data) {
				state.Reset()
				return nil, state, nil, ErrMetadataConflict
			}
			metadata = rec.Data
		case crcType:
			/*crc := decoder.crc.Sum32()
			// current crc of decoder must match the crc of the record.
			// do no need to match 0 crc, since the decoder is a new one at this case.
			if crc != 0 && rec.Validate(crc) != nil {
				state.Reset()
				return nil, state, nil, ErrCRCMismatch
			}
			decoder.updateCRC(rec.Crc)*/
		case snapshotType:
			var snap walpb.Snapshot
			pbutil.MustUnmarshal(&snap, rec.Data)
			if snap.Index == w.start.Index {
				if snap.Term != w.start.Term {
					state.Reset()
					return nil, state, nil, ErrSnapshotMismatch
				}
				match = true
			}
		default:
			state.Reset()
			return nil, state, nil, fmt.Errorf("unexpected block type %d", rec.Type)
		}
	}


	err = nil
	if !match {
		err = ErrSnapshotNotFound
	}

	w.start = walpb.Snapshot{}

	w.metadata = metadata

	return metadata, state, ents, err
}


func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.Db.Close()
}

func (w *WAL) saveEntry(e *raftpb.Entry) error {
	// TODO: add MustMarshalTo to reduce one allocation.
	b := pbutil.MustMarshal(e)
	rec := &walpb.Record{Type: entryType, Data: b}
	data,_:= rec.Marshal()
	if err := w.save(data); err != nil {
		return err
	}
	w.enti = e.Index
	return nil
}

func (w *WAL) saveState(s *raftpb.HardState) error {
	if raft.IsEmptyHardState(*s) {
		return nil
	}
	w.state = *s
	b := pbutil.MustMarshal(s)
	rec := &walpb.Record{Type: stateType, Data: b}
	data,_:= rec.Marshal()
	return  w.save(data)
}

func (w *WAL) Save(st raftpb.HardState, ents []raftpb.Entry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// short cut, do not call sync
	if raft.IsEmptyHardState(st) && len(ents) == 0 {
		return nil
	}

	// TODO(xiangli): no more reference operator
	for i := range ents {
		if err := w.saveEntry(&ents[i]); err != nil {
			return err
		}
	}
	if err := w.saveState(&st); err != nil {
		return err
	}

	return nil
}

func (w *WAL) SaveRecord(index uint64,rec *walpb.Record) error {
	data,err:= rec.Marshal()
	if err != nil {
		return err
	}
	if err := w.save(data); err != nil {
		return err
	}
	w.enti = index
	return nil
}

func (w *WAL) SaveSnapshot(e walpb.Snapshot) error {
	b := pbutil.MustMarshal(&e)
	w.mu.Lock()
	defer w.mu.Unlock()
	rec := &walpb.Record{Type: snapshotType, Data: b}
	w.SaveRecord(e.Index,rec)
	// update enti only when snapshot is ahead of last index
	if w.enti < e.Index {
		w.enti = e.Index
	}
	return nil
}


func mustUnmarshalEntry(d []byte) raftpb.Entry {
	var e raftpb.Entry
	pbutil.MustUnmarshal(&e, d)
	return e
}

func mustUnmarshalState(d []byte) raftpb.HardState {
	var s raftpb.HardState
	pbutil.MustUnmarshal(&s, d)
	return s
}