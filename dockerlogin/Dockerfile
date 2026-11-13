FROM golang:1.27.4-alpine AS build

WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod go build -o /action ./dockerlogin/cmd;

FROM quay.io/buildah/stable:v1.43.4

COPY --from=build /action /action

ENTRYPOINT ["/action"]
