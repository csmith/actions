FROM golang:1.25.6-alpine AS build

WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod go build -o /action ./dockerpush/cmd;

FROM quay.io/buildah/stable:v1.42.2

COPY --from=build /action /action

ENTRYPOINT ["/action"]
