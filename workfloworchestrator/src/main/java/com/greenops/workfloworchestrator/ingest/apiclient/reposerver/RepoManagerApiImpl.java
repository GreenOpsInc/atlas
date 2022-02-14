package com.greenops.workfloworchestrator.ingest.apiclient.reposerver;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.git.GitRepoSchemaInfo;
import com.greenops.util.datamodel.request.GetFileRequest;
import com.greenops.util.error.AtlasNonRetryableError;
import com.greenops.util.error.AtlasRetryableError;
import com.greenops.util.httpclient.Builder;
import com.greenops.util.kubernetesclient.KubernetesClient;
import lombok.extern.slf4j.Slf4j;
import org.apache.http.client.HttpClient;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.io.IOException;

import static com.greenops.workfloworchestrator.ingest.apiclient.util.ApiClientUtil.checkResponseStatus;

@Slf4j
@Component
public class RepoManagerApiImpl implements RepoManagerApi {

    private static final String ROOT_DATA_EXTENSION = "data";
    private static final String ROOT_REPO_EXTENSION = "repo";
    private static final String GET_FILE_EXTENSION = "file";
    private static final String GET_COMMIT_EXTENSION = "version";
    private static final String CHANGE_VERSION_EXTENSION = "resetToVersion";

    private final String serverRepoEndpoint;
    private final String serverDataEndpoint;
    private final ObjectMapper objectMapper;
    private final HttpClient httpClient;

    @Autowired
    public RepoManagerApiImpl(@Value("${application.repo-server-url}") String serverEndpoint, @Qualifier("eventAndRequestObjectMapper") ObjectMapper objectMapper, @Value("${application.repo-server-tls-secret-name}") String serverTlsSecretName, KubernetesClient kubernetesClient) {
        this.serverRepoEndpoint = serverEndpoint.endsWith("/") ? serverEndpoint + ROOT_REPO_EXTENSION : serverEndpoint + "/" + ROOT_REPO_EXTENSION;
        this.serverDataEndpoint = serverEndpoint.endsWith("/") ? serverEndpoint + ROOT_DATA_EXTENSION : serverEndpoint + "/" + ROOT_DATA_EXTENSION;
        this.objectMapper = objectMapper;
        try {
            this.httpClient = Builder.create(kubernetesClient).withCustomTls(serverTlsSecretName).build();
        }catch(Exception e) {
            log.error("Failed to create RepoServer HTTP client", e);
            throw new AtlasNonRetryableError(e);
        }
    }

    @Override
    public String getFileFromRepo(GetFileRequest getFileRequest, String orgName, String teamName) {
        var request = new HttpPost(serverDataEndpoint + String.format("/%s/%s/%s", GET_FILE_EXTENSION, orgName, teamName));
        try {
            var requestBody = objectMapper.writeValueAsString(getFileRequest);
            request.setEntity(new StringEntity(requestBody, ContentType.APPLICATION_JSON));
            var response = httpClient.execute(request);
            log.info("Fetch file request for repo {} + {} returned with status code {}", getFileRequest.getGitRepoSchemaInfo().getGitRepo(), getFileRequest.getGitRepoSchemaInfo().getPathToRoot(), response.getStatusLine().getStatusCode());
            checkResponseStatus(response);
            return new String(response.getEntity().getContent().readAllBytes());
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert GetFileRequest", e);
            throw new AtlasNonRetryableError(e);
        } catch (IOException e) {
            log.error("HTTP get file request failed for repo: {}", getFileRequest.getGitRepoSchemaInfo().getGitRepo(), e);
            throw new AtlasRetryableError(e);
        } finally {
            request.releaseConnection();
        }
    }

    @Override
    public void resetRepoVersion(String gitCommit, GitRepoSchemaInfo gitRepoSchemaInfo, String orgName, String teamName) {
        var request = new HttpPost(serverRepoEndpoint + String.format("/%s/%s/%s/%s", CHANGE_VERSION_EXTENSION, orgName, teamName, gitCommit));
        try {
            var requestBody = objectMapper.writeValueAsString(gitRepoSchemaInfo);
            request.setEntity(new StringEntity(requestBody, ContentType.APPLICATION_JSON));
            var response = httpClient.execute(request);
            log.info("Change version request for repo {} returned with status code {}", gitRepoSchemaInfo.getGitRepo(), response.getStatusLine().getStatusCode());
            checkResponseStatus(response);
        } catch (IOException e) {
            log.error("HTTP Change version request failed for repo: {}", gitRepoSchemaInfo.getGitRepo(), e);
            throw new AtlasRetryableError(e);
        } finally {
            request.releaseConnection();
        }
    }
}
