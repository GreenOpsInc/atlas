package com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper;

public interface ClientRequestQueue {

//    static final String DEPLOY_ARGO_REQUEST = "DeployArgoRequest";
//    static final String DEPLOY_KUBERNETES_REQUEST = "DeployKubernetesRequest";
//    static final String DEPLOY_TEST_REQUEST = "DeployTestRequest";
//    static final String DELETE_ARGO_REQUEST = "DeleteArgoRequest";
//    static final String DELETE_KUBERNETES_REQUEST = "DeleteKubernetesRequest";
//    static final String DELETE_TEST_REQUEST = "DeleteTestRequest";

    public void deploy(String clusterName, String orgName, String teamName, String type, Object payload);

    public void deployAndWatch(String clusterName, String orgName, String teamName, String pipelineName, String stepName, String deployType, Object payload, String watchType);

    public void deployArgoAppByName(String clusterName, String orgName, String teamName, String pipelineName, String stepName, String appName, String watchType);

    public void deployArgoAppByNameAndWatch(String clusterName, String orgName, String teamName, String pipelineName, String stepName, String appName, String watchType);

    public void rollbackAndWatch(String clusterName, String orgName, String teamName, String pipelineName, String stepName, String appName, String revisionHash, String watchType);

    public void deleteByConfig(String clusterName, String orgName, String teamName, String type, String configPayload);

    public void deleteByGvk(String clusterName, String orgName, String teamName, String type, String resourceName, String resourceNamespace, String group, String version, String kind);
}
