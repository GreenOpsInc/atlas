package com.greenops.workfloworchestrator.ingest.apiclient.reposerver;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workfloworchestrator.datamodel.requests.GetFileRequest;
import lombok.extern.slf4j.Slf4j;
import org.apache.http.client.HttpClient;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.HttpClientBuilder;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.io.IOException;

@Slf4j
@Component
public class RepoManagerApiImpl implements RepoManagerApi {

    private final String serverEndpoint;
    private final ObjectMapper objectMapper;
    private final HttpClient httpClient;

    @Autowired
    public RepoManagerApiImpl(@Value("${application.repo-server-url}") String serverEndpoint, @Qualifier("requestObjectMapper") ObjectMapper objectMapper) {
        this.serverEndpoint = serverEndpoint.endsWith("/") ? serverEndpoint + "yaml" : serverEndpoint + "/yaml";
        httpClient = HttpClientBuilder.create().build();
        this.objectMapper = objectMapper;
    }

    @Override
    public boolean getFileFromRepo(GetFileRequest getFileRequest, String orgName, String teamName) {
        try {
            var requestBody = objectMapper.writeValueAsString(getFileRequest);
            var request = new HttpPost(serverEndpoint + String.format("%s/%s", orgName, teamName));
            request.setEntity(new StringEntity(requestBody, ContentType.APPLICATION_JSON));
            var response = httpClient.execute(request);
            log.info("Fetch file request for repo {} returned with status code {}", getFileRequest.getGitRepoUrl(), response.getStatusLine().getStatusCode());
            return response.getStatusLine().getStatusCode() == 200;
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert GetFileRequest", e);
            return false;
        } catch (IOException e) {
            log.error("HTTP get file request failed for repo: {}", getFileRequest.getGitRepoUrl(), e);
            return false;
        }
    }
}
