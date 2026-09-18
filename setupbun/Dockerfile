FROM golang:1.26.6-alpine AS build

WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod CGO_ENABLED=0 go build -o /action ./setupbun/cmd;

FROM oven/bun:1.4.2

# Run as root like the other actions so the workspace can be written to.
USER root

COPY --from=build /action /action

ENTRYPOINT ["/action"]
