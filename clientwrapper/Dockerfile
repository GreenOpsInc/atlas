# syntax=docker/dockerfile:1
FROM golang:1.16
WORKDIR /ClientWrapper
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -v ./atlasoperator/atlas_operator.go
#RUN go build -i -v ./atlasoperator/atlas_operator.go

FROM alpine:latest
#RUN apk --no-cache add ca-certificates
COPY --from=0 /ClientWrapper/atlas_operator /
CMD ["./atlas_operator"]
