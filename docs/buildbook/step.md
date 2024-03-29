# Writing a Step

To quickly recap the core concepts, a step is an individual piece of the pipeline. A step contains an ArgoCD application, application infrastructure, tests, tasks, and step dependencies.

You can find examples of pipelines [here](https://github.com/GreenOpsInc/atlasexamples).

## Repo Structure

Atlas gets schemas and informations from an upstream Git repo, and expects necessary files to be present in the repository. Atlas will clone the entire repo, so files don't have to be restricted to the root folder. The paths specified in the schema must be relative to the root folder however. As an example:

    file1.yml
      root_folder/ #Root folder entered when registering pipeline
        folder/
          file2.yml

In the repo hierarchy, a path to file1.yml would be `../file1.yml` and a path to file2.yml would be `folder/file2.yml`, as all paths are relative to the root folder.

## Schema

A step schema would be as follows (all values are left blank):

    - name: #name of step
      cluster_name:
      application_path:
      additional_deployments:
      rollback:
      remediation_limit:
      tests:
      dependencies:

### Destination Cluster

Each step needs to be associated with a cluster where steps' commands will be delegated to. If all the steps in the pipeline need to interact with the same cluster, the `cluster_name` can be set at the pipeline level instead of the step level. Even if `cluster_name` is set at the pipeline level, setting it at the step level will override the value for that specific step. With the exception of the local Kubernetes cluster (named `in-cluster`), clusters need to be registered with Atlas before being added as destinations in pipelines and steps--see the [cluster API documentation](../userguide/cluster.md) for more on registering new clusters.

There is no default value, `cluster_name` has to be set at either the pipeline or step level.

### ArgoCD Application

The ArgoCD application file is expected to be added to the repo. The `application_path` variable is where the manifest file can be found. The default is empty. If nothing is entered, no application will be deployed. You can read more about setting up the ArgoCD Application [here](argocd.md).

### Application Infrastructure

If any resources are meant to be deployed outside of the context of ArgoCD (eg: service mesh), they can be deployed as a part of the application infrastructure. Add the file path to the `additional_deployments` variable. The default is empty. If nothing is entered, no application infrastructure will be deployed.

### Rollback

Step level rollbacks are provided. If the deployment of the application infrastructure or ArgoCD application fails, or if a test or task fail, the step can rollback to the previous stable version if `rollback_limit` is set to a value greater than 0. The default value is `0`. The step will only attempt rolling back `rollback_limit` number of times.

### Remediation Limit

If a deployment starts degrading after a pipeline run, Atlas can try to re-sync the current state and make it healthy. It will try `remediation_limit` times to fix the state. If the limit is met and `rollback_limit` is set to a value greater than 0, the application will rollback to the previous stable state. The default value for `remediation_limit` is `0`.

### Tasks/Tests

Tasks and tests are added as an array under the `tests` variable. Any type of operation can be added here, ranging from a Python script to an Argo Workflow. There are a number of different types of tasks and tests, which can be found in the [Tasks/Tests Registry](../tasktestregistry/overview.md). If you don't see something you want here, feel free to add it - it's easy!

### Dependencies

Steps are ordered into a DAG (directed acyclic graph), which is why complex and parallel execution is possible. Enter the direct step(s) that have to execute successfully for the current step to be run in `dependencies`. Values are entered as an array of strings, with the string being the names of steps.

### Example

An example of a fully filled out step is shown below. Remember that all variables have defaults and do not have to be filled out.

    - name: deploy_to_int
      cluster_name: in-cluster
      application_path: int/testapp.yml
      additional_deployments: int/istio_config.yml
      rollback_limit: 2
      remediation_limit: 3
      tests:
      - path: "tests/verifyendpoints.py"
        type: inject
        image: python:latest
        commands: [python, verifyendpoints.py]
        before: false
        variables:
          - name: SERVICE_INTERNAL_URL
            value: testapp.dev.svc.cluster.local
      - path: "tests/behavioral.yaml"
        type: custom
        before: false
        variables:
          - name: SERVICE_INTERNAL_URL
            value: testapp.dev.svc.cluster.local
      dependencies:
      - deploy_to_dev
