# Supplementing Argo

[Argo CD](https://argo-cd.readthedocs.io/en/stable/) is one of the dominant toolchains today for managing Kubernetes-based applications.

Atlas is build directly on top of Argo, and provides a lot of features that help supplement it. The combination of Atlas and Argo creates a deployment pipeline-to-endpoint complete lifecycle management solution.

## How does Atlas supplement Argo?

### Pipelines

Atlas allows users to link ArgoCD applications into a continuous pipeline, making it easier to manage a large number of applications at scale.

Atlas also adds custom task, test, and plugin capabilities to Argo CD. Users can add in Python scripts, coordinate third-party tools, and do much more.

In general, Argo Workflows is most commonly used for data processing and ETL pipelines, specifically because of its ephemeral processing nature. To build deployment pipelines with Argo Workflows, users have to set up Argo Events and write custom scripts to trigger an ArgoCD deployment, which can be a complicated process.

### Automated Rollbacks

Atlas abstracts away the rollback and remediation process for Argo CD. Atlas automatically rolls back applications to previous stable versions in case of deployment failures, test failures, or degradation.

### Audit Logs

Atlas provides rich audit logs that provide great visibility into both pipelines and steps/applications. Atlas logs provide insights into current pipeline runs (what steps are in progress, or what steps have failed, rolled back, or been remediated). Audit logs will also show selective syncs, what applications are unhealthy, and broken task/test logs.

These deployment-centric logs give great visibility into deployments and pipelines, providing an easily accessible layer (and singular view) on top of custom logic & Argo.

### Reduced Complexity

Atlas helps provide simple transitions to Argo for new users, and also helps automate and simplify large-scale ArgoCD deployments.
