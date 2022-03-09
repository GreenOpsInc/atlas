# noDeployments
This manifests contains pipeline with two linear steps `deploy_to_dev` and `deploy_to_int`. In both the step, we have removed the deployment step. 
It should fail at `TestCompletionEvent`

Expect rules are also defined for this manifest. Expected step status should fail, `status: SUCCESS` should be `status: FAILURE`.