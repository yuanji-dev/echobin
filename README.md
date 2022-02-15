# echobin

Yet another Golang port of [httpbin](https://httpbin.org/)(a HTTP request & response testing service), powered by [echo framework](https://echo.labstack.com/).

[![Docker Image Size (latest)](https://img.shields.io/docker/image-size/gimo/echobin/latest?color=green&label=docker%20image&style=flat-square)](https://hub.docker.com/r/gimo/echobin)
[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://github.com/masakichi/echobin/blob/main/LICENSE)
[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/masakichi/echobin/Run%20Tests?style=flat-square)](https://github.com/masakichi/echobin/actions)

## Online Instance

- [https://echobin.gimo.me](https://echobin.gimo.me)

## Run Locally

- Run in docker

```bash
docker run -p 8080:8080 gimo/echobin
```

- Or if you have Go 1.16+ installed

```bash
git clone https://github.com/masakichi/echobin.git
cd echobin
go run .
```
