package git

import (
	"strings"

	"gitlab.com/c0b/go-ordered-json"
	"github.com/greenopsinc/util/serializerutil"
	"k8s.io/apimachinery/pkg/util/json"
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

func MarshalGitCred(gitCred GitCred) *ordered.OrderedMap {
	mapObj := ordered.NewOrderedMap()

	switch gitCred.(type) {
	case *GitCredMachineUser:
		mapObj.Set("username", gitCred.(*GitCredMachineUser).Username)
		mapObj.Set("password", gitCred.(*GitCredMachineUser).Password)
		mapObj.Set("type", serializerutil.GitCredMachineUserType)
	case *GitCredToken:
		mapObj.Set("token", gitCred.(*GitCredToken).Token)
		mapObj.Set("type", serializerutil.GitCredTokenType)
	default: //Open cred
		mapObj.Set("type", serializerutil.GitCredOpenType)
	}

	return mapObj
}
