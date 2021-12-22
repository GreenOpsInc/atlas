# Granular Management Capabilities

## No-Deploy

Users can mark environments or entire clusters as "no-deploy", which means that pipelines cannot run in the specified areas (this includes deployments, tests, etc.). This could be done for a demo, testing purposes, or any number of things. [Cluster update privileges](https://argo-cd.readthedocs.io/en/stable/operator-manual/rbac/) are required via Argo CD.

```
atlas cluster nodeploy <cluster_name> --name <team_name> --reason <reason> (--namespace can be added)
```

To remove the no-deploy:

```
atlas cluster nodeploy <cluster_name> --remove --name <team_name> --reason <reason> (--namespace can be added)
```

The name, reason, and namespace have to match exactly when removing a no-deploy. Only people with the cluster update permissions can make this change.

## Force Deploys

When necessary, force deploys allow properly permissioned users to deploy a manifest to an environment without a pipeline or testing. Users require the override application permissions to be able to do this.

```
atlas pipeline force-deploy <pipeline_name> --step <step_name> --pipelineRevisionHash <pipeline repo revision> --argoRevisionHash <argo manifest revision> --team <team_name> --repo <git_repo> --root <path_to_root>
```

The `--pipelineRevisionHash` and `--argoRevisionHash` flags are optional, and can be added as needed.
