# mcs in a container: docker build -t mcs .
# The default command is `mcs mqtt`; its settings come from MCS_MQTT_* variables
# or /state/.config/mcs/config.toml. /state is the home directory: put the
# signed-in session there (state/.cache/mcs/token.json), see examples/mqtt.
FROM --platform=$BUILDPLATFORM golang:1.27 AS build
ARG TARGETOS TARGETARCH VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w -X main.Version=$VERSION" -o /mcs ./cmd/mcs

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /mcs /mcs
ENV HOME=/state
WORKDIR /state
ENTRYPOINT ["/mcs"]
CMD ["mqtt"]
