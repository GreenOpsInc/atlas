package com.greenops.verificationtool.ingest.apiclient.workflowtrigger;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.git.GitCredOpen;
import com.greenops.util.datamodel.git.GitRepoSchema;
import com.greenops.util.error.AtlasRetryableError;
import lombok.extern.slf4j.Slf4j;
import org.apache.http.client.HttpClient;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.config.RegistryBuilder;
import org.apache.http.conn.socket.ConnectionSocketFactory;
import org.apache.http.conn.socket.PlainConnectionSocketFactory;
import org.apache.http.conn.ssl.NoopHostnameVerifier;
import org.apache.http.conn.ssl.SSLConnectionSocketFactory;
import org.apache.http.entity.ContentType;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClientBuilder;
import org.apache.http.impl.conn.PoolingHttpClientConnectionManager;
import org.apache.http.ssl.SSLContextBuilder;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import javax.net.ssl.SSLContext;
import java.io.IOException;
import java.security.KeyManagementException;
import java.security.KeyStoreException;
import java.security.NoSuchAlgorithmException;
import java.util.Arrays;

import static com.greenops.verificationtool.ingest.apiclient.util.ApiClientUtil.checkResponseStatus;

@Slf4j
@Component
public class WorkflowTriggerApiImpl implements WorkflowTriggerApi {
    private final String UVN = "LATEST";
    private final String authToken = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDQ5MjQwMzAsImp0aSI6ImNmMjc0YTRlLTA3M2QtNGU3OC1hM2E3LWRiMjM4NTc4NjU5OSIsImlhdCI6MTY0NDgzNzYzMCwiaXNzIjoiYXJnb2NkIiwibmJmIjoxNjQ0ODM3NjMwLCJzdWIiOiJhZG1pbjpsb2dpbiJ9.sp7I3K9ufcDraV9kvRKvxlgSFrX-WEowgKzLu-BFsQU";
    private final String pipelineRevisionHash = "ROOT_COMMIT";
    private final String serverWorkflowTriggerEndpoint;
    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;

    @Autowired
    public WorkflowTriggerApiImpl(@Value("${application.workflowtrigger-url}") String serverEndpoint, @Qualifier("objectMapper") ObjectMapper objectMapper) throws NoSuchAlgorithmException, KeyManagementException, KeyStoreException {
        this.serverWorkflowTriggerEndpoint = serverEndpoint;
        this.objectMapper = objectMapper;
        this.httpClient = createHttpClient();
    }

    private CloseableHttpClient createHttpClient() throws NoSuchAlgorithmException, KeyStoreException, KeyManagementException {
        final SSLContext sslContext = new SSLContextBuilder()
                .loadTrustMaterial(null, (x509CertChain, authType) -> true)
                .build();

        return HttpClientBuilder.create()
                .setSSLContext(sslContext)
                .setConnectionManager(
                        new PoolingHttpClientConnectionManager(
                                RegistryBuilder.<ConnectionSocketFactory>create()
                                        .register("http", PlainConnectionSocketFactory.INSTANCE)
                                        .register("https", new SSLConnectionSocketFactory(sslContext,
                                                NoopHostnameVerifier.INSTANCE))
                                        .build()
                        ))
                .build();
    }

    @Override
    public void createTeam(String orgName, String parentTeamName, String teamName) {
        var url = String.format("%s/team/%s/%s/%s", this.serverWorkflowTriggerEndpoint, orgName, parentTeamName, teamName);
        var request = new HttpPost(url);
        try {
            request.setHeader("Authorization", this.authToken);
            var response = httpClient.execute(request);
            log.info("create team request for team {} returned with status code {}", teamName, response.getStatusLine().getStatusCode());
            if (response.getStatusLine().getStatusCode() != 200) {
                throw new RuntimeException(Arrays.toString(response.getEntity().getContent().readAllBytes()));
            }
        } catch (IOException e) {
            log.error("HTTP create team request failed for team: {}", teamName, e);
            throw new AtlasRetryableError(e);
        } finally {
            request.releaseConnection();
        }
    }

