FROM golang:1.25.6-alpine AS build

WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod go build -o /action ./imagetags/cmd;

FROM scratch

COPY --from=build /action /action

ENTRYPOINT ["/action"]
