# syntax=docker/dockerfile:1
FROM golang:1.16
WORKDIR /PipelineRepoServer
COPY . .

FROM alpine:latest
RUN apk update
RUN apk upgrade
RUN apk add bash
RUN apk add git
COPY --from=0 /PipelineRepoServer/pipelinereposerver /
EXPOSE 8080
CMD ["./pipelinereposerver"]