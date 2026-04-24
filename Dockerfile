FROM golang:1.26.0-bookworm AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/netops-lab ./cmd/netops-lab

FROM gcr.io/distroless/static-debian12:nonroot
USER nonroot:nonroot
COPY --from=build /out/netops-lab /netops-lab
EXPOSE 8080 9090 9091/udp
ENTRYPOINT ["/netops-lab"]

