FROM centurylink/ca-certs

ADD drone-swift /

ENTRYPOINT ["/drone-swift"]
