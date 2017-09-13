FROM       golang
ADD         run.sh /bin/run.sh
ADD         raft-redis /bin/
RUN         chmod +x /bin/run.sh
VOLUME      /data
EXPOSE      6389 12379
ENTRYPOINT  ["/bin/run.sh"]
