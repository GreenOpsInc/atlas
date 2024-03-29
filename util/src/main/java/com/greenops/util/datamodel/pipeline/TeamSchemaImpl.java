package com.greenops.util.datamodel.pipeline;


import com.greenops.util.datamodel.git.GitRepoSchema;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;

public class TeamSchemaImpl implements TeamSchema {

    private String teamName;
    private String parentTeam;
    private String orgName;
    private List<PipelineSchema> pipelines;

    public TeamSchemaImpl(String teamName, String parentTeam, String orgName) {
        this.teamName = teamName;
        this.parentTeam = parentTeam;
        this.orgName = orgName;
        this.pipelines = new ArrayList<>();
    }

    private TeamSchemaImpl(String teamName, String parentTeam, String orgName, List<PipelineSchema> pipelines) {
        this.teamName = teamName;
        this.parentTeam = parentTeam;
        this.orgName = orgName;
        this.pipelines = pipelines;
    }

    @Override
    public void setTeamName(String teamName) {
        this.teamName = teamName;
    }

    @Override
    public void setParentTeam(String parentTeam) {
        this.parentTeam = parentTeam;
    }

    @Override
    public void addPipeline(String pipelineName, GitRepoSchema gitRepoSchema) {
        pipelines.add(new PipelineSchemaImpl(pipelineName, gitRepoSchema));
    }

    @Override
    public void addPipeline(PipelineSchema pipelineSchema) {
        pipelines.add(pipelineSchema);
    }

    @Override
    public void removePipeline(String pipelineName) {
        pipelines = pipelines.stream().filter(
                pipelineSchema -> !pipelineSchema.getPipelineName().equals(pipelineName)
        ).collect(Collectors.toList());
    }

    @Override
    public String getTeamName() {
        return teamName;
    }

    @Override
    public String getParentTeam() {
        return parentTeam;
    }

    @Override
    public String getOrgName() {
        return orgName;
    }

    @Override
    public List<String> getPipelineNames() {
        return pipelines.stream().map(PipelineSchema::getPipelineName).collect(Collectors.toList());
    }

    @Override
    public List<PipelineSchema> getPipelineSchemas() {
        return pipelines;
    }


    @Override
    public PipelineSchema getPipelineSchema(String pipelineName) {
        var result = pipelines.stream().filter(
                pipelineSchema -> pipelineSchema.getPipelineName().equals(pipelineName)
        ).collect(Collectors.toList());
        if (result.size() > 0) {
            return result.get(0);
        }
        return null;
    }


    @Override
    public boolean equals(Object o) {
        if (o instanceof TeamSchemaImpl) {
            return (
                    (TeamSchemaImpl) o).orgName.equals(orgName)
                    && ((TeamSchemaImpl) o).parentTeam.equals(parentTeam)
                    && ((TeamSchemaImpl) o).teamName.equals(teamName)
                    && Arrays.deepEquals(pipelines.toArray(), ((TeamSchemaImpl) o).pipelines.toArray()
            );
        }
        return false;
    }
}
