FROM fingershock/haproxy-base:1.8-dev2

# # defaults for the variables used in haproxy.cfg
RUN apk update && \
    apk add bind-tools curl && \
    rm -rf /var/cache/apk/*

ENV DNS_TCP_ADDR 172.19.0.7
ENV DNS_TCP_PORT 8600
COPY ./haproxy.cfg /usr/local/etc/haproxy/haproxy.cfg
CMD [ "haproxy","-f","/usr/local/etc/haproxy/haproxy.cfg"]