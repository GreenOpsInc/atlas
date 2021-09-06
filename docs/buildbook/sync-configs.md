# ArgoCD Sync Configurations

As a part of pipeline steps, Atlas can create, update, and synchronize ArgoCD applications. Atlas will honor configurations for the ArgoCD application. Specific configurations are simple and effortless to set up.

Sync configurations are set up using Kubernetes annotations in the ArgoCD application manifest. All annotations are optional and do not have to be set up.

Any other sync-related configurations set up in individual resources or the ArgoCD application manifest will be honored.

## Resource Pruning

To configure stale resources to be pruned automatically, set the annotation `atlas-argo-sync-prune` to `"true"`. The default is `"false"`.

## Sync Strategy

To configure the ArgoCD sync strategy, set the annotation `atlas-argo-sync-strategy` to `"apply"` or `"hook"`.

To configure whether Kubernetes should force the applying of the resources, set the annotation `atlas-argo-sync-strategy-force` to `"true"`. The default is `"false"`.

### Selective Sync

To configure selective syncs during deployments (instead of syncing the entire application every time), set the annotation `atlas-argo-sync-selective` to `"true"`. The default is `"false"`.
