FROM golang:1.26.2-bookworm AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pulsemesh ./cmd/pulsemesh

FROM gcr.io/distroless/static-debian12:nonroot
USER nonroot:nonroot
COPY --from=build /out/pulsemesh /pulsemesh
EXPOSE 8080 9090 9091/udp
ENTRYPOINT ["/pulsemesh"]
