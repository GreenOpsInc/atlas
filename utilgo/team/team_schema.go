package team

import (
	"encoding/json"

	"gitlab.com/c0b/go-ordered-json"
	"github.com/greenopsinc/util/git"
	"github.com/greenopsinc/util/pipeline"
)

type TeamSchema struct {
	TeamName       string                     `json:"teamName"`
	ParentTeamName string                     `json:"parentTeam"`
	OrgName        string                     `json:"orgName"`
	Pipelines      []*pipeline.PipelineSchema `json:"pipelines"`
}

func New(teamName string, parentTeamName string, orgName string) *TeamSchema {
	return &TeamSchema{
		TeamName:       teamName,
		ParentTeamName: parentTeamName,
		OrgName:        orgName,
		Pipelines:      []*pipeline.PipelineSchema{},
	}
}

func (p *TeamSchema) GetTeamName() string {
	return p.TeamName
}

func (p *TeamSchema) GetParentTeamName() string {
	return p.ParentTeamName
}

func (p *TeamSchema) GetOrgName() string {
	return p.OrgName
}

func (p *TeamSchema) SetTeamName(teamName string) {
	p.TeamName = teamName
}

func (p *TeamSchema) SetParentTeamName(parentTeamName string) {
	p.ParentTeamName = parentTeamName
}

func (p *TeamSchema) AddPipeline(pipelineName string, schema git.GitRepoSchema) {
	p.Pipelines = append(p.Pipelines, pipeline.New(pipelineName, schema))
}

func (p *TeamSchema) RemovePipeline(pipelineName string) {
	for idx, val := range p.Pipelines {
		if val.GetPipelineName() == pipelineName {
			p.Pipelines = append(p.Pipelines[:idx], p.Pipelines[idx+1:]...)
		}
	}
}

func (p *TeamSchema) GetPipelineNames() []string {
	pipelineNames := make([]string, 0)
	for _, val := range p.Pipelines {
		pipelineNames = append(pipelineNames, val.GetPipelineName())
	}
	return pipelineNames
}

func (p *TeamSchema) GetPipelineSchemas() []*pipeline.PipelineSchema {
	return p.Pipelines
}

func (p *TeamSchema) GetPipelineSchema(pipelineName string) *pipeline.PipelineSchema {
	for _, val := range p.Pipelines {
		if val.GetPipelineName() == pipelineName {
			return val
		}
	}
	return nil
}

func (p *TeamSchema) UpdatePipeline(pipelineName string, schema git.GitRepoSchema) {
	for idx, val := range p.Pipelines {
		if val.GetPipelineName() == pipelineName {
			p.Pipelines[idx] = pipeline.New(pipelineName, schema)
		}
	}
}

func UnmarshallTeamSchema(m map[string]interface{}) TeamSchema {
	var pipelineList []*pipeline.PipelineSchema
	if pipelineStringList, ok := m["pipelines"]; ok && pipelineStringList != nil {
		for _, val := range pipelineStringList.([]interface{}) {
			unmarshalledPipeline := pipeline.UnmarshallPipelineSchema(val.(map[string]interface{}))
			pipelineList = append(pipelineList, &unmarshalledPipeline)
		}
	}
	return TeamSchema{
		TeamName:       m["teamName"].(string),
		ParentTeamName: m["parentTeam"].(string),
		OrgName:        m["orgName"].(string),
		Pipelines:      pipelineList,
	}
}

func UnmarshallTeamSchemaString(str string) TeamSchema {
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(str), &m)
	return UnmarshallTeamSchema(m)
}

func MarshalTeamSchema(schema TeamSchema) *ordered.OrderedMap {
	mapObj := ordered.NewOrderedMap()
	mapObj.Set("teamName", schema.TeamName)
	mapObj.Set("parentTeam", schema.ParentTeamName)
	mapObj.Set("orgName", schema.OrgName)

	var pipelineInterfaceList []*ordered.OrderedMap
	for _, val := range schema.GetPipelineSchemas() {
		pipelineInterfaceList = append(pipelineInterfaceList, pipeline.MarshalPipelineSchema(*val))
	}
	mapObj.Set("pipelines", pipelineInterfaceList)

	return mapObj
}
