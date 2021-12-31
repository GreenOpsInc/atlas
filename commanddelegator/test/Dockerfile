# syntax=docker/dockerfile:1
FROM golang:1.16
WORKDIR /CommandDelegator
COPY . .

FROM alpine:latest
COPY --from=0 /CommandDelegator/command_delegator_api /
CMD ["./command_delegator_api"]