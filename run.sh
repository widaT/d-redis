#!/bin/sh
# Check for $LISTEN_PORT
if [ -z ${LISTEN_PORT+x} ]; then
        LISTEN_PORT="6389"
        echo "Using default LISTEN_URLS ($LISTEN_PORT)"
else
        echo "Detected new LISTEN_URLS value of $LISTEN_PORT"
fi

# Check for $CLUSTER
if [ -z ${CLUSTER+x} ]; then
        CLUSTER="http://0.0.0.0:12379"
        echo "Using default ADVERTISE_URLS ($CLUSTER)"
else
        echo "Detected new ADVERTISE_URLS value of $CLUSTER"
fi

REDIS_CMD="/bin/raft-redis -data-dir=/data --cluster=${CLUSTER} --port=${LISTEN_PORT} $*"
echo -e "Running '$REDIS_CMD'\nBEGIN RAFT_REDIS OUTPUT\n"

exec $REDIS_CMD