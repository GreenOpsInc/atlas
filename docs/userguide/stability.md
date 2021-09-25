# Rollbacks and State Remediation

Atlas provides rollbacks and state remediation, which significantly help reduce deployment downtime and keep deployments stable.

Rollbacks can operate at any granularity, meaning that a failure can trigger the rollback of a specific step or an entire pipeline (or nothing at all, if configured appropriately). Although a rollback will ensure that a stable application is deployed, it is still considered a failure, and will not allow the pipeline to progress past the rolled-back step. Rollback limits can be set as well, ensuring that the deployed application won't be an entirely stale version.

State remediations, on the other hand, make it simple to monitor deployment state after a pipeline run. If an application starts to degrade, Atlas will try to stabilize the state by syncing the unhealthy/degraded resources. Just like with rollbacks, remediation limits can be set so degraded deployments that cannot be healed will not be stuck in a remediation loop. If the limit is and the pipeline is configured to do so, a rollback will be triggered to stabilize the deployment state.

Atlas' audit logs also provide great visibility into these rollbacks and state remediations, giving specific information about which revisions were deployed, how many times an application has rolled back, unhealthy resources, and more.
