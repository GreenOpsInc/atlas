# Script Injection

Script injection essentially takes a bash script and injects it into a Kubernetes Job. The schema looks as follows:

    path: #path to bash script
    type: inject
    before: #before or after the deployment
    image:
    namespace:
    variables: #key/values to pass in as environment variables

The `image` variable specifies which base image to use for injecting the script. The default is alpine.

The `namespace` variable specifies which namespace to run the resource in. The default is the "default" namespace.
