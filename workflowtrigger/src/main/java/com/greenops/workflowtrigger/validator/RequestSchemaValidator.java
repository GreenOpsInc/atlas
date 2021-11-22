package com.greenops.workflowtrigger.validator;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.git.ArgoRepoSchema;
import com.greenops.util.datamodel.request.GetFileRequest;
import com.greenops.util.error.AtlasAuthenticationError;
import com.greenops.util.error.AtlasNonRetryableError;
import com.greenops.workflowtrigger.api.argoauthenticator.ArgoAuthenticatorApi;
import com.greenops.workflowtrigger.api.reposerver.RepoManagerApi;
import com.greenops.workflowtrigger.datamodel.pipelinedata.PipelineData;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.stereotype.Component;

import java.util.Map;

import static com.greenops.workflowtrigger.api.argoauthenticator.ArgoAuthenticatorApi.*;
import static com.greenops.workflowtrigger.api.reposerver.RepoManagerApiImpl.PIPELINE_FILE_NAME;

@Slf4j
@Component
public class RequestSchemaValidator {

    private static final String DEFAULT_PROJECT = "default";

    private ObjectMapper yamlObjectMapper;
    private ObjectMapper objectMapper;
    private RepoManagerApi repoManagerApi;
    private ArgoAuthenticatorApi argoAuthenticatorApi;

    @Autowired
    public RequestSchemaValidator(@Qualifier("yamlObjectMapper") ObjectMapper yamlObjectMapper,
                                  @Qualifier("objectMapper") ObjectMapper objectMapper,
                                  RepoManagerApi repoManagerApi,
                                  ArgoAuthenticatorApi argoAuthenticatorApi) {
        this.yamlObjectMapper = yamlObjectMapper;
        this.objectMapper = objectMapper;
        this.repoManagerApi = repoManagerApi;
        this.argoAuthenticatorApi = argoAuthenticatorApi;
    }

    //actionResourceEntries expected to be an even number. Should be pairs of actions and resources.
    public boolean validateSchemaAccess(String orgName, String teamName, String gitRepoUrl, String gitCommitHash, String...actionResourceEntries) {

        var pipelineData = getPipelineData(orgName, teamName, gitRepoUrl, gitCommitHash);
        for (var stepName : pipelineData.getAllSteps()) {
            var step = pipelineData.getStep(stepName);
            if (step.getClusterName() == null) {
                throw new AtlasNonRetryableError("The pipeline schema is invalid. Step is missing a cluster.");
            }

            for (int i = 0; i < actionResourceEntries.length; i += 2) {
                var action = actionResourceEntries[i];
                var resource = actionResourceEntries[i + 1];
                if (resource.equals(CLUSTER_RESOURCE)) {
                    if (!argoAuthenticatorApi.checkRbacPermissions(action, CLUSTER_RESOURCE, step.getClusterName())) {
                        return false;
                    }
                } else if (resource.equals(APPLICATION_RESOURCE)) {
                    var applicationSubresource = getArgoApplicationProjectAndName(orgName, teamName, gitRepoUrl, gitCommitHash, step.getArgoApplicationPath());
                    if (!argoAuthenticatorApi.checkRbacPermissions(action, CLUSTER_RESOURCE, applicationSubresource)) {
                        return false;
                    }
                }
            }
        }
        return true;
    }

    public boolean checkAuthentication() {
        try {
            //Dummy request to check for authentication
            argoAuthenticatorApi.checkRbacPermissions(SYNC_ACTION, CLUSTER_RESOURCE, "abc");
            return true;
        } catch (AtlasAuthenticationError e) {
            return false;
        }
    }

    public boolean verifyRbac(String action, String resource, String subresource) {
        return argoAuthenticatorApi.checkRbacPermissions(action, resource, subresource);
    }

    private PipelineData getPipelineData(String orgName, String teamName, String gitRepoUrl, String gitCommitHash) {
        var getFileRequest = new GetFileRequest(gitRepoUrl, PIPELINE_FILE_NAME, gitCommitHash);
        try {
            return objectMapper.readValue(
                    objectMapper.writeValueAsString(
                            yamlObjectMapper.readValue(repoManagerApi.getFileFromRepo(getFileRequest, orgName, teamName), Object.class)
                    ),
                    PipelineData.class);
        } catch (JsonProcessingException e) {
            log.error("Could not parse YAML pipeline data file", e);
            throw new AtlasNonRetryableError(e);
        }
    }

    private String getArgoApplicationProjectAndName(String orgName, String teamName, String gitRepoUrl, String gitCommitHash, String applicationPath) {
        var getFileRequest = new GetFileRequest(gitRepoUrl, applicationPath, gitCommitHash);
        var argoApplicationConfig = repoManagerApi.getFileFromRepo(getFileRequest, orgName, teamName);
        try {
            var project = objectMapper.readTree(argoApplicationConfig).get("spec").get("project");
            var name = objectMapper.readTree(argoApplicationConfig).get("metadata").get("name");
            if (name.isNull()) {
                throw new AtlasNonRetryableError("The Argo CD app does not have a name");
            }
            if (project.isNull()) {
                return DEFAULT_PROJECT + "/" + name.asText();
            } else {
                return project.asText() + "/" + name.asText();
            }
        } catch (JsonProcessingException e) {
            throw new AtlasNonRetryableError("Argo app configuration cannot be parsed");
        }
    }

}
