package com.greenops.workfloworchestrator.ingest.handling;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.git.ArgoRepoSchema;
import com.greenops.util.datamodel.metadata.StepMetadata;
import com.greenops.util.datamodel.pipelinedata.PipelineData;
import com.greenops.util.dbclient.DbClient;
import com.greenops.util.error.AtlasNonRetryableError;
import com.greenops.workfloworchestrator.ingest.dbclient.DbKey;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

import static com.greenops.util.datamodel.pipelinedata.StepData.ROOT_STEP_NAME;

@Slf4j
@Component
public class MetadataHandlerImpl implements MetadataHandler {

    private DbClient dbClient;
    private ObjectMapper objectMapper;

    @Autowired
    MetadataHandlerImpl(DbClient dbClient, @Qualifier("yamlObjectMapper") ObjectMapper objectMapper) {
        this.dbClient = dbClient;
        this.objectMapper = objectMapper;
    }

    @Override
    public ArgoRepoSchema getArgoSourceRepoMetadata(String argoAppPayload) {
        try {
            var sourceJsonNode = objectMapper.readTree(argoAppPayload).get("spec").get("source");
            return new ArgoRepoSchema(
                    sourceJsonNode.get("repoURL").asText(null),
                    sourceJsonNode.get("targetRevision").asText(null),
                    sourceJsonNode.get("path").asText(null)
            );
        } catch (JsonProcessingException e) {
            throw new AtlasNonRetryableError("Argo app configuration cannot be parsed");
        }
    }

    private StepMetadata getCurrentMetadata(Event event, String stepName) {
        var key = DbKey.makeDbMetadataKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        return dbClient.fetchMetadata(key);
    }

    @Override
    public ArgoRepoSchema getCurrentArgoRepoMetadata(Event event, String stepName) {
        var key = DbKey.makeDbMetadataKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
        var metadata = dbClient.fetchMetadata(key);
        if (metadata != null) {
            return metadata.getArgoRepoSchema();
        }
        return null;
    }

    @Override
    public void assertArgoRepoMetadataExists(Event event, String currentStepName, String argoConfig) {
        var key = DbKey.makeDbMetadataKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), currentStepName);
        var argoRepoSchema = getArgoSourceRepoMetadata(argoConfig);
        var currentMetadata = getCurrentMetadata(event, currentStepName);
        if (currentMetadata == null) {
            currentMetadata = new StepMetadata(null);
        }
        if (currentMetadata.getArgoRepoSchema() == null || !currentMetadata.getArgoRepoSchema().equals(argoRepoSchema)) {
            currentMetadata.setArgoRepoSchema(argoRepoSchema);
            dbClient.storeValue(key, currentMetadata);
        }
    }

    @Override
    public String getPipelineLockRevisionHash(Event event, PipelineData pipelineData, String currentStepName) {
        log.info("Get Argo revision for pipeline locking");
        var argoRepoSchema = getCurrentArgoRepoMetadata(event, currentStepName);
        var precedingSteps = findAllPrecedingSteps(pipelineData, currentStepName);
        for (var stepName : precedingSteps) {
            var dependentArgoRepoSchema = getCurrentArgoRepoMetadata(event, stepName);
            if (dependentArgoRepoSchema.equals(argoRepoSchema)) {
                var logKey = DbKey.makeDbStepKey(event.getOrgName(), event.getTeamName(), event.getPipelineName(), stepName);
                var deploymentLog = dbClient.fetchLatestDeploymentLog(logKey);
                return deploymentLog.getArgoRevisionHash();
            }
        }
        return null;
    }

    public List<String> findAllStepsWithSameArgoRepoSrc(Event event, PipelineData pipelineData, String currentStepName) {
        var argoRepoSchema = getCurrentArgoRepoMetadata(event, currentStepName);
        return pipelineData.getAllSteps().stream().filter(stepName -> {
            var dependentArgoRepoSchema = getCurrentArgoRepoMetadata(event, stepName);
            return dependentArgoRepoSchema.equals(argoRepoSchema);
        }).collect(Collectors.toList());
    }

    private List<String> findAllPrecedingSteps(PipelineData pipelineData, String currentStepName) {
        var levelMarker = "|";
        var stepsToReturn = new ArrayList<String>();
        var stepsInScope = pipelineData.getChildrenSteps(ROOT_STEP_NAME);
        if (stepsInScope.contains(currentStepName)) return stepsToReturn;
        stepsInScope.add(levelMarker);
        while (stepsInScope.size() > 0) {
            var currentTraversalStep = stepsInScope.remove(0);
            if (currentTraversalStep.equals(levelMarker)) {
                continue;
            }
            stepsToReturn.add(currentTraversalStep);
            if (stepsInScope.size() > 0 && stepsInScope.get(0).equals(levelMarker)) {
                stepsToReturn.add(stepsInScope.remove(0));
            }

            var childrenSteps = pipelineData.getChildrenSteps(currentTraversalStep);
            if (childrenSteps.contains(currentStepName)) {
                var lastLevelMarkerIdx = stepsToReturn.lastIndexOf(levelMarker);
                if (lastLevelMarkerIdx == 0 || lastLevelMarkerIdx == -1) {
                    return List.of();
                }
                return stepsToReturn.subList(0, lastLevelMarkerIdx).stream().filter(str -> !str.equals(levelMarker)).collect(Collectors.toList());
            } else {
                stepsInScope.addAll(childrenSteps);
            }
        }
        //Could not find preceding step. Should never be happening
        return List.of();
    }
}
