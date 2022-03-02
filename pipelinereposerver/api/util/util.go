package util

import "github.com/greenopsinc/util/git"

func GetGitCredAccessibleFromGitCred(gitCred git.GitCred) git.GitCredAccessible {
	return gitCred.(git.GitCredAccessible)
}
