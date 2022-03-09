# withRollBacks
This manifests contains pipeline with two linear steps `deploy_to_dev` and `deploy_to_int`. In step `deploy_to_dev`, `rollback_limit` is set to 1 and there is a bug in `verifyendpoints.sh` which will cause `TestCompletionEvent` failed and it will rollback to `healthy` manifest.

Expected rules are also defined for this manifest.


