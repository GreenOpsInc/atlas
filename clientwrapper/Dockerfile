# syntax=docker/dockerfile:1
#Build in parent folder
FROM golang:1.16
WORKDIR /Atlas/ClientWrapper
COPY clientwrapper/go.mod .
COPY clientwrapper/go.sum .
COPY utilgo /Atlas/utilgo
RUN go mod download
COPY clientwrapper .
RUN CGO_ENABLED=0 go build -v ./atlasoperator/atlas_operator.go

FROM alpine:latest
COPY --from=0 /Atlas/ClientWrapper/atlas_operator /
CMD ["./atlas_operator"]
