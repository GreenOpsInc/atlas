# Security

## Authentication and Authorization

Both authentication and authorization are delegated to Argo CD. Atlas deploys out of the box with endpoints restricted by RBAC and authentication. Atlas shares Argo CD access tokens to make it simpler for users. The Argo CD RBAC policy can be [configured](https://argo-cd.readthedocs.io/en/stable/operator-manual/rbac/) however the operators desire, and that policy will then be used by Atlas.

## TLS Configuration

Communication between Atlas components, ArgoCD and Kafka is secured with self-signed certificates by default.

### Adding custom TLS configuration

Custom TLS certificate and private key should be provided via Kubernetes secret values.

Example:
```
kubectl create secret tls workflowtrigger-tls --cert ./cert.pem --key ./key.pem
```

Kubernetes TLS secret names:
```
Workflow Trigger:   workflowtrigger-tls
Client Wrapper:     clientwrapper-tls
Repo Server:        pipelinereposerver-tls
Command Delegator:  commanddelegator-tls
ArgoCD Repo Server: argocd-repo-server-tls (in the argocd namespace)
Kafka:              kafka-tls
```

Each service will listen for kubernetes secret changes and update their servers and clients with new TLS configuration.

### ArgoCD

ArgoCD TLS should be configured using `argocd-repo-server-tls` secret in the `argocd` namespace.

Example:
```
kubectl create secret tls argocd-repo-server-tls --cert ./cert.pem --key ./key.pem -n argocd
```

### Strimzi Kafka TlS

Create a Kafka User with Authentication TLS & Simple Authorization:
```
kubectl apply -f https://raw.githubusercontent.com/anoopl/strimzi-kafka-operator-authn-authz/master/kafka-user.yaml
```

Download the Cluster CA Cert and PKCS12 ( .p12) keys of User to use with Kafka Client:
```
kubectl get secret -n kafka my-cluster-cluster-ca-cert -o jsonpath='{.data.ca\.crt}' | base64 -d > kafla.cert.crt
kubectl get secret -n kafka my-cluster-cluster-ca-cert -o jsonpath='{.data.ca\.password}' | base64 -d > kafka.keystore.credentials
kubectl get secret -n kafka my-cluster-cluster-ca-cert -o jsonpath='{.data.ca\.p12}' | base64 -d > kafka.p12
```

Convert the ca.cert to truststore jks and user.p12 to Keystore.jks:
```
keytool -keystore kafka.truststore.jks -alias CARoot -import -file kafka.cert.crt
keytool -importkeystore -srckeystore kafka.p12 -srcstoretype pkcs12 -destkeystore kafka.keystore.jks -deststoretype jks
```

Create `kafka.cert.pem` file for Go services configuration:
```
openssl x509 -in kafka.cert.crt -out kafka.cert.pem -outform PEM
```

Copy contents of `kafka.keystore.credentials` file to the `kafka.key.credentials` and `kafka.truststore.credentials` files.

Create kubernets secret:
```
kubectl create secret generic kafka-tls \
  --from-file=./certs/kafka.keystore.jks \
  --from-file=./certs/kafka.truststore.jks \
  --from-file=./certs/kafka.cert.pem \
  --from-file=./certs/kafka.keystore.credentials \
  --from-file=./certs/kafka.truststore.credentials \
  --from-file=./certs/kafka.key.credentials
```
