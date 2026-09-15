FROM golang:1.26.6-alpine AS build

WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod go build -o /action ./dockerlogin/cmd;

FROM quay.io/buildah/stable:v1.43.1

COPY --from=build /action /action

ENTRYPOINT ["/action"]
