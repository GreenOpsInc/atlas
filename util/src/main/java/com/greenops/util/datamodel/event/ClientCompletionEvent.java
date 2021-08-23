package com.greenops.util.datamodel.event;

import java.util.ArrayList;
import java.util.List;

public class ClientCompletionEvent implements Event {

    //Argo health status codes
    public static final String HEALTHY = "Healthy";
    public static final String PROGRESSING = "Progressing";
    //Unknown is also for syncs
    public static final String UNKNOWN = "Unknown";
    public static final String DEGRADED = "Degraded";
    public static final String SUSPENDED = "Suspended";
    public static final String MISSING = "Missing";

    //Argo sync statuses
    public static final String SYNCED = "Synced";
    public static final String OUT_OF_SYNC = "OutOfSync";

    private String healthStatus;
    private String syncStatus;
    private List<ResourceStatus> resourceStatuses;
    private String orgName;
    private String teamName;
    private String pipelineName;
    private String stepName;
    private String argoName;
    private String operation;
    private String project;
    private String repo;
    private String revisionHash;

    public ClientCompletionEvent(String healthStatus, String syncStatus, String orgName, String teamName, String pipelineName, String stepName,
                                 String argoName, String operation, String project, String repo, String revisionHash, List<ResourceStatus> resourceStatuses) {
        this.healthStatus = healthStatus;
        this.syncStatus = syncStatus;
        this.orgName = orgName;
        this.teamName = teamName;
        this.pipelineName = pipelineName;
        this.stepName = stepName;
        this.argoName = argoName;
        this.operation = operation;
        this.project = project;
        this.repo = repo;
        this.revisionHash = revisionHash;
        this.resourceStatuses = resourceStatuses;
    }

    public ClientCompletionEvent(String healthStatus, String syncStatus, String orgName, String teamName, String pipelineName, String stepName,
                                 String argoName, String operation, String project, String repo, String revisionHash) {
        this(healthStatus, syncStatus, orgName, teamName, pipelineName, stepName, argoName, operation, project, repo, revisionHash, new ArrayList<>());
    }

    @Override
    public String getOrgName() {
        return orgName;
    }

    @Override
    public String getTeamName() {
        return teamName;
    }

    @Override
    public String getPipelineName() {
        return pipelineName;
    }

    @Override
    public String getStepName() {
        return stepName;
    }

    public String getHealthStatus() {
        return healthStatus;
    }

    public String getSyncStatus() {
        return syncStatus;
    }

    public String getRepoUrl() {
        return repo;
    }

    public String getRevisionHash() {
        return revisionHash;
    }

    public List<ResourceStatus> getResourceStatuses() {
        return resourceStatuses;
    }
}