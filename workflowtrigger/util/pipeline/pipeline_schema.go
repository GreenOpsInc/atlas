package pipeline

import (
	"gitlab.com/c0b/go-ordered-json"
	"greenops.io/workflowtrigger/util/git"
)

type PipelineSchema struct {
	PipelineName  string            `json:"pipelineName"`
	GitRepoSchema git.GitRepoSchema `json:"gitRepoSchema"`
}

func New(pipelineName string, gitRepoSchema git.GitRepoSchema) *PipelineSchema {
	return &PipelineSchema{
		PipelineName:  pipelineName,
		GitRepoSchema: gitRepoSchema,
	}
}

func (p *PipelineSchema) GetPipelineName() string {
	return p.PipelineName
}

func (p *PipelineSchema) GetGitRepoSchema() git.GitRepoSchema {
	return p.GitRepoSchema
}

func (p *PipelineSchema) SetPipelineName(pipelineName string) {
	p.PipelineName = pipelineName
}

func (p *PipelineSchema) SetGitRepoSchema(gitRepoSchema git.GitRepoSchema) {
	p.GitRepoSchema = gitRepoSchema
}

func UnmarshallPipelineSchema(m map[string]interface{}) PipelineSchema {
	gitRepoSchema := git.UnmarshallGitRepoSchema(m["gitRepoSchema"].(map[string]interface{}))
	return PipelineSchema{
		PipelineName:  m["pipelineName"].(string),
		GitRepoSchema: gitRepoSchema,
	}
}

func MarshalPipelineSchema(schema PipelineSchema) *ordered.OrderedMap {
	mapObj := ordered.NewOrderedMap()
	mapObj.Set("pipelineName", schema.PipelineName)

	repo := git.MarshalGitRepoSchema(schema.GetGitRepoSchema())
	mapObj.Set("gitRepoSchema", repo)

	return mapObj
}
