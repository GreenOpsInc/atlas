#minikube has its own Docker daemon, run the following command to built the image to minikube:
#minikube docker-env
FROM golang:1.13
WORKDIR /ClientWrapper
COPY . .
RUN go build -i -v ./k8sdriver/kubernetes_driver.go
CMD ["./kubernetes_driver"]