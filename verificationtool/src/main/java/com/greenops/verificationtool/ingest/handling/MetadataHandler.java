package com.greenops.verificationtool.ingest.handling;

import com.greenops.util.datamodel.event.Event;
import com.greenops.util.datamodel.git.ArgoRepoSchema;
import com.greenops.verificationtool.datamodel.pipelinedata.PipelineData;

import java.util.List;

public interface MetadataHandler {

    ArgoRepoSchema getArgoSourceRepoMetadata(String argoAppPayload);

    ArgoRepoSchema getCurrentArgoRepoMetadata(Event event, String stepName);

    void assertArgoRepoMetadataExists(Event event, String currentStepName, String argoConfig);

    String getPipelineLockRevisionHash(Event event, PipelineData pipelineData, String currentStepName);

    List<String> findAllStepsWithSameArgoRepoSrc(Event event, PipelineData pipelineData, String currentStepName);
}
