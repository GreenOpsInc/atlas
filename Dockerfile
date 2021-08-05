# syntax=docker/dockerfile:1
FROM golang:1.16
WORKDIR /ClientWrapper
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...
RUN go build -i -v ./atlasoperator/atlas_operator.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /ClientWrapper/atlas_operator ./
CMD ["./atlas_operator"]