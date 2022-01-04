package client

type ClientName string

const (
	ClientRepoServer       ClientName = "reposerver"
	ClientCommandDelegator ClientName = "commanddelegator"
	ClientArgoCDRepoServer ClientName = "argocdreposerver"
)
