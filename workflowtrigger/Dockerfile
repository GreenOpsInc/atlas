# syntax=docker/dockerfile:1
FROM golang:1.16
WORKDIR /WorkflowTrigger
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -v ./

FROM alpine:latest
#RUN apk --no-cache add ca-certificates
COPY --from=0 /WorkflowTrigger/workflowtrigger /
EXPOSE 8080
CMD ["./workflowtrigger"]