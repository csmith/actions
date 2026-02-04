FROM golang:1.25.7-alpine AS build

WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod go build -o /action ./dockerlogin/cmd;

FROM quay.io/buildah/stable:v1.42.2

COPY --from=build /action /action

ENTRYPOINT ["/action"]
