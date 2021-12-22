# syntax=docker/dockerfile:1
FROM golang:1.16
WORKDIR /WorkflowTrigger
COPY . .

FROM alpine:latest
COPY --from=0 /WorkflowTrigger/workflowtrigger /
EXPOSE 8080
CMD ["./workflowtrigger"]