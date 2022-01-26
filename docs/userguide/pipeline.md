# Setting Up and Managing Pipelines

## Running Pipelines After Creation

Pipelines can be run again using syncs. When a sync is triggered, the latest version of the pipeline repo will be fetched and subsequently run.

```
atlas pipeline sync pipeline_name --team team_name  --repo git_repo --root path_to_root (No flags specified means open access)
```

If there is a specific revision that you want to deploy, you can add the `--pipelineRevisionHash <pipeline repo revision>` flag.

## Queueing Pipeline Runs

Only one instance of a pipeline can be running at any given time. If a pipeline run is triggered while it is already running, the triggered run will be queued up and run once the current pipeline has finished running or is cancelled.

## Cancelling Pipeline Runs

Atlas allows users to cancel pipelines that are currently running. When a cancellation request is sent, each step in the current pipeline will be marked as cancelled. When a pipeline run is cancelled, all the events sent in the future related to that specific pipeline run will be marked as stale and ignored, ensuring that there will not be "leakage" between pipeline runs.

```
atlas pipeline cancel pipeline_name --team team_name
```
