package db

type StepMetadata struct {
	ArgoRepoSchema *ArgoRepoSchema `json:"argoRepoSchema"`
}

type ArgoRepoSchema struct {
	RepoURL        string `json:"repoURL"`
	TargetRevision string `json:"targetRevision"`
	Path           string `json:"path"`
}

func NewArgoRepoSchema(repoURL, targetRevision, path string) *ArgoRepoSchema {
	if targetRevision == "" {
		targetRevision = "main"
	}
	if path == "" {
		path = "/"
	}
	return &ArgoRepoSchema{
		RepoURL:        repoURL,
		TargetRevision: targetRevision,
		Path:           path,
	}
}
