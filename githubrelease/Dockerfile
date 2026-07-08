FROM golang:1.26.5-alpine AS build

WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod go build -o /action ./githubrelease/cmd;

FROM alpine:3.24.0

COPY --from=build /action /action

ENTRYPOINT ["/action"]
