# cluster

## Create a cluster 

```
atlas cluster create <cluster name> [flags]
```

### Options

```
-h, --help        help for create
    --ip string   cluster IP address
    --port int    exposed port (default 80)
```

## Delete a cluster

```
 atlas cluster delete <cluster name> [flags]
```

### Options

```
  -h, --help   help for delete
```

## Options inherited from parent commands

```
--org string    override the default org name for a single command execution (default "org")
--url string    override the default atlas url for a single command execution  (default "localhost:8081")
```