#!/usr/bin/bash
export SHELL=/bin/bash

kubectl create -n argocd secret tls argocd-repo-server-tls \
  --cert=./certs/argo_reposerver_cert.pem \
  --key=./certs/argo_reposerver_key.pem

kubectl create secret tls workflowtrigger-tls \
  --cert ./certs/workflowtrigger_cert.pem \
  --key ./certs/workflowtrigger_key.pem

kubectl create secret tls commanddelegator-tls \
  --cert ./certs/commanddelegator_cert.pem \
  --key ./certs/commanddelegator_key.pem

kubectl create secret tls reposerver-tls \
  --cert ./certs/reposerver_cert.pem \
  --key ./certs/reposerver_key.pem
