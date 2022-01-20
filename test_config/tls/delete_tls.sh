#!/usr/bin/bash
export SHELL=/bin/bash

kubectl delete secret -n argocd argocd-repo-server-tls

kubectl delete secret workflowtrigger-tls

kubectl delete secret commanddelegator-tls

kubectl delete secret reposerver-tls

