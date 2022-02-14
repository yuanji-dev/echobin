FROM golang:alpine AS build
COPY . /go/src/app
WORKDIR /go/src/app
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o echobin

FROM scratch
EXPOSE 8080
ENV LISTEN_ADDR 0.0.0.0:8080
COPY --from=build /go/src/app/echobin /echobin
CMD ["/echobin"]
