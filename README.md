# d-redis

[中文简介](https://github.com/widaT/d-redis/tree/master/doc)

simple redis server base etcd raft lib

# Getting Started
### Running single node d-redis

First start a single-member cluster of raftexample:

```sh
d-redis --id 1 --cluster http://127.0.0.1:12379 --port 6389
```

Next, store a value ("hello") to a key ("key") use redis-cli:

```
redis-cli -p 6389 set key hello
```

Finally, retrieve the stored key:

```
redis-cli -p 6389 get key
```

### Running a local cluster

First install [goreman](https://github.com/mattn/goreman), which manages Procfile-based applications.

The [Procfile script](./Procfile) will set up a local example cluster. Start it with:

```sh
goreman start
```

This will bring up three d-redis instances.

### Dynamic cluster reconfiguration

Nodes can be added to or removed from a running cluster using redis command-line client.

For example, suppose we have a 3-node cluster that was started with the commands:
```sh
raftexample --id 1 --cluster http://127.0.0.1:12379,http://127.0.0.1:22379,http://127.0.0.1:32379 --port 6389
raftexample --id 2 --cluster http://127.0.0.1:12379,http://127.0.0.1:22379,http://127.0.0.1:32379 --port 6399
raftexample --id 3 --cluster http://127.0.0.1:12379,http://127.0.0.1:22379,http://127.0.0.1:32379 --port 6169
```

A fourth node with ID 4 can be added by useing a addnode request :
```sh
redis-cli -p 6389 addnode 4 http://127.0.0.1:42379
```

Then the new node can be started as the others were, using the --join option:
```sh
d-redis --id 4 --cluster http://127.0.0.1:12379,http://127.0.0.1:22379,http://127.0.0.1:32379,http://127.0.0.1:42379 --port 6059 --join
```

The new node should join the cluster and be able to service redis client requests.

We can remove a node using a removenode request:
```sh
redis-cli -p 6389 removenode 4 http://127.0.0.1:42379
```

Node 3 should shut itself down once the cluster has processed this request.

# benchmark

![benchmark](https://raw.githubusercontent.com/widaT/d-redis/master/doc/benchmark.png)

# Thanks

[etcd-raft](https://github.com/coreos/etcd/tree/master/raft)