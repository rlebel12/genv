FROM golang:latest as builder

WORKDIR /go/src/genv
ENV CGO_ENABLED=0
ENV GOOS=linux

COPY . .
RUN go build -o /genv ./example


FROM scratch as final

WORKDIR /go/src/genv

COPY --from=builder /genv .
