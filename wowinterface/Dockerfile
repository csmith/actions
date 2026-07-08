FROM golang:1.26.5-alpine AS build

WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 go build -o /action ./wowinterface/cmd;

FROM alpine:3.23.4

RUN apk add --no-cache ca-certificates
COPY --from=build /action /action

ENTRYPOINT ["/action"]
