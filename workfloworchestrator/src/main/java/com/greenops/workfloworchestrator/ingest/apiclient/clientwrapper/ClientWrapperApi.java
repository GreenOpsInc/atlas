package com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper;

import com.greenops.workfloworchestrator.datamodel.requests.DeployResponse;
import com.greenops.workfloworchestrator.datamodel.requests.WatchRequest;

public interface ClientWrapperApi {

    static final String LATEST_REVISION = "LATEST_REVISION";

    static final String DEPLOY_ARGO_REQUEST = "DeployArgoRequest";
    static final String DEPLOY_KUBERNETES_REQUEST = "DeployKubernetesRequest";
    static final String DEPLOY_TEST_REQUEST = "DeployTestRequest";
    static final String DELETE_ARGO_REQUEST = "DeleteArgoRequest";
    static final String DELETE_KUBERNETES_REQUEST = "DeleteKubernetesRequest";
    static final String DELETE_TEST_REQUEST = "DeleteTestRequest";

    public DeployResponse deploy(String clusterName, String orgName, String type, String revisionHash, Object payload);

    public DeployResponse deployArgoAppByName(String clusterName, String orgName, String appName);

    public DeployResponse rollback(String clusterName, String orgName, String appName, String revisionHash);

    public void delete(String clusterName, String orgName, String type, String resourceName, String resourceNamespace, String group, String version, String kind);

    public void delete(String clusterName, String orgName, String type, String configPayload);

    public void deleteApplication(String clusterName, String group, String version, String kind, String applicationName);

    public void watchApplication(String clusterName, String orgName, WatchRequest watchRequest);
}
