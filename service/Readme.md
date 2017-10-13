### Build and Run

```
docker build -t sampleservice .
```

SAMPLE_SERVICE is the required prefix for env vars

```
docker run -dit  --name=sampleservice1 \
 -e SAMPLE_SERVICE_NAME=sampleservice \
 -e SAMPLE_SERVICE_ID=1 \
 -e CONSUL_HTTP_ADDR=node4:8500 \
 --net proxynet sampleservice
 ```

docker run -dit  --name=sampleservice3 \
 -e SAMPLE_SERVICE_NAME=sampleservice \
 -e SAMPLE_SERVICE_ID=3 \
 -e CONSUL_HTTP_ADDR=node4:8500 \
 --net proxynet sampleservice