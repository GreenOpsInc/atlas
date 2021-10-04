# pipeline

##  Required Flags

```
--team string   team name
```

## Create a pipeline

Pipelines will be run upon creation.

```
 atlas pipeline create <pipeline name> [flags]
```

### Options

```
  -h, --help              help for create
  -p, --password string   git password
      --repo string       git repo url
      --root string       path to root
  -t, --token string      name of git cred token
  -u, --username string   git username
```

## Read a pipeline's information

```
atlas pipeline read <pipeline name> [flags]
```

### Options

```
  -h, --help              help for create
  -p, --password string   git password
      --repo string       git repo url
      --root string       path to root
  -t, --token string      name of git cred token
  -u, --username string   git username
```

## Delete a pipeline

```
atlas pipeline delete <pipeline name> [flags]
```
### Options

```
 -h, --help   help for delete
```

## Update a pipeline

```
atlas pipeline update <pipeline name> [flags]
```

### Options

```
   -h, --help             help for update
  -p, --password string   git password
      --repo string       git repo url
      --root string       path to root
  -t, --token string      Name of git cred token
  -u, --username string   git username
```

## Sync a pipeline

```
atlas pipeline sync <pipeline name> [flags]
```

### Options

```
  -h, --help              help for sync
  -p, --password string   git password
      --repo string       git repo url
      --root string       path to root
  -t, --token string      Name of git cred token
  -u, --username string   git username
```

## Cancel a pipeline

```
atlas pipeline cancel <pipeline name> [flags]
```

### Options

```
-h, --help   help for cancel
```



## Options inherited from parent commands

```
--org string    override the default org name for a single command execution (default "org")
--url string    override the default atlas url for a single command execution  (default "localhost:8081")
```
