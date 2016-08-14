# Docker image for the Drone build runner
#
#     CGO_ENABLED=0 go build -a -tags netgo
#     docker build --rm=true -t plugins/swift .

FROM alpine:3.3
RUN apk update && apk add ca-certificates mailcap && rm -rf /var/cache/apk/*
ADD drone-swift /bin/
ENTRYPOINT ["/bin/drone-swift"]
