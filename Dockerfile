FROM docker.io/golang:alpine as builder

COPY . /src
WORKDIR /src
RUN apk add git make && make

FROM docker.io/alpine

LABEL maintainer="George <zhoreeq@users.noreply.github.com>"

COPY --from=builder /src/meshnamed /usr/bin/meshnamed

USER nobody

COPY docker-entrypoint.sh /usr/local/bin/
ENTRYPOINT ["docker-entrypoint.sh"]

EXPOSE 53535/udp
CMD ["meshnamed"]