    @Override
    public void createPipeline(String orgName, String pipelineName, String teamName, String gitRepoUrl, String pathToRoot) {
        var url = String.format("%s/pipeline/%s/%s/%s", this.serverWorkflowTriggerEndpoint, orgName, teamName, pipelineName);
        var request = new HttpPost(url);
        GitRepoSchema gitRepoSchema = new GitRepoSchema(gitRepoUrl, pathToRoot, new GitCredOpen());
        try {
            request.setHeader("Authorization", this.authToken);
            request.setEntity(new StringEntity(this.objectMapper.writeValueAsString(gitRepoSchema), ContentType.APPLICATION_JSON));
            var response = httpClient.execute(request);
            log.info("create pipeline request for team {} and pipeline {} returned with status code {}", teamName, pipelineName, response.getStatusLine().getStatusCode());
            if (response.getStatusLine().getStatusCode() != 200) {
                throw new RuntimeException(Arrays.toString(response.getEntity().getContent().readAllBytes()));
            }
        } catch (IOException e) {
            e.printStackTrace();
        } finally {
            request.releaseConnection();
        }
    }

    @Override
    public void syncPipeline(String orgName, String pipelineName, String teamName, String gitRepoUrl, String pathToRoot) {
        var url = String.format("%s/sync/%s/%s/%s/%s", this.serverWorkflowTriggerEndpoint, orgName, teamName, pipelineName, this.pipelineRevisionHash);
        var request = new HttpPost(url);
        GitRepoSchema gitRepoSchema = new GitRepoSchema(gitRepoUrl, pathToRoot, new GitCredOpen());
        try {
            request.setHeader("Authorization", this.authToken);
            request.setEntity(new StringEntity(this.objectMapper.writeValueAsString(gitRepoSchema), ContentType.APPLICATION_JSON));
            var response = httpClient.execute(request);
            log.info("sync pipeline request for team {} and pipeline {} returned with status code {}", teamName, pipelineName, response.getStatusLine().getStatusCode());
            if (response.getStatusLine().getStatusCode() != 200) {
                throw new RuntimeException(Arrays.toString(response.getEntity().getContent().readAllBytes()));
            }
        } catch (IOException e) {
            e.printStackTrace();
        } finally {
            request.releaseConnection();
        }
    }

    @Override
    public String getPipelineStatus(String orgName, String pipelineName, String teamName) {
        var url = String.format("%s/status/%s/%s/pipeline/%s/%s", this.serverWorkflowTriggerEndpoint, orgName, teamName, pipelineName, this.UVN);
        var request = new HttpGet(url);
        try {
            request.setHeader("Authorization", this.authToken);
            var response = httpClient.execute(request);
            log.info("GET pipeline status request for pipeline {} returned with status code {}", pipelineName, response.getStatusLine().getStatusCode());
            checkResponseStatus(response);
            return new String(response.getEntity().getContent().readAllBytes());
        } catch (IOException e) {
            log.error("HTTP get pipeline status request failed for pipeline: {}", pipelineName, e);
            throw new AtlasRetryableError(e);
        } finally {
            request.releaseConnection();
        }
    }

    @Override
    public String getStepLevelStatus(String orgName, String pipelineName, String teamName, String stepName, Integer count) {
        var url = String.format("%s/status/%s/%s/pipeline/%s/step/%s/%s", this.serverWorkflowTriggerEndpoint, orgName, teamName, pipelineName, stepName, count.toString());
        var request = new HttpGet(url);
        try {
            request.setHeader("Authorization", this.authToken);
            var response = httpClient.execute(request);
            log.info("GET step level status request for pipeline {} returned with status code {}", pipelineName, response.getStatusLine().getStatusCode());
            checkResponseStatus(response);
            return new String(response.getEntity().getContent().readAllBytes());
        } catch (IOException e) {
            log.error("HTTP get pipeline status request failed for pipeline: {}", pipelineName, e);
            throw new AtlasRetryableError(e);
        } finally {
            request.releaseConnection();
        }
    }
}
