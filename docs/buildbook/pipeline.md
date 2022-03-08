# Writing a Pipeline

To quickly recap the core concepts, a step is an individual piece of the pipeline. A step contains an ArgoCD application, application infrastructure, tests, tasks, and step dependencies.

You can find examples of pipelines [here](https://github.com/GreenOpsInc/atlasexamples).

## Repo Structure

Atlas gets schemas and informations from an upstream Git repo, and expects necessary files to be present in the repository. The pipeline schema is expected to be stored in a file called `pipeline.yaml` directly located in the root folder specified when registering a pipeline. As an example:

      root_folder/ #Root folder entered when registering pipeline
        pipeline.yaml

## Schema

A pipeline schema would be as follows (all values are left blank):

      argo_version_lock:
      cluster_name:
      steps:

### Version Locking

ArgoCD applications link to upstream manifests containing the specific resources that need to be deployed. If multiple applications link to the same manifest, and an update is made to the manifest in the middle of a pipeline run, state-related complications can arise. If users want to avoid this, they can set `argo_version_lock` to `true`, which will ensure that the steps with ArgoCD applications linked to the same manifests will use the same version. The default value is `false`.

### Destination Cluster

Each step needs to be associated with a cluster where steps' commands will be delegated to. If all the steps in the pipeline need to interact with the same cluster, the `cluster_name` can be set at the pipeline level instead of the step level. Even if `cluster_name` is set at the pipeline level, setting it at the step level will override the value for that specific step. With the exception of the local Kubernetes cluster (named `in-cluster`), clusters need to be registered with Atlas before being added as destinations in pipelines and steps--see the [cluster API documentation](../userguide/cluster.md) for more on registering new clusters.

There is no default value, `cluster_name` has to be set at either the pipeline or step level.

### Steps

Steps are added as an array under the `steps` variable. Specifics on what a step schema looks like can be found in [Writing a Step](step.md).

### Example

An example of a fully filled out pipeline is shown below. Remember that all variables have defaults and do not have to be filled out.

    argo_version_lock: true
    cluster_name: in-cluster
    steps:
    - name: deploy_to_dev
      application_path: dev/testapp.yml
      additional_deployments: dev/istio_config.yml
      tests:
      - path: "tests/verifyendpoints.sh"
        type: inject
        image: alpine
        commands: [sh, -c, ./verifyendpoints.sh]
        before: false
        variables:
          - name: SERVICE_INTERNAL_URL
            value: testapp.dev.svc.cluster.local
    - name: deploy_to_int
      application_path: int/testapp.yml
      additional_deployments: int/istio_config.yml
      tests:
      - path: "tests/verifyendpoints.sh"
        type: inject
        image: alpine
        commands: [sh, -c, ./verifyendpoints.sh]
        before: false
        variables:
          - name: SERVICE_INTERNAL_URL
            value: testapp.dev.svc.cluster.local
      dependencies:
      - deploy_to_dev
