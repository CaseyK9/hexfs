FROM golang:1.15-alpine AS build
RUN apk update && apk add --no-cache git
RUN apk --no-cache add ca-certificates

COPY . $GOPATH/src/hexfs
WORKDIR $GOPATH/src/hexfs
RUN go get -d ./...
RUN CGO_ENABLED=0 go build -o /bin/hexfs

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bin/hexfs /bin/hexfs
EXPOSE 3030 3031
ENTRYPOINT ["/bin/hexfs"]