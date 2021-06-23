package com.greenops.workfloworchestrator.ingest.apiclient.clientwrapper;

import com.greenops.workfloworchestrator.datamodel.requests.DeployResponse;

public interface ClientWrapperApi {
    public DeployResponse deploy(String group, String version, String kind, String body);
    public boolean deleteApplication(String group, String version, String kind, String applicationName);
    public boolean checkStatus(String group, String version, String kind, String applicationName);
    public boolean watchApplication(String pipelineName, String stepName, String applicationName, String namespace);
}
