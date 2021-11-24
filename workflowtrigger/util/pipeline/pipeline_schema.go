package pipeline

import (
	"encoding/json"
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

func UnmarshallPipelineSchemaString(str string) PipelineSchema {
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(str), &m)
	return UnmarshallPipelineSchema(m)
}

func MarshalPipelineSchema(schema PipelineSchema) map[string]interface{} {
	bytes, err := json.Marshal(schema)
	if err != nil {
		panic(err)
	}
	var mapObj map[string]interface{}
	err = json.Unmarshal(bytes, &mapObj)
	if err != nil {
		panic(err)
	}
	mapObj["gitRepoSchema"] = git.MarshalGitRepoSchema(schema.GetGitRepoSchema())
	return mapObj
}
