# status

## Fetch the status of a pipeline or specific pipeline step 

```
atlas status <pipeline name> [flags]
```

##  Required Flags

```
--team string   team name
```

### Options

```
-c, --count int     count (default 15)
-h, --help          help for status
-s, --step string   step name
    --team string   team name
-u, --uvn string    Pipeline UVN (default "LATEST")
```


## Options inherited from parent commands

```
--org string    override the default org name for a single command execution (default "org")
--url string    override the default atlas url for a single command execution  (default "localhost:8081")
```