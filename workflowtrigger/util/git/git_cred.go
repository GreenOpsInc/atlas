package git

import (
	"greenops.io/workflowtrigger/util/serializerutil"
	"k8s.io/apimachinery/pkg/util/json"
	"strings"
)

const (
	hiddenInfo         string = "Hidden cred info"
	secureGitUrlPrefix string = "https://"
)

type GitCred interface {
	Hide()
}

type GitCredAccessible interface {
	Hide()
	ConvertGitCredToString(gitRepoLink string) string
}

// GitCredOpen Setting up GitCredOpen
type GitCredOpen struct{}

func (g *GitCredOpen) Hide() {}

func (g *GitCredOpen) ConvertGitCredToString(gitRepoLink string) string {
	return gitRepoLink
}

// GitCredMachineUser Setting up GitCredMachineUser
type GitCredMachineUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (g *GitCredMachineUser) Hide() {
	g.Username = hiddenInfo
	g.Password = hiddenInfo
}

func (g *GitCredMachineUser) ConvertGitCredToString(gitRepoLink string) string {
	splitUrl := strings.Split(gitRepoLink, secureGitUrlPrefix)
	splitUrl[1] = g.Username + ":" + g.Password + "@" + splitUrl[1]
	return secureGitUrlPrefix + splitUrl[1]
}

// GitCredToken Setting up GitCredToken
type GitCredToken struct {
	Token string `json:"token"`
}

func (g *GitCredToken) Hide() {
	g.Token = hiddenInfo
}

func (g *GitCredToken) ConvertGitCredToString(gitRepoLink string) string {
	splitUrl := strings.Split(gitRepoLink, secureGitUrlPrefix)
	splitUrl[1] = g.Token + "@" + splitUrl[1]
	return secureGitUrlPrefix + splitUrl[1]
}

func UnmarshallGitCred(m map[string]interface{}) GitCred {
	credType := m["type"]
	delete(m, "type")
	gitCredBytes, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	switch credType {
	case serializerutil.GitCredMachineUserType:
		var gitCredMachineUser GitCredMachineUser
		_ = json.Unmarshal(gitCredBytes, &gitCredMachineUser)
		return &gitCredMachineUser
	case serializerutil.GitCredTokenType:
		var gitCredToken GitCredToken
		_ = json.Unmarshal(gitCredBytes, &gitCredToken)
		return &gitCredToken
	default:
		var gitCredOpen GitCredOpen
		_ = json.Unmarshal(gitCredBytes, &gitCredOpen)
		return &gitCredOpen
	}
}

func UnmarshallGitCredString(str string) GitCred {
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(str), &m)
	return UnmarshallGitCred(m)
}

func MarshalGitCred(gitCred GitCred) map[string]interface{} {
	switch gitCred.(type) {
	case *GitCredMachineUser:
		mapObj := serializerutil.GetMapFromStruct(gitCred)
		mapObj["type"] = serializerutil.GitCredMachineUserType
		return mapObj
	case *GitCredToken:
		mapObj := serializerutil.GetMapFromStruct(gitCred)
		mapObj["type"] = serializerutil.GitCredTokenType
		return mapObj
	default: //Open cred
		mapObj := serializerutil.GetMapFromStruct(gitCred)
		mapObj["type"] = serializerutil.GitCredOpenType
		return mapObj
	}
}
