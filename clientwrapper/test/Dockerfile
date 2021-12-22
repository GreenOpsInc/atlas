# syntax=docker/dockerfile:1
FROM golang:1.16
WORKDIR /ClientWrapper
COPY . .

FROM alpine:latest
COPY --from=0 /ClientWrapper/atlas_operator /
CMD ["./atlas_operator"]