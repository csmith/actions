FROM golang:1.26.1-alpine AS build

WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod go build -o /action ./checkout/cmd;

FROM alpine:3.23.3

RUN apk add --no-cache git
COPY --from=build /action /action

ENTRYPOINT ["/action"]
