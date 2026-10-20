FROM golang:1.27.3 AS build

WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod go build -o /action ./setupgo/cmd;

FROM golang:1.27.3

COPY --from=build /action /action

ENTRYPOINT ["/action"]
