# d-redis

基于raft协议试下的简单redis 服务器

# 开始使用
### 单一节点跑d-redis

第一步：

```sh
d-redis --id 1 --cluster http://127.0.0.1:12379 --port 6389
```

第二步：

```
redis-cli -p 6389 set key hello
```

最后：

```
redis-cli -p 6389 get key
```

### 本地伪分布式集群

首先安装 [goreman](https://github.com/mattn/goreman), 然后执行如下shell，会启动3个d-redis 实例

```sh
goreman start
```

# benchmark

![benchmark](https://raw.githubusercontent.com/widaT/d-redis/master/doc/benchmark.png)

# 感谢

[etcd-raft](https://github.com/coreos/etcd/tree/master/raft)

[go-redis-server](https://github.com/docker/go-redis-server)