# TLS Configuration

Communication between Atlas components, ArgoCD and Kafka is secured with selfsigned certificates by default.

### Adding custom TLS configuration

Custom TLS certificate and private key should be provided via Kubernetes secret values.

Example:
```shell
kubectl create secret tls workflowtrigger-tls --cert ./cert.pem --key ./key.pem
```

Kubernetes TLS secret names:
```shell
Workflow Trigger:   workflowtrigger-tls
Client Wrapper:     clientwrapper-tls
Repo Server:        pipelinereposerver-tls
Command Delegator:  commanddelegator-tls
ArgoCD Repo Server: argocd-repo-server-tls (in the argocd namespace)
Kafka:              kafka-tls
```

Each service will listen for kubernetes secret changes and update their servers and clients with new TLS configuration.

### Test certificates

`certs` directory contains self signed certificates for each Atlas service.

NOTE: those should be used only for testing purpose!

Those certificates could be applied and removed as kubernetes secrets by `./add_tls.sh` and `./delete_tls.sh` scripts respectively.

### ArgoCD

ArgoCD TLS should be configured using `argocd-repo-server-tls` secret in the `argocd` namespace.

Example:
```shell
kubectl create secret tls argocd-repo-server-tls --cert ./cert.pem --key ./key.pem -n argocd
```

### Kafka

We could configure TLS for Kafka service by providing TLS env variables in the docker-compose config file.

More info in docs: 

- Configuring TLS for Kafka and Pring: https://www.baeldung.com/spring-boot-kafka-ssl
- Configuring TLS for Kafka and Go: https://www.process-one.net/blog/using-tls-authentication-for-your-go-kafka-client/

If Kafka is secured with TLS config we should add cert and key value to the `kafka-tls` kubernetes tls secret:
```shell
kubectl create secret generic kafka-tls \
  --from-file=./certs/kafka.keystore.jks \
  --from-file=./certs/kafka.truststore.jks \
  --from-file=./certs/kafka.cert.pem \
  --from-file=./certs/kafka.keystore.credentials \
  --from-file=./certs/kafka.truststore.credentials \
  --from-file=./certs/kafka.key.credentials
```

#### Generating keystore and truststore

Create a certificate in a new keystore:
```shell
keytool -genkey -alias localhost -keyalg RSA -keypass SS28qmtOJH4OFLUP -keystore kafka.keystore.jks -storepass SS28qmtOJH4OFLUP -ext SAN=DNS:{FQDN}
```

Export the certificate to a file:
```shell
keytool -export -alias localhost -file kafka.key.cer -keystore kafka.keystore.jks -storepass SS28qmtOJH4OFLUP
```

Import the certificate into a new trust store:
```shell
keytool -import -v -trustcacerts -alias localhost -keypass SS28qmtOJH4OFLUP -file kafka.key.cer -keystore kafka.truststore.jks -storepass SS28qmtOJH4OFLUP
```

#### Convert keystore into PEM

Convert the JKS into PKCS12:
```shell
keytool -importkeystore -srckeystore kafka.keystore.jks \
   -destkeystore kafka.keystore.p12 \
   -srcstoretype jks \
   -deststoretype pkcs12
```

Encode keystore.p12 into a PEM file:
```shell
openssl pkcs12 -in kafka.keystore.p12 -out kafka.keystore.pem
```

After that we should get a single keystore.pem file with cert and key injected. 

To use this file as a kubernetes tls secret we should create two separate files `kafka.cert.pem` and `kafka.key_enc.pem` and copy respective contents from `kafka.keystore.pem` to those files.

`key_enc.pem` file is encrypted. To create `key.pem` file and use it in kubernetes tls secret run:
```shell
openssl pkey -in kafka.key_enc.pem -out kafka.key.pem
```

#### Mounting a local directory to minikube

```shell
minikube mount <path>/certs:/certs
```

### Strimzi Kafka TlS

Create a Kafka User with Authentication TLS & Simple Authorization:
```shell
kubectl apply -f https://raw.githubusercontent.com/anoopl/strimzi-kafka-operator-authn-authz/master/kafka-user.yaml
```

Download the Cluster CA Cert and PKCS12 ( .p12) keys of User to use with Kafka Client:
```shell
kubectl get secret -n kafka my-cluster-cluster-ca-cert -o jsonpath='{.data.ca\.crt}' | base64 -d > kafla.cert.crt
kubectl get secret -n kafka my-cluster-cluster-ca-cert -o jsonpath='{.data.ca\.password}' | base64 -d > kafka.keystore.credentials
kubectl get secret -n kafka my-cluster-cluster-ca-cert -o jsonpath='{.data.ca\.p12}' | base64 -d > kafka.p12
```

Convert the ca.cert to truststore jks and user.p12 to Keystore.jks:
```shell
keytool -keystore kafka.truststore.jks -alias CARoot -import -file kafka.cert.crt
keytool -importkeystore -srckeystore kafka.p12 -srcstoretype pkcs12 -destkeystore kafka.keystore.jks -deststoretype jks
```

Create `kafka.cert.pem` file for Go services configuration:
```shell
openssl x509 -in kafka.cert.crt -out kafka.cert.pem -outform PEM
```

Copy contents of `kafka.keystore.credentials` file to the `kafka.key.credentials` and `kafka.truststore.credentials` files.

Create kubernets secret:
```shell
kubectl create secret generic kafka-tls \
  --from-file=./certs/kafka.keystore.jks \
  --from-file=./certs/kafka.truststore.jks \
  --from-file=./certs/kafka.cert.pem \
  --from-file=./certs/kafka.keystore.credentials \
  --from-file=./certs/kafka.truststore.credentials \
  --from-file=./certs/kafka.key.credentials
```
