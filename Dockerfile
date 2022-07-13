ARG ARCH

FROM ${ARCH}/alpine:latest

COPY bin/intercept-dns /usr/local/bin/intercept-dns

ENV REMOTE_DNS=1.1.1.1
ENV REMOTE_PORT=53

EXPOSE 53

CMD [ "sh", "-c", "/usr/local/bin/intercept-dns -remote-dns-ip $REMOTE_DNS -remote-dns-port $REMOTE_PORT" ]