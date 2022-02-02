# syntax=docker/dockerfile:1
#Build in parent folder
FROM golang:1.16
WORKDIR /Atlas/WorkflowTrigger
COPY workflowtrigger/go.mod .
COPY workflowtrigger/go.sum .
COPY utilgo /Atlas/utilgo
RUN go mod download
COPY workflowtrigger .
RUN CGO_ENABLED=0 go build -v ./

FROM alpine:latest
COPY --from=0 /Atlas/WorkflowTrigger/workflowtrigger /
EXPOSE 8080
CMD ["./workflowtrigger"]
