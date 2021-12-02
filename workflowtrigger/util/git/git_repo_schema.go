package git

import (
	"encoding/json"
)

const (
	gitRepoName string = "Git Repo"
	rootName    string = "Root"
	gitCredName string = "Git Cred"
)

type GitRepoSchema struct {
	GitRepo    string  `json:"gitRepo"`
	PathToRoot string  `json:"pathToRoot"`
	GitCred    GitCred `json:"gitCred"`
}

func New(gitRepo string, pathToRoot string, gitCred GitCred) GitRepoSchema {
	return GitRepoSchema{
		GitRepo:    gitRepo,
		PathToRoot: pathToRoot,
		GitCred:    gitCred,
	}
}

func (g *GitRepoSchema) SetGitRepo(gitRepo string) {
	g.GitRepo = gitRepo
}

func (g *GitRepoSchema) SetPathToRoot(pathToRoot string) {
	g.PathToRoot = pathToRoot
}

func (g *GitRepoSchema) SetGitCred(gitCred GitCred) {
	g.GitCred = gitCred
}

func (g *GitRepoSchema) GetGitRepo() string {
	return g.GitRepo
}

func (g *GitRepoSchema) GetPathToRoot() string {
	return g.PathToRoot
}

func (g *GitRepoSchema) GetGitCred() GitCred {
	return g.GitCred
}

func UnmarshallGitRepoSchema(m map[string]interface{}) GitRepoSchema {
	gitCred := UnmarshallGitCred(m["gitCred"].(map[string]interface{}))
	return GitRepoSchema{
		GitRepo:    m["gitRepo"].(string),
		PathToRoot: m["pathToRoot"].(string),
		GitCred:    gitCred,
	}
}

func UnmarshallGitRepoSchemaString(str string) GitRepoSchema {
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(str), &m)
	return UnmarshallGitRepoSchema(m)
}

func MarshalGitRepoSchema(schema GitRepoSchema) map[string]interface{} {
	bytes, err := json.Marshal(schema)
	if err != nil {
		panic(err)
	}
	var mapObj map[string]interface{}
	err = json.Unmarshal(bytes, &mapObj)
	if err != nil {
		panic(err)
	}
	mapObj["gitCred"] = MarshalGitCred(schema.GetGitCred())
	return mapObj
}