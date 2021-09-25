# Getting Started

## Requirements

- The kubectl command-line tool
- ArgoCD installed in the cluster

## 1. Set up Kafka

In case you do not have a Kafka instance, you can use [Strimzi](https://strimzi.io/) (a CNCF cloud-native Kafka project) to set up a Kafka cluster in your Kubernetes environment. If you are using minikube, it is recommended to have 4 GB memory for the VM, as 2 GB isnt always enough.

Set up Strimzi using:

    kubectl create namespace kafka
    kubectl create -f 'https://strimzi.io/install/latest?namespace=kafka' -n kafka
    kubectl apply -f https://raw.githubusercontent.com/GreenOpsInc/atlas/main/manifest/kafka.yaml -n kafka
    kubectl wait kafka/atlas --for=condition=Ready --timeout=300s -n kafka

These commands will set up the Strimzi Kafka operator in the Kafka namespace, create a single-broker Kafka cluster, and wait for the cluster to be set up correctly.

## 2. Set up Atlas

    kubectl create namespace atlas
    kubectl apply -f https://raw.githubusercontent.com/GreenOpsInc/atlas/main/manifest/atlas.yaml -n atlas

This will create the atlas namespace and provision the atlas control plane in said namespace.

During deployment, the Atlas Client Wrapper (which acts as the delegate in clusters and communicates with Argo) reads the ArgoCD ConfigMap to get the admin username and password. If you want to provide a different set of credentials, add environment variables `ARGOCD_USER_ACCOUNT` and `ARGOCD_USER_PASSWORD` to the Client Wrapper Deployment in the manifest file.

## 3. Download Atlas CLI

## 4. Access the Atlas API

By default, the Atlas API is not exposed via an external IP. There are a few ways to access the API server.

### Load Balancer

Patch the Workflow Trigger's Service spec to use a LoadBalancer:

    kubectl patch svc workflowtrigger -n atlas -p '{"spec": {"type": "LoadBalancer"}}'

### Port Forwarding

    kubectl port-forward svc/workflowtrigger -n atlas 8081:8080

Running the command above will allow the Atlas API to be accessible at localhost:8081.

## 5. Set Up and Run Your First Pipeline

Now that you have access to the Atlas API and have the set up completed, you can now create a team and run a pipeline.

First, create a team using the CLI:

    atlas team create exampleTeam

The team is what has ownership over the pipeline. Multiple pipelines can be created per team, and teams can also be created under other teams.

Atlas follows a GitOps approach to pipeline management. We have set up an [example pipeline repository](https://github.com/GreenOpsInc/atlasexamples/tree/main/basic) that you can run. The repository contains the pipeline schema (information on how many steps there are, what each step does), tests, and the ArgoCD deployment manifest/Kubernetes manifest. For the sake of simplicity, all the files are on the same level. For more specifics on the schema and pipeline structure, check out the [Build Book](buildbook/step.md).

The pipeline has two steps, one of which will deploy an application to the `dev` namespace, and another which will deploy an application to the `int` namespace. The pipeline will run a test after each deployment to make sure the deployment is stable and available.

Create the pipeline:

    atlas pipeline create examplePipeline --gitRepo https://github.com/GreenOpsInc/atlasexamples.git --teamName exampleTeam --pathToRoot basic/

Creating a new pipeline will automatically trigger the pipeline run.

## 6. Check the Audit Logs

You can view the status of the pipeline run by running:

    atlas logs get examplePipeline

You will now be able to see the pipeline status, which will share what steps are currently in progress, whether the deployed steps are stable or not, if the pipeline run is complete, if the pipeline run was cancelled, and if a step failed (and if it did, the specific issue it had). A sample response is as follows:

    {
        "progressingSteps": [
          "deploy_to_dev"
        ],
        "stable": true,
        "complete": false,
        "cancelled": false,
        "failedSteps": []
    }

You can also get step-specific logs for a pipeline, which contain much more granular information.

You can view the step-specific logs with:

    atlas logs get examplePipeline --step deploy_to_dev

Step-specific logs contain information like the application name, Argo revision, Git revision, whether the application was rolled back, whether a test broke (and what the logs were if one did), etc. A sample response is as follows:

    [
        {
            "type": "deployment",
            "pipelineUniqueVersionNumber": "568e1640-97de-4c22-b140-a459b6453402",
            "rollbackUniqueVersionNumber": null,
            "uniqueVersionInstance": 0,
            "status": "PROGRESSING",
            "deploymentComplete": true,
            "argoApplicationName": "atlastestapp-dev",
            "argoRevisionHash": "99c798442ebb8c58e0a8246f3a09627db3269170",
            "gitCommitVersion": "3650d61a65265d55329e3b0a30597a1bb27428df",
            "brokenTest": null,
            "brokenTestLog": null
        }
    ]

The Argo visualization tools are also still active. The Argo UI can be used to see the state of the application/deployment.

![Placeholder](https://argoproj.github.io/argo-cd/assets/guestbook-tree.png)
