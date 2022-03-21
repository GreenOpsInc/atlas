package handling

import (
	"errors"
	"log"

	"github.com/greenopsinc/util/array"
	"github.com/greenopsinc/util/db"
	"github.com/greenopsinc/util/event"
	"github.com/greenopsinc/util/pipeline/data"
	"github.com/greenopsinc/workfloworchestrator/ingest/dbkey"
	"gopkg.in/yaml.v2"
)

type MetadataHandler interface {
	GetArgoSourceRepoMetadata(argoAppPayload string) (*db.ArgoRepoSchema, error)
	GetCurrentArgoRepoMetadata(event event.Event, stepName string) *db.ArgoRepoSchema
	AssertArgoRepoMetadataExists(event event.Event, currentStepName string, argoConfig string) error
	GetPipelineLockRevisionHash(event event.Event, pipelineData *data.PipelineData, currentStepName string) string
	FindAllStepsWithSameArgoRepoSrc(event event.Event, pipelineData *data.PipelineData, currentStepName string) []string
}

type metadataHandler struct {
	dbClient db.DbClient
}

func NewMetadataHandler(dbClient db.DbClient) MetadataHandler {
	return &metadataHandler{dbClient: dbClient}
}

func (m *metadataHandler) GetArgoSourceRepoMetadata(argoAppPayload string) (*db.ArgoRepoSchema, error) {
	var schema map[interface{}]interface{}
	if err := yaml.Unmarshal([]byte(argoAppPayload), schema); err != nil {
		return nil, err
	}
	spec, ok := schema["spec"].(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("cannot unmarshal argoAppPayload yaml file contents: field spec is not a map")
	}
	source, ok := spec["source"].(map[interface{}]interface{})
	if !ok {
		return nil, errors.New("cannot unmarshal argoAppPayload yaml file contents: field spec.source is not a map")
	}

	repoURL, ok := source["repoURL"].(string)
	if !ok {
		repoURL = ""
	}
	targetRevision, ok := source["targetRevision"].(string)
	if !ok {
		targetRevision = "main"
	}
	path, ok := source["path"].(string)
	if !ok {
		path = ""
	}

	return &db.ArgoRepoSchema{
		RepoURL:        repoURL,
		TargetRevision: targetRevision,
		Path:           path,
	}, nil
}

func (m *metadataHandler) GetCurrentArgoRepoMetadata(event event.Event, stepName string) *db.ArgoRepoSchema {
	key := dbkey.MakeDbMetadataKey(event.GetOrgName(), event.GetTeamName(), event.GetPipelineName(), stepName)
	metadata := m.dbClient.FetchMetadata(key)
	if metadata != nil {
		return metadata.ArgoRepoSchema
	}
	return nil
}

func (m *metadataHandler) AssertArgoRepoMetadataExists(event event.Event, currentStepName string, argoConfig string) error {
	key := dbkey.MakeDbMetadataKey(event.GetOrgName(), event.GetTeamName(), event.GetPipelineName(), currentStepName)
	repoSchema, err := m.GetArgoSourceRepoMetadata(argoConfig)
	if err != nil {
		return err
	}

	currMetadata := m.getCurrentMetadata(event, currentStepName)
	if currMetadata == nil {
		currMetadata = &db.StepMetadata{}
	}

	if !argoRepoSchemasEqual(currMetadata.ArgoRepoSchema, repoSchema) {
		currMetadata.ArgoRepoSchema = repoSchema
		m.dbClient.StoreValue(key, currMetadata)
	}
	return nil
}

func (m *metadataHandler) GetPipelineLockRevisionHash(event event.Event, pipelineData *data.PipelineData, currentStepName string) string {
	log.Println("get Argo revision for pipeline locking")

	repoSchema := m.GetCurrentArgoRepoMetadata(event, currentStepName)
	precedingSteps := m.findAllPrecedingSteps(pipelineData, currentStepName)

	for _, stepName := range precedingSteps {
		dependentArgoRepoSchema := m.GetCurrentArgoRepoMetadata(event, stepName)
		if argoRepoSchemasEqual(dependentArgoRepoSchema, repoSchema) {
			logKey := dbkey.MakeDbStepKey(event.GetOrgName(), event.GetTeamName(), event.GetPipelineName(), stepName)
			deploymentLog := m.dbClient.FetchLatestDeploymentLog(logKey)
			return deploymentLog.GetArgoRevisionHash()
		}
	}
	return ""
}

func (m *metadataHandler) FindAllStepsWithSameArgoRepoSrc(event event.Event, pipelineData *data.PipelineData, currentStepName string) []string {
	repoSchema := m.GetCurrentArgoRepoMetadata(event, currentStepName)
	res := make([]string, 0)

	for _, step := range pipelineData.GetAllSteps() {
		dependentArgoRepoSchema := m.GetCurrentArgoRepoMetadata(event, step)
		if argoRepoSchemasEqual(dependentArgoRepoSchema, repoSchema) {
			res = append(res, step)
		}
	}
	return res
}

func (m *metadataHandler) getCurrentMetadata(event event.Event, stepName string) *db.StepMetadata {
	key := dbkey.MakeDbMetadataKey(event.GetOrgName(), event.GetTeamName(), event.GetPipelineName(), stepName)
	return m.dbClient.FetchMetadata(key)
}

func (m *metadataHandler) findAllPrecedingSteps(pipelineData *data.PipelineData, currentStepName string) []string {
	levelMarker := "|"
	stepsToReturn := make([]string, 0)
	stepsInScope := pipelineData.StepChildren[data.RootStepName]

	for _, step := range stepsInScope {
		if currentStepName == step {
			return stepsToReturn
		}
	}
	stepsInScope = append(stepsInScope, levelMarker)

	for len(stepsInScope) > 0 {
		isEndOfLevel := false
		currentTraversalStep := stepsInScope[0]
		stepsInScope = stepsInScope[1:]
		if currentTraversalStep == levelMarker {
			continue
		}

		stepsToReturn = append(stepsToReturn, currentTraversalStep)
		if len(stepsInScope) > 0 && stepsInScope[0] == levelMarker {
			stepsToReturn = append(stepsToReturn, stepsToReturn[0])
			stepsToReturn = stepsToReturn[1:]
			isEndOfLevel = true
		}

		childrenSteps := pipelineData.StepChildren[currentTraversalStep]
		for _, s := range childrenSteps {
			if s == currentStepName {
				lastLevelMarkerIdx := array.LastIndexOf(stepsToReturn, levelMarker)
				if lastLevelMarkerIdx == 0 || lastLevelMarkerIdx == -1 {
					return []string{}
				}

				sublist := stepsToReturn[0:lastLevelMarkerIdx]
				var res []string
				for _, el := range sublist {
					if el != levelMarker {
						res = append(res, el)
					}
				}
				return res
			}
		}

		for _, el := range childrenSteps {
			stepsInScope = append(stepsInScope, el)
			if isEndOfLevel {
				stepsInScope = append(stepsInScope, levelMarker)
			}
		}
	}
	return []string{}
}

func argoRepoSchemasEqual(sc1, sc2 *db.ArgoRepoSchema) bool {
	if sc1 == nil || sc2 == nil {
		return false
	}
	if sc1.Path != sc2.Path {
		return false
	}
	if sc1.RepoURL != sc2.RepoURL {
		return false
	}
	if sc1.TargetRevision != sc2.TargetRevision {
		return false
	}
	return true
}
