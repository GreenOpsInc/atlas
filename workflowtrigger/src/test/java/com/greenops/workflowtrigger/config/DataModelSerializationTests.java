package com.greenops.workflowtrigger.config;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.git.*;
import com.greenops.util.datamodel.pipeline.PipelineSchema;
import com.greenops.util.datamodel.pipeline.PipelineSchemaImpl;
import com.greenops.util.datamodel.pipeline.TeamSchema;
import com.greenops.util.datamodel.pipeline.TeamSchemaImpl;
import net.minidev.json.JSONObject;
import net.minidev.json.parser.JSONParser;
import net.minidev.json.parser.ParseException;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.util.stream.Collectors;

import static org.junit.jupiter.api.Assertions.assertEquals;

public class DataModelSerializationTests {
    ObjectMapper mapper;
    static JSONObject testData;

    @BeforeAll
    static void beforeAll() throws IOException, ParseException {
        var parser = new JSONParser(JSONParser.MODE_JSON_SIMPLE);
        testData = (JSONObject) parser.parse(Files.readAllLines(Paths.get("src/test/java/com/greenops/workflowtrigger/config/DataModelSerializationTestData.json")).stream().collect(Collectors.joining()));
    }

    @BeforeEach
    void beforeEach() {
        mapper = new SpringConfiguration().objectMapper();
    }

    TeamSchema createTeamSchemaObject() {
        var teamSchema = new TeamSchemaImpl("testTeam", "testParentTeam", "testOrg");
        var gitRepoSchema1 = new GitRepoSchema("https://github.com/argoproj/argo-workflows.git", "temporary/temp/workflow", new GitCredMachineUser("root", "admin"));
        var pipeline1 = new PipelineSchemaImpl("testPipeline1", gitRepoSchema1);
        var gitRepoSchema2 = new GitRepoSchema("https://github.com/argoproj/argo-workflows.git", "temporary/temp/workflow2", new GitCredToken("testtoken"));
        var pipeline2 = new PipelineSchemaImpl("testPipeline2", gitRepoSchema2);
        var gitRepoSchema3 = new GitRepoSchema("https://github.com/argoproj/argo-workflows.git", "temporary/temp/workflow1", new GitCredOpen());
        var pipeline3 = new PipelineSchemaImpl("testPipeline3", gitRepoSchema3);
        teamSchema.addPipeline(pipeline1);
        teamSchema.addPipeline(pipeline2);
        teamSchema.addPipeline(pipeline3);
        return teamSchema;
    }


    @Test
    void testGitCredToString() throws JsonProcessingException {
        var gitCredMachineUser = new GitCredMachineUser("root", "admin");
        var gitCredToken = new GitCredToken("token");
        var gitCredOpen = new GitCredOpen();

        assertEquals(
                testData.getAsString("machineUser"),
                mapper.writeValueAsString(gitCredMachineUser)
        );
        assertEquals(
                testData.getAsString("token"),
                mapper.writeValueAsString(gitCredToken)
        );
        assertEquals(
                testData.getAsString("open"),
                mapper.writeValueAsString(gitCredOpen)
        );
    }

    @Test
    void testGitRepoSchemaToString() throws JsonProcessingException {
        var schema = new GitRepoSchema("test-team", "/", new GitCredOpen());
        assertEquals(
                testData.getAsString("gitRepoSchema"),
                mapper.writeValueAsString(schema)
        );
    }

    @Test
    void testPipelineSchemaSerializationToString() throws JsonProcessingException {
        var schema = new PipelineSchemaImpl("test_pipeline", new GitRepoSchema("test", "/", new GitCredOpen()));
        assertEquals(
                testData.getAsString("pipelineSchema"),
                mapper.writeValueAsString(schema)
        );
    }

    @Test
    void testPipelineSchemaStringToObject() throws JsonProcessingException {
        var schema = new PipelineSchemaImpl("test_pipeline", new GitRepoSchema("test", "/", new GitCredOpen()));
        var deserializedPipelineSchema = mapper.readValue(testData.getAsString("pipelineSchema"), PipelineSchema.class);
        assertEquals(schema.getGitRepoSchema(), deserializedPipelineSchema.getGitRepoSchema());
        assertEquals(schema.getPipelineName(), deserializedPipelineSchema.getPipelineName());
    }

    @Test
    void gitRepoStringToObject() throws JsonProcessingException {
        var schema = new GitRepoSchema("test-team", "/", new GitCredOpen());
        var deserializedGitRepoSchema = mapper.readValue(testData.getAsString("gitRepoSchema"), GitRepoSchema.class);
        assertEquals(schema, deserializedGitRepoSchema);
    }

    @Test
    void gitCredMachineUserStringToObject() throws IOException {
        var gitCredMachineUser = new GitCredMachineUser("root", "admin");
        assertEquals(gitCredMachineUser, mapper.readValue(testData.getAsString("machineUser"), GitCred.class));
    }

    @Test
    void gitCredTokenStringToObject() throws JsonProcessingException {
        var gitCredToken = new GitCredToken("token");
        assertEquals(gitCredToken, mapper.readValue(testData.getAsString("token"), GitCred.class));
    }

    @Test
    void gitCredOpenStringToObject() throws JsonProcessingException {
        var gitCredOpen = new GitCredOpen();
        assertEquals(gitCredOpen, mapper.readValue(testData.getAsString("open"), GitCred.class));
    }

    @Test
    void teamSchemaJsonObjectToString() throws IOException, ParseException {
        var teamSchema = createTeamSchemaObject();
        var file = new File("src/test/java/com/greenops/workflowtrigger/config/TestTeamSchema.json");
        assertEquals(mapper.readTree(file), mapper.readTree(mapper.writeValueAsString(teamSchema)));
    }

    @Test
    void teamSchemaJsonToObject() throws IOException {
        var teamSchema = createTeamSchemaObject();
        var file = new File("src/test/java/com/greenops/workflowtrigger/config/TestTeamSchema.json");
        assertEquals(teamSchema, mapper.readValue(file, TeamSchema.class));
    }


}
