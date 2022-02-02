package com.greenops.verificationtool.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.greenops.util.datamodel.auditlog.DeploymentLog;
import com.greenops.util.datamodel.auditlog.PipelineInfo;
import com.greenops.util.datamodel.auditlog.RemediationLog;
import com.greenops.util.datamodel.clientmessages.*;
import com.greenops.util.datamodel.event.*;
import com.greenops.util.datamodel.git.*;
import com.greenops.util.datamodel.metadata.StepMetadata;
import com.greenops.util.datamodel.mixin.auditlog.DeploymentLogMixin;
import com.greenops.util.datamodel.mixin.auditlog.PipelineInfoMixin;
import com.greenops.util.datamodel.mixin.auditlog.RemediationLogMixin;
import com.greenops.util.datamodel.mixin.clientmessages.*;
import com.greenops.util.datamodel.mixin.event.*;
import com.greenops.util.datamodel.mixin.git.*;
import com.greenops.util.datamodel.mixin.metadata.StepMetadataMixin;
import com.greenops.util.datamodel.mixin.pipeline.PipelineSchemaMixin;
import com.greenops.util.datamodel.mixin.pipeline.TeamSchemaMixin;
import com.greenops.util.datamodel.mixin.pipelinedata.*;
import com.greenops.util.datamodel.mixin.pipelinestatus.FailedStepMixin;
import com.greenops.util.datamodel.mixin.pipelinestatus.PipelineStatusMixin;
import com.greenops.util.datamodel.mixin.request.DeployResponseMixin;
import com.greenops.util.datamodel.mixin.request.GetFileRequestMixin;
import com.greenops.util.datamodel.mixin.request.GitRepoSchemaInfoMixin;
import com.greenops.util.datamodel.pipeline.PipelineSchemaImpl;
import com.greenops.util.datamodel.pipeline.TeamSchemaImpl;
import com.greenops.util.datamodel.pipelinedata.*;
import com.greenops.util.datamodel.pipelinestatus.FailedStep;
import com.greenops.util.datamodel.pipelinestatus.PipelineStatus;
import com.greenops.util.datamodel.request.DeployResponse;
import com.greenops.util.datamodel.request.GetFileRequest;
import com.greenops.verificationtool.datamodel.mixin.requests.VerifyPipelineRequestBodyMixin;
import com.greenops.verificationtool.datamodel.mixin.rules.RuleDataMixin;
import com.greenops.verificationtool.datamodel.requests.VerifyPipelineRequestBody;
import com.greenops.verificationtool.datamodel.rules.RuleDataImpl;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Slf4j
@Configuration
public class SpringConfiguration {

    @Bean
    @Qualifier("yamlObjectMapper")
    ObjectMapper yamlObjectMapper() {
        return new ObjectMapper(new YAMLFactory());
    }

