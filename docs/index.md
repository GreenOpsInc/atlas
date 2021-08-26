# Introduction

## What is Atlas?

Atlas is a GitOps deployment pipeline tool built on top of ArgoCD. It allows users to link multiple ArgoCD applications and frictionlessly create simple or complex pipelines, complete with tasks, tests, comprehensive audit logs, and automatic rollbacks/state remediation.

## Why Atlas?

Atlas provides a very powerful & pipeline-to-endpoint management solution, and creates an enterprise-grade deployment solution that can be used at any scale. Through simple GitOps-based schemas, users can automate deployments across all their clusters and environments while easily adding in their custom deployment logic.

Even users that have never used ArgoCD can easily adopt Atlas. Event with the simplest possible ArgoCD schema, Atlas can do the rest.

## Key Features

### Pipelines

Users can have parallel and linear pipeline steps, allowing the simple creation of complex pipelines. Processing is durable, asynchronous, and error accommodating, meaning that pipelines run quickly, in a lightweight manner, and can withstand deployment errors or services being down.

### Automated Tasks/Tests

Users can add in tasks and tests with custom logic to pipelines, and Atlas will manage their lifecycles as desired. Tasks/tests can be run before deployments, after deployments, and can have any number of custom variables injected. Task/test logs will be also saved and shown to users in case of failures.

There are also a registry of different types of tasks and tests that users find useful. You can find them in the Task/Test Library. If you don't see what you need, feel free to write it! Contributions are always welcome.

### Audit Logs

Atlas keeps rich audit logs for all deployments, steps, pipelines, and every other action that takes place. Users can get any level of visibility desired, be it team level, pipeline level, or step level. Users can see what steps are progressing, what steps have failed, what resources or tests are unhealthy, and so much more.

### Automated Rollbacks & State Remediation

Atlas will check if the deployments and tests/tasks ran successfully, and roll back the application in case of failure. Rollbacks can be configured at the step or pipeline level.

Atlas also watches applications and environments past a pipeline run, and fixes the state in case of unhealthy or degrading resources.

## How does Atlas supplement ArgoCD?

### Pipelines

Atlas allows users to link ArgoCD applications into a continuous pipeline, making it easier to manage a large number of applications at scale.

### Automated Rollbacks

Atlas will automatically rollback applications to previous stable versions in case of deployment failures or test failures.

### Audit Logs

Atlas provides rich audit logs that provide great visibility into both pipelines and steps/applications. Atlas logs provide insights into current pipeline runs (what steps are in progress, or what steps have failed, rolled back, or been remediated). Audit logs will also show selective syncs, what applications are unhealthy, and broken task/test logs.

### Reduced Complexity

Atlas can help provide simple transitions to ArgoCD for new users, and also help automate and simplify large-scale ArgoCD deployments.
