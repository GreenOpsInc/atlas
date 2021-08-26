# Contributing

There is a test interface in the Atlas data model that contains a few requirements for a task or test. Namely, a `path`, a `type`, specifying whether it happens `before` or after a deployment, and any `variables` that may need to be passed in. Obviously, `before` and `variables` have defaults, meaning they can be ignored.

With the exception of these few requirements, tasks and tests can have any scope and do any number of things. As is, tasks and tests are made using the `KubernetesCreationRequest` object, but feel free to go beyond that scope and add in custom logic or requests, as long as it doesn't affect other functionality.

 When making a PR for a new task or test for the registry, add "Task/Test Registry" into the description, and add in documentation describing how to use it.
