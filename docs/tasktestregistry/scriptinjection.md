# Script Injection

Injection allows users to take any type of action, setup script, or test suite (written in any language) and inject it in a Kubernetes Job. The schema looks as follows:

    path: #path to bash script
    type: inject
    image: #image name
    commands:
    before: #before or after the deployment
    namespace:
    variables: #key/values to pass in as environment variables

The `image` variable specifies which base image to use.

The `commands` variable specifies what commands to run when the image starts.

The `namespace` variable specifies which namespace to run the resource in. The default is the "default" namespace.
