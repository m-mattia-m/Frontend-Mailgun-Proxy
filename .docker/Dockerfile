FROM docker.io/alpine:3.18.2

ENV GIN_MODE="release"

WORKDIR /app

COPY .. /app/mg-proxy

EXPOSE 8080

ENTRYPOINT [ "/app/mg-proxy" ]