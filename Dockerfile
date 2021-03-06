FROM golang:alpine as app-builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /go/src/app
COPY . .
# Static build required so that we can safely copy the binary over.
# `-tags timetzdata` embeds zone info from the "time/tzdata" package.
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go install -ldflags '-extldflags "-static"' -tags timetzdata

FROM --platform=$BUILDPLATFORM scratch
# the test program:
COPY --from=app-builder /go/bin/speedtest2mqtt /speedtest2mqtt
# the tls certificates:
# NB: this pulls directly from the upstream image, which already has ca-certificates:
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/speedtest2mqtt"]