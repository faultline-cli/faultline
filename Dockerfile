# syntax=docker/dockerfile:1.7

FROM golang:1.24-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/faultline ./cmd

FROM alpine:3.21
RUN addgroup -S faultline && adduser -S -G faultline faultline
WORKDIR /workspace
ENV HOME=/home/faultline

COPY --from=build /out/faultline /usr/local/bin/faultline
COPY playbooks /playbooks

RUN mkdir -p /home/faultline/.faultline/packs && chown -R faultline:faultline /home/faultline /workspace

USER faultline
ENTRYPOINT ["faultline"]
