# testCompletionFailedDueToTypo
This manifests contains pipeline with two linear steps `deploy_to_dev` and `deploy_to_int`. In the `verifyendpoints.sh`, there is a typo in line 7 & 8. It is tends to fail at `TestCompletionEvent` of `deploy_to_dev`.

Expect rules are also defined for this manifest.