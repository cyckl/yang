# syntax=docker/dockerfile:1

FROM golang:1.16-buster AS build

WORKDIR /container

# Pull dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Build binary
COPY *.go ./
RUN go build -o /yang

# Build stripped container thing
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /yang /yang

CMD ["/yang"]