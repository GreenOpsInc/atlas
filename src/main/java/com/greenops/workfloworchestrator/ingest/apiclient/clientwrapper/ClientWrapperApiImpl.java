package com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workfloworchestrator.datamodel.requests.DeployResponse;
import com.greenops.workfloworchestrator.datamodel.requests.KubernetesCreationRequest;
import com.greenops.workfloworchestrator.datamodel.requests.WatchRequest;
import com.greenops.workfloworchestrator.error.AtlasNonRetryableError;
import com.greenops.workfloworchestrator.error.AtlasRetryableError;
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
import java.util.Optional;

import static com.greenops.workfloworchestrator.ingest.apiclient.util.ApiClientUtil.checkResponseStatus;

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
    public DeployResponse deploy(String orgName, String type, Optional<String> configPayload, Optional<KubernetesCreationRequest> kubernetesCreationRequest) {
        var request = new HttpPost(serverEndpoint + String.format("/deploy/%s/%s", orgName, type));
        try {
            var body = type.equals(DEPLOY_TEST_REQUEST) ? objectMapper.writeValueAsString(kubernetesCreationRequest.get()) : configPayload.get();
            request.setEntity(new StringEntity(body, ContentType.DEFAULT_TEXT));
            var response = httpClient.execute(request);
            checkResponseStatus(response);
            return objectMapper.readValue(response.getEntity().getContent().readAllBytes(), DeployResponse.class);
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert payload to DeployResponse", e);
            throw new AtlasNonRetryableError(e);
        } catch (IOException e) {
            log.error("HTTP deploy request failed", e);
            throw new AtlasRetryableError(e);
        } finally {
            request.releaseConnection();
        }
    }

    @Override
    public DeployResponse rollback(String orgName, String appName, int revisionId) {
        var request = new HttpPost(serverEndpoint + String.format("/rollback/%s/%s/%d", orgName, appName, revisionId));
        try {
            var response = httpClient.execute(request);
            checkResponseStatus(response);
            return objectMapper.readValue(response.getEntity().getContent().readAllBytes(), DeployResponse.class);
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert payload to DeployResponse", e);
            throw new AtlasNonRetryableError(e);
        } catch (IOException e) {
            log.error("HTTP deploy request failed", e);
            throw new AtlasRetryableError(e);
        } finally {
            request.releaseConnection();
        }
    }

    @Override
    public void deleteApplication(String group, String version, String kind, String applicationName) {
        var request = new HttpPost(serverEndpoint + String.format("/delete/%s/%s/%s/%s", group, version, kind, applicationName));
        try {
            var response = httpClient.execute(request);
            checkResponseStatus(response);
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert payload to boolean", e);
            throw new AtlasNonRetryableError(e);
        } catch (IOException e) {
            log.error("HTTP delete request failed", e);
            throw new AtlasRetryableError(e);
        } finally {
            request.releaseConnection();
        }
    }

    @Override
    public void watchApplication(String orgName, WatchRequest watchRequest) {
        var request = new HttpPost(serverEndpoint + String.format("/watch/%s", orgName));
        try {
            var body = objectMapper.writeValueAsString(watchRequest);
            request.setEntity(new StringEntity(body, ContentType.DEFAULT_TEXT));
            var response = httpClient.execute(request);
            checkResponseStatus(response);
        } catch (JsonProcessingException e) {
            log.error("Object mapper could not convert payload to boolean", e);
            throw new AtlasNonRetryableError(e);
        } catch (IOException e) {
            log.error("HTTP delete request failed", e);
            throw new AtlasRetryableError(e);
        } finally {
            request.releaseConnection();
        }
    }
}
