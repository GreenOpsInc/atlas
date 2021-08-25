# Writing a Pipeline

To quickly recap the core concepts, a step is an individual piece of the pipeline. A step contains an ArgoCD application, application infrastructure, tests, tasks, and step dependencies.

## Repo Structure

Atlas gets schemas and informations from an upstream Git repo, and expects necessary files to be present in the repository. The pipeline schema is expected to be stored in a file called `pipeline.yaml` directly located in the root folder specified when registering a pipeline. As an example:

      root_folder/ #Root folder entered when registering pipeline
        pipeline.yaml

## Schema

A pipeline schema would be as follows (all values are left blank):

      argo_version_lock:
      steps:

### Version Locking

ArgoCD applications link to upstream manifests containing the specific resources that need to be deployed. If multiple applications link to the same manifest, and an update is made to the manifest in the middle of a pipeline run, state-related complications can arise. If users want to avoid this, they can set `argo_version_lock` to `true`, which will ensure that the steps with ArgoCD applications linked to the same manifests will use the same version. The default value is `false`.

### Steps

Steps are added as an array under the `steps` variable. Specifics on what a step schema looks like can be found in [Writing a Step](step.md).

### Example

An example of a fully filled out pipeline is shown below. Remember that all variables have defaults and do not have to be filled out.

    argo_version_lock: true
    steps:
    - name: deploy_to_dev
      application_path: dev/testapp.yml
      additional_deployments: dev/istio_config.yml
      tests:
      - path: "tests/verifyendpoints.sh"
        type: inject
        before: false
        variables:
          SERVICE_INTERNAL_URL: testapp.dev.svc.cluster.local
    - name: deploy_to_int
      application_path: int/testapp.yml
      additional_deployments: int/istio_config.yml
      tests:
      - path: "tests/verifyendpoints.sh"
        type: inject
        before: false
        variables:
          SERVICE_INTERNAL_URL: testapp.int.svc.cluster.local
      dependencies:
      - deploy_to_dev
