#
# Build
#
FROM golang:1.22 AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
WORKDIR /build

COPY . .

RUN go build -o main ./...

#
# Deploy
#
FROM gcr.io/distroless/static-debian12:latest

WORKDIR /

COPY --from=builder /build/main /main

USER nonroot

CMD [ "/main" ]
