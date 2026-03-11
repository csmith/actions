FROM golang:1.26.0-alpine AS build

WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod go build -o /action ./githubrelease/cmd;

FROM alpine:3.23.3

COPY --from=build /action /action

ENTRYPOINT ["/action"]
