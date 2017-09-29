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

# benchmark

![benchmark](https://raw.githubusercontent.com/widaT/d-redis/master/doc/benchmark.png)

# Thanks

[etcd-raft](https://github.com/coreos/etcd/tree/master/raft)