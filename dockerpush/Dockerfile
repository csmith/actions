FROM golang:1.27.2-alpine AS build

WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod go build -o /action ./dockerpush/cmd;

FROM quay.io/skopeo/stable:v1.22.0

COPY --from=build /action /action

ENTRYPOINT ["/action"]
