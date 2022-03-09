# humanError
This manifests contains pipeline with two linear steps `deploy_to_dev` and `deploy_to_int`. In the `tests` of `deploy_to_int` there is a typo in line 20. It is tends to fail because of this typo and send a `FailureEvent`.

Expected rules are also defined for this manifests.