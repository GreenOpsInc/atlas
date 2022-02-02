# Argo Workflows

This task allows an Argo Workflow to run as a part of the pipeline:

    path: #path to bash script
    type: ArgoWorkflowTask
    before: #before or after the deployment

If Argo Workflows is authenticated via Kubeconfig, ensure that the Client Wrapper has access to it. In case auth is configured via access token, the token has to be passed in as an environment variable (`ARGO_TOKEN`) to the Atlas Client Wrapper.
