package com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.workfloworchestrator.datamodel.requests.DeployResponse;
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
    public DeployResponse deploy(String clusterName, String orgName, String teamName, String pipelineName, String stepName, String type, Object payload) {
        var request = new HttpPost(serverEndpoint + String.format("/deploy/%s/%s/%s/%s/%s", orgName, teamName, pipelineName, stepName, type));
        try {
            var body = type.equals(DEPLOY_TEST_REQUEST) ? objectMapper.writeValueAsString(payload) : (String)payload;
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
    public DeployResponse deployArgoAppByName(String clusterName, String orgName, String appName) {
        var request = new HttpPost(serverEndpoint + String.format("/deploy/%s/%s/%s", orgName, DEPLOY_ARGO_REQUEST, appName));
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
    public DeployResponse rollback(String clusterName, String orgName, String appName, String revisionHash) {
        var request = new HttpPost(serverEndpoint + String.format("/rollback/%s/%s/%s", orgName, appName, revisionHash));
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
    public void delete(String clusterName, String orgName, String type, String resourceName, String resourceNamespace, String group, String version, String kind) {
        var request = new HttpPost(serverEndpoint + String.format("/delete/%s/%s/%s/%s/%s/%s/%s", orgName, type, resourceName, resourceNamespace, group, version, kind));
        try {
            var response = httpClient.execute(request);
            checkResponseStatus(response);
            if (!objectMapper.readValue(response.getEntity().getContent().readAllBytes(), Boolean.class)) {
                throw new AtlasRetryableError("Delete request failed.");
            }
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
    public void delete(String clusterName, String orgName, String type, String configPayload) {
        var request = new HttpPost(serverEndpoint + String.format("/delete/%s/%s", orgName, type));
        try {
            request.setEntity(new StringEntity(configPayload, ContentType.DEFAULT_TEXT));
            var response = httpClient.execute(request);
            checkResponseStatus(response);
            if (!objectMapper.readValue(response.getEntity().getContent().readAllBytes(), Boolean.class)) {
                throw new AtlasRetryableError("Delete request failed.");
            }
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
    public void deleteApplication(String clusterName, String group, String version, String kind, String applicationName) {
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
    public void watchApplication(String clusterName, String orgName, WatchRequest watchRequest) {
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
