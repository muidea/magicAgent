FROM golang:1.19 AS building

COPY . /building
WORKDIR /building

RUN make build

FROM alpine:3.13.4

ARG AppName=magieAgent
LABEL Author="rangh"
LABEL Application=$AppName

RUN mkdir -p /var/app
RUN mkdir -p /var/app/config

WORKDIR /var/app

COPY --from=building /building/bin/$AppName /var/app/$AppName
COPY --from=building /building/docker/bootstrap.sh /var/app/bootstrap.sh

RUN chmod +x /var/app/bootstrap.sh

ENTRYPOINT [ "/var/app/bootstrap.sh" ]

EXPOSE 8080
