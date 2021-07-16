package com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper;

import com.greenops.workfloworchestrator.datamodel.requests.DeployResponse;
import com.greenops.workfloworchestrator.datamodel.requests.KubernetesCreationRequest;
import com.greenops.workfloworchestrator.datamodel.requests.WatchRequest;

import java.util.Optional;

public interface ClientWrapperApi {

    public static final String DEPLOY_ARGO_REQUEST = "DeployArgoRequest";
    public static final String DEPLOY_KUBERNETES_REQUEST = "DeployKubernetesRequest";
    public static final String DEPLOY_TEST_REQUEST = "DeployTestRequest";

    public DeployResponse deploy(String orgName, String type, Optional<String> configPayload, Optional<KubernetesCreationRequest> kubernetesCreationRequest);

    public DeployResponse rollback(String orgName, String appName, int revisionId);

    public boolean deleteApplication(String group, String version, String kind, String applicationName);

    public boolean checkStatus(String group, String version, String kind, String applicationName);

    public boolean watchApplication(String orgName, WatchRequest watchRequest);
}
