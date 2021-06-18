#minikube has its own Docker daemon, run the following command to built the image to minikube:
#minikube docker-env
FROM golang:1.16
WORKDIR /ClientWrapper
COPY . .
RUN go build -i -v ./atlasoperator/atlas_operator.go
CMD ["./atlas_operator"]