    @Bean
    @Qualifier("eventAndRequestObjectMapper")
    ObjectMapper eventAndRequestObjectMapper() {
        return new ObjectMapper()
                .addMixIn(ClientCompletionEvent.class, ClientCompletionEventMixin.class)
                .addMixIn(PipelineCompletionEvent.class, PipelineCompletionEventMixin.class)
                .addMixIn(ResourceStatus.class, ResourceStatusMixin.class)
                .addMixIn(TestCompletionEvent.class, TestCompletionEventMixin.class)
                .addMixIn(FailureEvent.class, FailureEventMixin.class)
                .addMixIn(ApplicationInfraTriggerEvent.class, ApplicationInfraTriggerEventMixin.class)
                .addMixIn(ApplicationInfraCompletionEvent.class, ApplicationInfraCompletionEventMixin.class)
                .addMixIn(ClientRequestPacket.class, ClientRequestPacketMixin.class)
                .addMixIn(ClientDeleteByConfigRequest.class, ClientDeleteByConfigRequestMixin.class)
                .addMixIn(ClientDeleteByGvkRequest.class, ClientDeleteByGvkRequestMixin.class)
                .addMixIn(ClientDeployAndWatchRequest.class, ClientDeployAndWatchRequestMixin.class)
                .addMixIn(ClientDeployNamedArgoAppAndWatchRequest.class, ClientDeployNamedArgoAppAndWatchRequestMixin.class)
                .addMixIn(ClientDeployNamedArgoApplicationRequest.class, ClientDeployNamedArgoApplicationRequestMixin.class)
                .addMixIn(ClientDeployRequest.class, ClientDeployRequestMixin.class)
                .addMixIn(ClientRollbackAndWatchRequest.class, ClientRollbackAndWatchRequestMixin.class)
                .addMixIn(GetFileRequest.class, GetFileRequestMixin.class)
                .addMixIn(GitRepoSchemaInfo.class, GitRepoSchemaInfoMixin.class)
                .addMixIn(ResourcesGvkRequest.class, ResourcesGvkRequestMixin.class)
                .addMixIn(ResourceGvk.class, ResourceGvkMixin.class)
                .addMixIn(DeployResponse.class, DeployResponseMixin.class)
                .addMixIn(TriggerStepEvent.class, TriggerStepEventMixin.class)
                .addMixIn(PipelineTriggerEvent.class, PipelineTriggerEventMixin.class);
    }

    @Bean
    @Qualifier("objectMapper")
    ObjectMapper objectMapper() {
        return new ObjectMapper()
                .addMixIn(PipelineDataImpl.class, PipelineDataMixin.class)
                .addMixIn(StepDataImpl.class, StepDataMixin.class)
                .addMixIn(ArgoWorkflowTask.class, ArgoWorkflowTaskMixin.class)
                .addMixIn(InjectScriptTest.class, InjectScriptTestMixin.class)
                .addMixIn(CustomJobTest.class, CustomJobTestMixin.class)
                .addMixIn(PipelineSchemaImpl.class, PipelineSchemaMixin.class)
                .addMixIn(TeamSchemaImpl.class, TeamSchemaMixin.class)
                .addMixIn(GitRepoSchema.class, GitRepoSchemaMixin.class)
                .addMixIn(GitCredMachineUser.class, GitCredMachineUserMixin.class)
                .addMixIn(GitCredToken.class, GitCredTokenMixin.class)
                .addMixIn(GitCredOpen.class, GitCredOpenMixin.class)
                .addMixIn(DeploymentLog.class, DeploymentLogMixin.class)
                .addMixIn(RemediationLog.class, RemediationLogMixin.class)
                .addMixIn(ResourceStatus.class, ResourceStatusMixin.class)
                .addMixIn(StepMetadata.class, StepMetadataMixin.class)
                .addMixIn(PipelineInfo.class, PipelineInfoMixin.class)
                .addMixIn(ArgoRepoSchema.class, ArgoRepoSchemaMixin.class)
                .addMixIn(ClientDeleteByConfigRequest.class, ClientDeleteByConfigRequestMixin.class)
                .addMixIn(ClientDeleteByGvkRequest.class, ClientDeleteByGvkRequestMixin.class)
                .addMixIn(ClientDeployAndWatchRequest.class, ClientDeployAndWatchRequestMixin.class)
                .addMixIn(ClientDeployRequest.class, ClientDeployRequestMixin.class)
                .addMixIn(ClientRollbackAndWatchRequest.class, ClientRollbackAndWatchRequestMixin.class)
                .addMixIn(ClientSelectiveSyncAndWatchRequest.class, ClientSelectiveSyncAndWatchRequestMixin.class)
                .addMixIn(ResourcesGvkRequest.class, ResourcesGvkRequestMixin.class)
                .addMixIn(ResourceGvk.class, ResourceGvkMixin.class)
                .addMixIn(FailedStep.class, FailedStepMixin.class)
                .addMixIn(PipelineStatus.class, PipelineStatusMixin.class)
                .addMixIn(VerifyPipelineRequestBody.class, VerifyPipelineRequestBodyMixin.class)
                .addMixIn(RuleDataImpl.class, RuleDataMixin.class);
    }
}
