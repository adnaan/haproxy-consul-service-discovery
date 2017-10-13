### Consul 

#### Server

```
docker run -d --name node1 -h node1 consul agent -server -bootstrap-expect 3
JOIN_IP="$(docker inspect -f '{{.NetworkSettings.IPAddress}}' node1)"
docker run -d --name node2 -h node2 consul agent -server -join $JOIN_IP
docker run -d --name node3 -h node3 consul agent -server -join $JOIN_IP
```

#### Client
```
docker run -d -p 8500:8500 -p 8600:8600/udp --name node4 -h node4 consul agent -join $JOIN_IP  -ui -client=0.0.0.0 -bind='{{ GetPrivateIP }}'
```


### Network

```
docker network create proxynet
docker network connect proxynet node1
docker network connect proxynet node2
docker network connect proxynet node3
docker network connect proxynet node4
```

### APP

Change service id to simulate a new node

```

cd service/cmd
docker bulild -t sampleservice .

docker run -dit  --name=sampleservice1 \
 -e SAMPLE_SERVICE_NAME=sampleservice \
 -e SAMPLE_SERVICE_ID=1 \
 -e CONSUL_HTTP_ADDR=node4:8500 \
 --net proxynet sampleservice
 
```

Service registers itself to consul and deregisters on stop.

Make sure to stop the service by `docker stop` and NOT `docker rm -f`

```
docker stop sampleservice1
```

#### Haproxy

```

cd haproxy 
docker build -t haproxy .

docker rm -f proxy || true && \
 docker run -dit -v $(PWD):/var/log --name proxy \
 --net proxynet -p 80:80 haproxy && \
 docker logs proxy
```


### Register App Backend
```
curl -X PUT -d '{"ID": "myapp3", "Name": "sampleapp", "Address": "172.19.0.9", "Port": 3344}' http://localhost:8500/v1/agent/service/register
```

### Deregister App Backend

```
curl --request PUT http://localhost:8500/v1/agent/service/deregister/myapp3
```

### dig

```
dig -t srv @localhost -p 8600 sampleapp.service.consul

docker run -it --rm --net=proxynet sequenceiq/busybox dig -t srv @172.19.0.7 -p 8600 sampleapp.service.consul
```


### GET IP Addrress

```
docker inspect --format '{{ .NetworkSettings.Networks.proxynet.IPAddress }}{{ .Name }}' <containerid>
```

