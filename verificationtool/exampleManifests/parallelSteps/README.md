# parallelSteps
This manifests contains pipeline with two parallel steps named `deploy_to_dev` and `deploy_to_int`. It tends to pass all the order, step status and pipeline status verification without any failed Events.

`deploy_to_int` is not depend on `deploy_to_dev` which means events from `deploy_to_dev` and `deploy_to_int` can receive in any order.