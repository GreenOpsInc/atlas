# team

## Create a team 

```
atlas team create <team name> [flags]
```

### Options

```
  -h, --help             help for create
  -p, --parent string    parent team name (default "na")
  -s, --schemas string   path to pipeline schemas JSON file
```

## Read a team's information

```
 atlas team read <team name> [flags]
```

### Options

```
  -h, --help             help for read
```


## Delete a team

```
atlas team delete <team name> [flags]
```

### Options

```
  -h, --help             help for delete
```

## Update a team's information (parent team and team name can be changed)

```
atlas team update team_name -p new_parent_team_name -n new_team_name
```
### Options

```
  -h, --help              help for update
  -n, --new-name string   new team name
  -p, --parent string     new parent team name (default "na")
```

## Options inherited from parent commands

```
--org string    override the default org name for a single command execution (default "org")
--url string    override the default atlas url for a single command execution  (default "localhost:8081")
```