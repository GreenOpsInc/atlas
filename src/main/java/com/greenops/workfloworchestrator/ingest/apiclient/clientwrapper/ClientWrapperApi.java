package com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper;

import com.greenops.workfloworchestrator.datamodel.requests.DeployResponse;
import com.greenops.workfloworchestrator.datamodel.requests.WatchRequest;

public interface ClientWrapperApi {

    static final String DEPLOY_ARGO_REQUEST = "DeployArgoRequest";
    static final String DEPLOY_KUBERNETES_REQUEST = "DeployKubernetesRequest";
    static final String DEPLOY_TEST_REQUEST = "DeployTestRequest";

    public DeployResponse deploy(String orgName, String type, Object payload);

    public DeployResponse rollback(String orgName, String appName, int revisionId);

    public void deleteApplication(String group, String version, String kind, String applicationName);

    public void watchApplication(String orgName, WatchRequest watchRequest);
}
