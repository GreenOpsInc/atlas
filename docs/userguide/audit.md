# Audit Logs & Visualization

The Atlas audit logs provide a rich methodology to understand and get great visibility into pipeline runs and specific steps in pipelines.

## Organization of Pipelines

Pipelines are organized and ordered using unique version numbers (UVNs). Each pipeline run, including rollbacks and state remediations, are bound to one specific UVN. Atlas stores all the logs of each pipeline run, and they can all be accessed to get full visibility and audit capabilities into the history of a pipeline.

## Pipeline-Level Logs

Pipeline-level statuses can be fetched using the CLI or API, which return payloads similar to this:

    {
        "progressingSteps": [
          "deploy_to_int"
        ],
        "stable": true,
        "complete": false,
        "cancelled": false,
        "failedSteps": []
    }

Or this in case of failures:

    {
        "progressingSteps": [],
        "stable": false,
        "complete": false,
        "cancelled": false,
        "failedSteps": [
          {
            "step": "deploy_to_int",
            "deploymentFailed": false,
            "brokenTest": "verifyendpoints.sh",
            "brokenTestLog": ".....",
          }
        ]
    }

The pipeline level status update shows a few key components:

- Which steps are currently running.
- Whether the deployments are stable or not. This applies for deployments made as a part of an pipeline run, but also for rollbacks and state remediations. If a state remediation takes place and the application is not able to stabilize, the `stable` variable will be false. If it is able to stabilize, it will be true.
- If the pipeline run was able to complete. Completion refers to whether the pipeline was able to finish all the steps successfully. If there was a failure at a specific step, or a rollback (even if the rollback is successful and the deployment is stable), `complete` will still be marked as false. State remediations generally happen past the scope of a pipeline run, and aren't factored into evaluating the completeness of a pipeline run.
- Whether the pipeline run was cancelled or not. If the pipeline run is cancelled, it is marked as incomplete as well.
- Failed steps/tests/tasks, if any. If any step(s) fail, they will be added to the `failedSteps` list. This list gives insight into the step that failed, whether the deployment of the application failed or not, and if any tasks or tests failed. If a task/test failed, the logs from the pod or pods will be collected and displayed in the `brokenTestLog`.

## Step-Level Logs

### Deployment Log

Step-level logs provide a much lower level view of a pipeline run, offering many specifics of a step. Here is an example of a step-level deployment log:

    {
        "type": "deployment",
        "pipelineUniqueVersionNumber": "d5046cc4-7da8-4286-9ce6-ba7fc5e92594",
        "rollbackUniqueVersionNumber": null,
        "uniqueVersionInstance": 0,
        "status": "SUCCESS",
        "deploymentComplete": true,
        "argoApplicationName": "atlastestapp-dev",
        "argoRevisionHash": "99c798442ebb8c58e0a8246f3a09627db3269170",
        "gitCommitVersion": "3650d61a65265d55329e3b0a30597a1bb27428df",
        "brokenTest": null,
        "brokenTestLog": null
    }

There are a few key components here:

- The pipeline unique version number.
- The rollback unique version number represents which pipeline UVN the step rolled back to. If no rollback occurs, the `rollbackUniqueVersionNumber` will be null.
- The unique version instance of the deployment. It essentially represents how many times the step has to be run before a stable configuration is reached.

If the step was successful and did not have to rollback, the `uniqueVersionInstance` would be 0. If the step does have a failure and is configured to rollback, the value would be an integer greater than or equal to 1 (1, 2, 3, etc.). The number represents how many times a step had to rollback (for the first rollback, the value would be 1; for the second rollback, the value of the second rollback log would be 2).
- The status of the deployment; the only possibilities are `SUCCESS`, `FAILURE`, `PROGRESSING`, and `CANCELLED`.
- The `deploymentComplete` represents whether the ArgoCD application specified in `argoApplicationName` was able to be deployed successfully or not.
- The `argoRevisionHash` represents what revision of the Kubernetes manifest linked in the ArgoCD Application was deployed.
- The `gitCommitVersion` represents what revision of the pipeline repository registered with Atlas was run.
- Failed tests/tasks, if any. If a test/task fails, the name will be listed in `brokenTest` and the logs from the pod(s) will be collected and displayed in the `brokenTestLog`.

### Remediation Log

Atlas also provides remediation logs which provide a history of the number of times a deployment degrades, which components degraded, and whether the state remediation attempt was successful. An example is below:

    {
        "type": "remediation",
        "pipelineUniqueVersionNumber": "d5046cc4-7da8-4286-9ce6-ba7fc5e92594",
        "uniqueVersionInstance": 1,
        "unhealthyResources": [
            "testapp", "testappsvc"
        ],
        "remediationStatus": "SUCCESS"
    }

As mentioned before, remediation attempts are still linked to pipelines, hence requiring a pipeline UVN. Remediation logs also show the remediation attempt number (first attempt, second attempt, etc.), along with the list of unhealthy resources that require syncing and whether the remediation attempt was successful.

## Argo Visualizations

As Atlas is built on Argo, all Argo visualizations and UI components are still running, and can be used to help visualize the current state of the application.
