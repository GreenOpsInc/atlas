package com.greenops.workflowtrigger.api.model.pipeline;

import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

public class TeamSchemaImpl implements TeamSchema {

    private String teamName;
    private String parentTeam;
    private List<PipelineSchema> pipelines;

    public TeamSchemaImpl(String teamName, String parentTeam) {
        this.teamName = teamName;
        this.parentTeam = parentTeam;
        this.pipelines = new ArrayList<>();
    }

    private TeamSchemaImpl(String teamName, String parentTeam, List<PipelineSchema> pipelines) {
        this.teamName = teamName;
        this.parentTeam = parentTeam;
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
    public List<String> getPipelineNames() {
        return pipelines.stream().map(PipelineSchema::getPipelineName).collect(Collectors.toList());
    }
}
