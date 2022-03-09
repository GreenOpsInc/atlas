# failedWithRollbacks
This manifests contains pipeline with two linear steps `deploy_to_dev` and `deploy_to_int`. This manifest have an error in `verifyendpoints.sh` and it should failed at `TestCompletionEvent` and then rollback to healthy manifest.

Expected rules are also defined for this manifest.


