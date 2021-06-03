package com.greenops.workflowtrigger.api.reposerver;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workflowtrigger.api.model.git.GitCredMachineUser;
import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;
import com.greenops.workflowtrigger.api.model.mixin.git.GitCredMachineUserMixin;
import com.greenops.workflowtrigger.api.model.mixin.git.GitRepoSchemaMixin;
import org.apache.http.client.HttpClient;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.HttpClientBuilder;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.io.IOException;

@Component
public class RepoManagerApiImpl implements RepoManagerApi {

    private final String serverEndpoint;
    private final ObjectMapper objectMapper;
    private final HttpClient httpClient;

    public RepoManagerApiImpl(@Value("${application.repo-server-url}") String serverEndpoint) {
        this.serverEndpoint = serverEndpoint.endsWith("/") ? serverEndpoint + "repo" : serverEndpoint + "/repo";
        httpClient = HttpClientBuilder.create().build();
        objectMapper = new ObjectMapper()
                .addMixIn(GitRepoSchema.class, GitRepoSchemaMixin.class)
                .addMixIn(GitCredMachineUser.class, GitCredMachineUserMixin.class);
    }

    @Override
    public boolean cloneRepo(GitRepoSchema gitRepoSchema) {
        try {
            var requestBody = objectMapper.writeValueAsString(gitRepoSchema);
            var request = new HttpPost(serverEndpoint);
            request.setEntity(new StringEntity(requestBody, ContentType.APPLICATION_JSON));
            var response = httpClient.execute(request);
            return response.getStatusLine().getStatusCode() == 200;
        } catch (JsonProcessingException e) {
            return false;
        } catch (IOException e) {
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
            return response.getStatusLine().getStatusCode() == 200;
        } catch (JsonProcessingException e) {
            return false;
        } catch (IOException e) {
            return false;
        }
        //TODO: Catch branches left separate for future processing, logic, and logging.
    }
}
