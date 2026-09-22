FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /stream_exporter .

FROM busybox:glibc
COPY --from=build /stream_exporter /usr/bin/stream_exporter
USER 1000
ENTRYPOINT ["/usr/bin/stream_exporter"]
EXPOSE 9178
