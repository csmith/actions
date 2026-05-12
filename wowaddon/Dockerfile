FROM golang:1.26.3-alpine AS build

WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 go build -o /action ./wowaddon/cmd;

FROM scratch

COPY --from=build /action /action

ENTRYPOINT ["/action"]
