package com.greenops.workflowtrigger.api.model.pipeline;

import com.greenops.util.datamodel.git.GitCredMachineUser;
import com.greenops.util.datamodel.git.GitCredOpen;
import com.greenops.util.datamodel.git.GitCredToken;
import com.greenops.util.datamodel.git.GitRepoSchema;
import com.greenops.util.datamodel.pipeline.PipelineSchemaImpl;
import com.greenops.util.datamodel.pipeline.TeamSchema;
import com.greenops.util.datamodel.pipeline.TeamSchemaImpl;
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
        assertEquals(teamSchema.getPipelineNames().size(), 2);
        teamSchema.removePipeline(pipeline1.getPipelineName());
        assertEquals(teamSchema.getPipelineNames().size(), 1);
        assertTrue(teamSchema.getPipelineNames().contains("pipeline2"));
        assertEquals("team1", teamSchema.getTeamName());
        assertEquals(TeamSchema.ROOT_TEAM, teamSchema.getParentTeam());
    }

    @Test
    void testPipelineSchemaCanSetAndGetPipeline() {
        var gitCredMachineUser = new GitCredMachineUser("root", "admin");
        var gitRepoSchema = new GitRepoSchema("https://github.com/argoproj/argocd-example-apps.git", "guestbook/", gitCredMachineUser);
        var pipelineSchema = new PipelineSchemaImpl("test_pipeline", gitRepoSchema);
        assertEquals("test_pipeline", pipelineSchema.getPipelineName());
        assertEquals("https://github.com/argoproj/argocd-example-apps.git", pipelineSchema.getGitRepoSchema().getGitRepo());
        assertEquals("guestbook/", pipelineSchema.getGitRepoSchema().getPathToRoot());
        pipelineSchema.setPipelineName("changed_test_pipeline_name");
        assertEquals("changed_test_pipeline_name", pipelineSchema.getPipelineName());

        gitRepoSchema = new GitRepoSchema("https://github.com/argoproj/argo-workflows.git", "src/", new GitCredOpen());
        pipelineSchema.setGitRepoSchema(gitRepoSchema);
        assertEquals("https://github.com/argoproj/argo-workflows.git", pipelineSchema.getGitRepoSchema().getGitRepo());
        assertEquals("src/", pipelineSchema.getGitRepoSchema().getPathToRoot());
        assert (pipelineSchema.getGitRepoSchema().getGitCred() instanceof GitCredOpen);

        gitRepoSchema.setGitCred(new GitCredToken("test_token"));
        assert (pipelineSchema.getGitRepoSchema().getGitCred() instanceof GitCredToken);

        gitRepoSchema.setGitCred(new GitCredMachineUser("test_user", "test_pwd"));
        assert (pipelineSchema.getGitRepoSchema().getGitCred() instanceof GitCredMachineUser);
    }

    @Test
    void testGitRepoSchema() {
        var gitRepoSchema = new GitRepoSchema("https://github.com/argoproj/argo-workflows.git", "src/", new GitCredOpen());
        assertEquals("https://github.com/argoproj/argo-workflows.git", gitRepoSchema.getGitRepo());
        assertEquals("src/", gitRepoSchema.getPathToRoot());

        gitRepoSchema.setGitRepo("https://github.com/argoproj/argocd-example.git");
        assertEquals("https://github.com/argoproj/argocd-example.git", gitRepoSchema.getGitRepo());

        gitRepoSchema.setPathToRoot("test/");
        assertEquals("test/", gitRepoSchema.getPathToRoot());
    }
}
