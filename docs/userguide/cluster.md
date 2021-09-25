# Setting Up and Managing Clusters

## Basics

Each step needs to be associated with a cluster where steps' commands will be delegated to. If all the steps in the pipeline need to interact with the same cluster, the destination cluster can be set at the pipeline level instead of the step level. Even if the destination cluster is set at the pipeline level, setting it at the step level will override the value for that specific step.

## Existing Clusters

The Kubernetes cluster where Atlas is deployed (we will call this the local cluster) does not need to be registered or set up. The local cluster is pre-registered as `kubernetes_local`.

## Adding New Clusters

When adding new clusters, there a few steps to follow:

First, the cluster should be registered using the Atlas API.

Next, a Client Wrapper instance should be deployed to said cluster. While only one Atlas control plane is required across all clusters, each cluster needs a delegate to receive and carry out commands. When deploying the Client Wrapper instance in the cluster, remember to update any parameters as necessary. For example, Atlas needs to communicate with Argo, but if the Argo operator is not local to the cluster the Client Wrapper is being deployed to, the `ARGOCD_SERVER` environment variable needs to be set to the correct destination, along with the `ARGOCD_USER_ACCOUNT` and `ARGOCD_USER_PASSWORD` variables.

It is also assumed that whatever cluster destination is listed in the Argo manifests have been pre-registered with the Argo APIs.
