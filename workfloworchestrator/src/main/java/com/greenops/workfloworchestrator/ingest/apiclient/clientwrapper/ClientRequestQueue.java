package com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper;

import com.greenops.util.datamodel.clientmessages.ResourcesGvkRequest;

public interface ClientRequestQueue {

    static final String LATEST_REVISION = "LATEST_REVISION";

    static final String DEPLOY_ARGO_REQUEST = "DeployArgoRequest";
    static final String DEPLOY_KUBERNETES_REQUEST = "DeployKubernetesRequest";
    static final String DEPLOY_TEST_REQUEST = "DeployTestRequest";
    static final String DELETE_ARGO_REQUEST = "DeleteArgoRequest";
    static final String DELETE_KUBERNETES_REQUEST = "DeleteKubernetesRequest";
    static final String DELETE_TEST_REQUEST = "DeleteTestRequest";

    static final String RESPONSE_EVENT_APPLICATION_INFRA = "ApplicationInfraCompletionEvent";

    public void deploy(String clusterName, String orgName, String teamName, String pipelineName, String uvn, String stepName, String responseEventType, String type, String revisionHash, Object payload);

    public void deployAndWatch(String clusterName, String orgName, String teamName, String pipelineName, String uvn, String stepName, String deployType, String revisionHash, Object payload, String watchType, int testNumber);

    public void selectiveSyncArgoApplication(String clusterName, String orgName, String teamName, String pipelineName, String uvn, String stepName, String revisionHash, ResourcesGvkRequest resourcesGvkRequest, String appName);

    public void deployArgoAppByName(String clusterName, String orgName, String pipelineName, String stepName, String appName, String watchType);

    public void deployArgoAppByNameAndWatch(String clusterName, String orgName, String teamName, String pipelineName, String uvn, String stepName, String appName, String watchType);

    public void rollbackAndWatch(String clusterName, String orgName, String teamName, String pipelineName, String uvn, String stepName, String appName, String revisionHash, String watchType);

    public void deleteByConfig(String clusterName, String orgName, String teamName, String pipelineName, String uvn, String stepName, String type, String configPayload);

    public void deleteByGvk(String clusterName, String orgName, String teamName, String pipelineName, String uvn, String stepName, String type, String resourceName, String resourceNamespace, String group, String version, String kind);
}
