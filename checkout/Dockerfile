FROM golang:1.25.6-alpine AS build

WORKDIR /go/src/app
COPY . .

RUN find .
RUN --mount=type=cache,target=/go/pkg/mod go build -o /action ./checkout/cmd;

FROM alpine:3.23.2

RUN apk add --no-cache git
COPY --from=build /action /action

ENTRYPOINT ["/action"]
