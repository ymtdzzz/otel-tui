#
# Build
#
FROM golang:1.24 AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
WORKDIR /build

COPY . .

RUN go build -o otel-tui ./...

#
# Deploy
#
FROM gcr.io/distroless/static-debian12:latest

WORKDIR /

COPY --from=builder /build/otel-tui /otel-tui

USER nonroot

CMD [ "/otel-tui" ]
