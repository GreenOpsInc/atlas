package com.greenops.workflowtrigger.api.model.pipeline;

import com.greenops.workflowtrigger.api.model.git.GitCredOpen;
import com.greenops.workflowtrigger.api.model.git.GitRepoSchema;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

public class TeamAndPipelineSchemaTest {

    private TeamSchema teamSchema;

    @BeforeEach
    void beforeEach() {
        teamSchema = new TeamSchemaImpl("team1", TeamSchema.ROOT_TEAM, "org name");
    }

    @Test
    void testTeamSchemaCanAddAndRemovePipelines() {
        var gitRepoSchema = new GitRepoSchema("https://github.com/argoproj/argocd-example-apps.git", "guestbook/", new GitCredOpen());
        var pipeline1 = new PipelineSchemaImpl("pipeline1", gitRepoSchema);
        teamSchema.addPipeline(pipeline1);
        teamSchema.addPipeline("pipeline2", gitRepoSchema);
        assertTrue(teamSchema.getPipelineNames().contains("pipeline1"));
        assertTrue(teamSchema.getPipelineNames().contains("pipeline2"));
        assertEquals(teamSchema.getPipelineNames().size(),2);
        teamSchema.removePipeline(pipeline1.getPipelineName());
        assertEquals(teamSchema.getPipelineNames().size(),1);
        assertTrue(teamSchema.getPipelineNames().contains("pipeline2"));
        assertEquals("team1", teamSchema.getTeamName());
        assertEquals(TeamSchema.ROOT_TEAM, teamSchema.getParentTeam());
    }
}
