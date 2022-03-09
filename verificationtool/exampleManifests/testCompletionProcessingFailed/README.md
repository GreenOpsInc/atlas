# testCompletionProcessingFailed
This manifests contains pipeline with two linear steps `deploy_to_dev` and `deploy_to_int`. In the `verifyendpoints.sh`, we are returning `exit 1` to fail a test processing and raise a `TestCompletonEvent`. 

Expect rules are also defined for this manifest.