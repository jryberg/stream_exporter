FROM golang:1.27.4@sha256:205c7636d74e0ffbe3d007162df7d74f859f5a8c8322dfb9f907977686c126cf AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /stream_exporter .

FROM busybox:glibc@sha256:99813cb178ace2cd9c6c53ee3bddef5d2f5708aa438335111d7a9e3c5e1aaebe
COPY --from=build /stream_exporter /usr/bin/stream_exporter
USER 1000
ENTRYPOINT ["/usr/bin/stream_exporter"]
EXPOSE 9178
