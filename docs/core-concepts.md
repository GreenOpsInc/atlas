# Core Concepts

This page will introduce the core concepts and terminology of Atlas.

- **Org**: The highest grouping possible. The org is an umbrella for all teams. In general, this is your organization or company's name.
- **Team**: The lowest level of pipeline grouping possible. Teams are where the pipelines are actually registered to. Team hierarchy is supported (teams can contain other teams).
- **Pipeline**: An ordered collection of steps, applications, tasks/tests, and deployment configurations.
- **Step**: A step is an individual piece of the pipeline. The step contains the ArgoCD application, application infrastructure, tests, tasks, and step dependencies.
- **ArgoCD Application**: In the context of Atlas, the ArgoCD application is just the YAML manifest.
- **Application Infrastructure**: Some pieces of deployments, like service meshes may not want to be managed or watched by ArgoCD. Application infrastructure is deployed separately from ArgoCD and not watched actively.
- **Task/Test**: Custom deployment logic that users can inject into their pipelines. Could be tests for custom business logic, a blocker that waits for 2 days before progressing to the next step, or anything else.
- **State Remediation**: Atlas watches applications and environments past the deployment or pipeline run. State remediation is when the application degrades and Atlas takes action to make the app healthy again.

To see how to put these pieces together and write your own pipelines, see the Build Book.
