FROM golang:1.21.8-alpine3.19 as build
ENV GOPROXY=https://goproxy.cn
WORKDIR /openhydra
COPY . /openhydra
RUN apk add make bash which && make go-build

FROM rockylinux:8.8
COPY --from=build /openhydra/cmd/open-hydra-server/open-hydra-server /usr/bin/
EXPOSE 443
WORKDIR /usr/bin