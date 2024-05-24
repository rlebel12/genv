FROM public.ecr.aws/docker/library/golang:1.22.3-alpine3.20 as builder

WORKDIR /go/src/genv
ENV CGO_ENABLED=0
ENV GOOS=linux

COPY . .
RUN go build -o /genv ./foo


FROM scratch as final

WORKDIR /go/src/genv

COPY --from=builder /genv .
