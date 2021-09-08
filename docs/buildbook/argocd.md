# Writing an ArgoCD Application

For more information on ArgoCD, check out the [docs](https://argo-cd.readthedocs.io/en/stable/).

## Configuration

The base configuration for an Application manifest is as follows:

    apiVersion: argoproj.io/v1alpha1
    kind: Application
    metadata:
      name: atlastestapp-dev
      namespace: argocd
    spec:
      project: default
      source:
        repoURL: <Git repo URL with application manifest>
        targetRevision: <branch name, Git revision hash, or tag>
        path: <path to application manifest file>
      destination:
        server: <External cluster IP, or https://kubernetes.default.svc if internal>
        namespace: <namespace to deploy application in>

This, accompanied by the actual application manifest (which contains the Kubernetes resources that need to be deployed), is all that is required from the ArgoCD application manifest. The deployment, management, stability, cleanup, and logging will all be automated by Atlas.

Any other configurations made to the manifest will be honored when deploying the ArgoCD application. Information on specifying configurations like selective synchronization, pruning, and more can be found in the [ArgoCD Sync Configurations](sync-configs.md) page.
