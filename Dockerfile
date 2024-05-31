FROM public.ecr.aws/docker/library/golang:1.22.3-alpine3.20 as builder

WORKDIR /go/src/genv
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY . .
RUN go build -o /foo ./foo


FROM scratch as final

WORKDIR /go/src/genv

COPY --from=builder /foo /
CMD [ "/foo" ]