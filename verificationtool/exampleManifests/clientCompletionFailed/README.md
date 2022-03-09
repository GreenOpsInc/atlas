# clientCompletionFailed
This manifests contains pipeline with two linear steps `deploy_to_dev` and `deploy_to_int`. In the `argo_manifest/manifest.yaml`, we have removed `image` name. It is tends to fail at `ClientCompletionEvent`.

Expected rules are also defined for this manifest.