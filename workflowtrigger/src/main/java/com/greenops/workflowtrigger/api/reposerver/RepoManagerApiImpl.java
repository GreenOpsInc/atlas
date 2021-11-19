package com.greenops.workflowtrigger.api.reposerver;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.git.GitCredMachineUser;
import com.greenops.util.datamodel.git.GitCredToken;
import com.greenops.util.datamodel.git.GitRepoSchema;
import com.greenops.util.datamodel.mixin.git.GitCredMachineUserMixin;
import com.greenops.util.datamodel.mixin.git.GitCredTokenMixin;
import com.greenops.util.datamodel.mixin.git.GitRepoSchemaMixin;
import com.greenops.util.datamodel.request.GetFileRequest;
import com.greenops.util.error.AtlasNonRetryableError;
import com.greenops.util.error.AtlasRetryableError;
import lombok.extern.slf4j.Slf4j;
import org.apache.http.client.HttpClient;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.HttpClientBuilder;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.io.IOException;

import static com.greenops.util.api.ApiClientUtil.checkResponseStatus;

@Slf4j
@Component
public class RepoManagerApiImpl implements RepoManagerApi {

    public static final String ROOT_COMMIT = "ROOT_COMMIT";
    private static final String GET_FILE_EXTENSION = "file";
    public static final String PIPELINE_FILE_NAME = "pipeline.yaml";

    private final String serverEndpoint;
    private final ObjectMapper objectMapper;
    private final HttpClient httpClient;

    public RepoManagerApiImpl(@Value("${application.repo-server-url}") String serverEndpoint) {
        this.serverEndpoint = serverEndpoint.endsWith("/") ? serverEndpoint + "repo" : serverEndpoint + "/repo";
        httpClient = HttpClientBuilder.create().build();
        objectMapper = new ObjectMapper()
                .addMixIn(GitRepoSchema.class, GitRepoSchemaMixin.class)
                .addMixIn(GitCredMachineUser.class, GitCredMachineUserMixin.class)
                .addMixIn(GitCredToken.class, GitCredTokenMixin.class);
    }

    @Override
    public boolean cloneRepo(String orgName, GitRepoSchema gitRepoSchema) {
        try {
            var requestBody = objectMapper.writeValueAsString(gitRepoSchema);
            var request = new HttpPost(serverEndpoint + "/clone/" + orgName);
            request.setEntity(new StringEntity(requestBody, ContentType.APPLICATION_JSON));
            var response = httpClient.execute(request);
            log.info("Clone request for repo {} returned with status code {}", gitRepoSchema.getGitRepo(), response.getStatusLine().getStatusCode());
            return response.getStatusLine().getStatusCode() == 200;
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert Git repo schema for repo: {}", gitRepoSchema.getGitRepo(), e);
            return false;
        } catch (IOException e) {
            log.error("HTTP request failed for repo: {}", gitRepoSchema.getGitRepo(), e);
            return false;
        }
        //TODO: Catch branches left separate for future processing, logic, and logging.
    }

    @Override
    public boolean deleteRepo(GitRepoSchema gitRepoSchema) {
        try {
            var requestBody = objectMapper.writeValueAsString(gitRepoSchema);
            var request = new HttpPost(serverEndpoint + "/delete");
            request.setEntity(new StringEntity(requestBody, ContentType.APPLICATION_JSON));
            var response = httpClient.execute(request);
            log.info("Delete folder request for repo {} returned with status code {}", gitRepoSchema.getGitRepo(), response.getStatusLine().getStatusCode());
            return response.getStatusLine().getStatusCode() == 200;
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert Git repo schema for repo: {}", gitRepoSchema.getGitRepo(), e);
            return false;
        } catch (IOException e) {
            log.error("HTTP request failed for repo: {}", gitRepoSchema.getGitRepo(), e);
            return false;
        }
        //TODO: Catch branches left separate for future processing, logic, and logging.
    }

    @Override
    public boolean sync(GitRepoSchema gitRepoSchema) {
        try {
            var requestBody = objectMapper.writeValueAsString(gitRepoSchema);
            var request = new HttpPost(serverEndpoint + "/sync");
            request.setEntity(new StringEntity(requestBody, ContentType.APPLICATION_JSON));
            var response = httpClient.execute(request);
            log.info("sync folder request for repo {} returned with status code {}", gitRepoSchema.getGitRepo(), response.getStatusLine().getStatusCode());
            return response.getStatusLine().getStatusCode() == 200;
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert Git repo schema for repo: {}", gitRepoSchema.getGitRepo(), e);
            return false;
        } catch (IOException e) {
            log.error("HTTP request failed for repo: {}", gitRepoSchema.getGitRepo(), e);
            return false;
        }
        //TODO: Catch branches left separate for future processing, logic, and logging.
    }

    @Override
    public String getFileFromRepo(GetFileRequest getFileRequest, String orgName, String teamName) {
        var request = new HttpPost(serverEndpoint + String.format("/%s/%s/%s", GET_FILE_EXTENSION, orgName, teamName));
        try {
            var requestBody = objectMapper.writeValueAsString(getFileRequest);
            request.setEntity(new StringEntity(requestBody, ContentType.APPLICATION_JSON));
            var response = httpClient.execute(request);
            log.info("Fetch file request for repo {} returned with status code {}", getFileRequest.getGitRepoUrl(), response.getStatusLine().getStatusCode());
            checkResponseStatus(response);
            return new String(response.getEntity().getContent().readAllBytes());
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert GetFileRequest", e);
            throw new AtlasNonRetryableError(e);
        } catch (IOException e) {
            log.error("HTTP get file request failed for repo: {}", getFileRequest.getGitRepoUrl(), e);
            throw new AtlasRetryableError(e);
        } finally {
            request.releaseConnection();
        }
    }
}
