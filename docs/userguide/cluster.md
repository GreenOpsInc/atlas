# Setting Up and Managing Clusters

## Basics

Each step needs to be associated with a cluster where steps' commands will be delegated to. If all the steps in the pipeline need to interact with the same cluster, the destination cluster can be set at the pipeline level instead of the step level. Even if the destination cluster is set at the pipeline level, setting it at the step level will override the value for that specific step.

## Existing Clusters

The Kubernetes cluster where Atlas is deployed (we will call this the local cluster) does not need to be registered or set up. The local cluster is pre-registered as `in-cluster`.

## Adding New Clusters

Atlas shares cluster management with Argo CD. If the cluster context exists in kubectl, it can easily be registered via Atlas with:

    atlas cluster add <CONTEXTNAME>

If the cluster is already registered with Argo CD but not with Atlas, add the `server` flag to only register it with Atlas:

    atlas cluster add <CONTEXTNAME> --server <IP address>

Next, a Client Wrapper instance should be deployed to said cluster. While only one Atlas control plane is required across all clusters, each cluster needs a delegate to receive and carry out commands. When deploying the Client Wrapper instance in the cluster, remember to update any parameters as necessary. For example, Atlas needs to communicate with Argo, but if the Argo operator is not local to the cluster the Client Wrapper is being deployed to, the `ARGOCD_SERVER` environment variable needs to be set to the correct destination, along with the `ARGOCD_METRICS_SERVER_ADDR`, `ARGOCD_USER_ACCOUNT`, and `ARGOCD_USER_PASSWORD` variables. The `WORKFLOW_TRIGGER_SERVER_ADDR` and `COMMAND_DELEGATOR_URL` may also need to be updated. `CLUSTER_NAME` should be set to the name of the cluster the Client Wrapper will be querying information for.

It is also assumed that whatever cluster destination is listed in the Argo manifests have been pre-registered with the Argo APIs.

The delegate-specific manifest can be found [here](https://github.com/GreenOpsInc/atlas/blob/main/manifest/cluster/install_delegate.yaml). The delegate can be installed using:

    kubectl apply -f https://raw.githubusercontent.com/GreenOpsInc/atlas/main/manifest/cluster/install_delegate.yaml -n atlas

Create the cluster in Atlas:

```
atlas cluster create cluster_name --ip 192.0.2.42 --port 9376
```
