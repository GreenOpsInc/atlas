package com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workfloworchestrator.datamodel.requests.DeployResponse;
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
public class ClientWrapperApiImpl implements ClientWrapperApi {

    private final String serverEndpoint;
    private final ObjectMapper objectMapper;
    private final HttpClient httpClient;

    @Autowired
    public ClientWrapperApiImpl(@Value("${application.client-wrapper-url}") String serverEndpoint, @Qualifier("eventAndRequestObjectMapper") ObjectMapper objectMapper) {
        this.serverEndpoint = serverEndpoint.endsWith("/") ? serverEndpoint.substring(0, serverEndpoint.length() - 1) : serverEndpoint;
        httpClient = HttpClientBuilder.create().build();
        this.objectMapper = objectMapper;
    }

    @Override
    public DeployResponse deploy(String group, String version, String kind, String body) {
        try {
            var request = new HttpPost(serverEndpoint + String.format("/deploy/%s/%s/%s", group, version, kind));
            request.setEntity(new StringEntity(body, ContentType.DEFAULT_TEXT));
            var response = httpClient.execute(request);
            return objectMapper.readValue(response.getEntity().getContent().readAllBytes(), DeployResponse.class);
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert payload to DeployResponse", e);
            return null;
        } catch (IOException e) {
            log.error("HTTP deploy request failed", e);
            return null;
        }
    }

    @Override
    public boolean deleteApplication(String group, String version, String kind, String applicationName) {
        try {
            var request = new HttpPost(serverEndpoint + String.format("/delete/%s/%s/%s/%s", group, version, kind, applicationName));
            var response = httpClient.execute(request);
            return objectMapper.readValue(response.getEntity().getContent().readAllBytes(), Boolean.class);
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert payload to boolean", e);
            return false;
        } catch (IOException e) {
            log.error("HTTP delete request failed", e);
            return false;
        }
    }

    @Override
    public boolean checkStatus(String group, String version, String kind, String applicationName) {
        try {
            var request = new HttpPost(serverEndpoint + String.format("/checkStatus/%s/%s/%s/%s", group, version, kind, applicationName));
            var response = httpClient.execute(request);
            return objectMapper.readValue(response.getEntity().getContent().readAllBytes(), Boolean.class);
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert payload to boolean", e);
            return false;
        } catch (IOException e) {
            log.error("HTTP delete request failed", e);
            return false;
        }
    }

    @Override
    public boolean watchApplication(String orgName, String teamName, String pipelineName, String stepName, String namespace, String applicationName) {
        try {
            var request = new HttpPost(serverEndpoint + String.format("/watchApplication/%s/%s/%s/%s/%s/%s", orgName, teamName, pipelineName, stepName, namespace, applicationName));
            var response = httpClient.execute(request);
            return objectMapper.readValue(response.getEntity().getContent().readAllBytes(), Boolean.class);
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert payload to boolean", e);
            return false;
        } catch (IOException e) {
            log.error("HTTP delete request failed", e);
            return false;
        }
    }
}
