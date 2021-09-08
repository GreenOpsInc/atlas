# Supplementing Argo

Atlas is build directly on top of Argo, and provides a lot of features that help supplement it and provide enterprise features. The combination of Atlas and Argo creates a deployment pipeline to endpoint management complete lifecycle management solution.

## How does Atlas supplement Argo?

### Pipelines

Atlas allows users to link ArgoCD applications into a continuous pipeline, making it easier to manage a large number of applications at scale.

In general, Argo Workflows is most commonly used for data processing and ETL pipelines, specifically because of its ephemeral processing nature. To build deployment pipelines with Argo Workflows, users have to set up Argo Events, and configure events to trigger an ArgoCD deployment, which can be a complicated/difficult process.

### Automated Rollbacks

With Argo, rolling back applications is a manual process and require parsing through limited Kubernetes application manifest logs to find the correct revision to deploy. With Atlas, this challenge can be abstracted away. Atlas automatically rolls back applications to previous stable versions in case of deployment failures or test failures.

### Audit Logs

It can be difficult to keep clear and insightful deployment audit trails of Argo deployments. Atlas provides rich audit logs that provide great visibility into both pipelines and steps/applications. Atlas logs provide insights into current pipeline runs (what steps are in progress, or what steps have failed, rolled back, or been remediated). Audit logs will also show selective syncs, what applications are unhealthy, and broken task/test logs.

These deployment-centric logs give great visibility into deployments and pipelines, providing an easily accessible layer (and singular view) on top of ArgoCD and Argo Workflows.

### Reduced Complexity

Atlas helps provide simple transitions to Argo for new users, and also helps automate and simplify large-scale ArgoCD deployments.
