# Introduction

## What is Atlas?

Atlas is an open-source deployment pipeline platform for Kubernetes applications. Atlas makes it easy to run a continuous deployment pipeline across all environments and clusters, add custom tasks, tests, and plugins, and automatically rollback or remediate your applications. Atlas also has a management plane, allowing highly granular control over deployments and environments. Atlas is built on top of [Argo CD](https://argo-cd.readthedocs.io/en/stable/), one of the dominant toolchains today for managing Kubernetes-based applications.

## Why Atlas?

Atlas provides a very powerful pipeline-to-endpoint management solution, and creates an enterprise-grade deployment solution that can be used at any scale. Through simple GitOps-based schemas, users can automate deployments across all their clusters and environments while easily adding in their custom deployment logic and ensuring no downtime.

Even users that have never used ArgoCD can easily adopt Atlas. With the simplest possible ArgoCD schema, Atlas can do the rest.

## Key Features

### Pipelines

Users can have linear or branching pipeline steps across any number of environments or clusters, allowing the simple creation of complex pipelines. Processing is durable, asynchronous, and error accommodating, meaning that pipelines run quickly, in a lightweight manner, and can withstand deployment errors or services being down.

### Automated Tasks/Tests

Users can add in tasks, tests, and plugins with custom logic to pipelines, and Atlas will manage their lifecycles as desired. K8S, Python, Bash, Argo Workflows, Terraform - any type of task, test, or integration can be run in any order, and with custom variables injected. Task/test logs will be also saved and shown to users in case of failures.

There are also a registry of different types of tasks and tests that users may find useful. You can find them in the Task/Test Library. If you don't see what you need, feel free to write it! Contributions are always welcome.

### Audit Logs

Atlas keeps rich audit logs for all deployments, steps, pipelines, and every other action that takes place. Users can get any level of visibility desired, be it team level, pipeline level, or step level. Users can see what steps are progressing, what steps have failed, what resources or tests are unhealthy, and so much more.

### Health, Automated Rollbacks, & State Remediation

Atlas will check if the deployments and tests/tasks ran successfully, and roll back the application in case of failure. Rollbacks can be configured at the step or pipeline level.

Atlas also watches applications and environments past a pipeline run, and automatically fixes the state in case of unhealthy or degrading resources.

### Fine Grained Management

Users have direct access pipelines and applications. Even if a cluster or environment is inaccessible, pipelines can still be run and applications in other environments can still be interacted with. Clusters/environments can be marked as no-deploy to stop any changes from occurring, different versions of pipelines and applications can be run, and deployments can be forced (in specific cases & with the right permissions).

### Security & Argo CD

Atlas can be configured with TLS, and delegates all authentication and authorization to Argo CD, ensuring security. All Argo features are still fully accessible, so Argo-lovers have nothing to worry about!
