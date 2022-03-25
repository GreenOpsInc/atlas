### Generating the code

1. Run `export GOROOT=$(go env GOROOT)`
2. Clone https://github.com/kubernetes/code-generator
3. Put `atlas` project directory in the `greenopsinc.io` directory
4. Put `code-generator` directory in the same directory where `greenopsinc.io` is
5. Run code generation script:
```shell
  path-to-root-dir/code-generator/generate-groups.sh all \
    greenopsinc.io/atlas/pkg/client \
    greenopsinc.io/atlas/pkg/apis \
    "pipeline:v1" \ 
    -h path-to-root-dir/code-generator/hack/boilerplate.go.txt \ 
    -o path-to-root-dir
```
