# simpleLinearPipeline
This manifests contains simple pipeline with two steps named `deploy_to_dev` and `deploy_to_int`. It tends to pass all the order, step status and pipeline status verification without any failed Events.

`deploy_to_int` is depend on `deploy_to_dev` which means after receiving all the events of `deploy_to_dev` we will expect events of `deploy_to_int`. 