FROM golang:1.27.2-alpine AS build

WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod go build -o /action ./dockerbuild/cmd;

FROM quay.io/buildah/stable:v1.43.1

RUN sed -i 's/short-name-mode = "enforcing"/short-name-mode = "disabled"/g' /etc/containers/registries.conf

COPY --from=build /action /action

ENTRYPOINT ["/action"]
