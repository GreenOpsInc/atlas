# complexPipelineWith4Steps
This manifests contains pipeline with 4 steps `deploy_to_dev`, `deploy_to_int`, `deploy_to_prod` and `deploy_to_stage`. They are in this order,
```
    root(PipelineTriggerEvent)
    /                   \
deploy_to_dev       deploy_to_int
    |                   /
deploy_to_prod         /
    \                 /
     \               /
      \             /
       \           /
        \         /
       deploy_to_stage
```
This pipeline should pass all the verification order in the DAG, step status and pipeline status.