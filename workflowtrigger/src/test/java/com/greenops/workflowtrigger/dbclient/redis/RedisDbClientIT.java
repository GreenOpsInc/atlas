package com.greenops.workflowtrigger.dbclient.redis;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.git.GitCredOpen;
import com.greenops.util.datamodel.git.GitRepoSchema;
import com.greenops.util.datamodel.pipeline.PipelineSchema;
import com.greenops.util.datamodel.pipeline.TeamSchema;
import com.greenops.util.datamodel.pipeline.TeamSchemaImpl;
import com.greenops.util.dbclient.DbClient;
import com.greenops.util.dbclient.redis.RedisDbClient;
import com.greenops.workflowtrigger.kubernetesclient.KubernetesClient;
import org.junit.ClassRule;
import org.junit.jupiter.api.*;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.testcontainers.containers.DockerComposeContainer;
import org.testcontainers.containers.wait.strategy.Wait;
import org.testcontainers.junit.jupiter.Testcontainers;

import java.io.File;
import java.util.ArrayList;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@Testcontainers
public class RedisDbClientIT {

    //ClassRule makes sure the container is for the lifecycle of the class, not a test
    @ClassRule
    private static final DockerComposeContainer<?> compose =
            new DockerComposeContainer<>(new File("src/test/test_config/docker-compose.yml"))
                    .withServices("redisserver")
                    .waitingFor("redisserver", Wait.forListeningPort());

    @MockBean
    KubernetesClient kubernetesClient;

    @Value("${application.redis-url}")
    String redisUrl;

    @Autowired
    private ObjectMapper objectMapper;

    private DbClient redisDbClient;

    @BeforeAll
    static void beforeAll() {
        //beforeAll will be executed before Spring dependency injection
        compose.start();
    }

    @BeforeEach
    void beforeEach() {
        redisDbClient = new RedisDbClient(redisUrl, objectMapper);
    }

    @AfterEach
    void afterEach() {
        System.err.println("AFTER REACH RUNNING");
        ((RedisDbClient) redisDbClient).shutdown();
    }

    @Test
    void redisDbClientStoresAndFetchesTeamSchemaCorrectly() {
        var originalTeam = new TeamSchemaImpl("team1", TeamSchema.ROOT_TEAM, "test_organization");
        originalTeam.addPipeline(
                "pipeline1",
                new GitRepoSchema("http://fake.com/repo.git", "/", new GitCredOpen())
        );
        try {
            redisDbClient.storeValue("team-test-key", originalTeam);
        } catch (RuntimeException e) {
            fail();
        }
        var fetchedTeam = redisDbClient.fetchTeamSchema("team-test-key");
        assertNotNull(fetchedTeam);
        assertEquals(originalTeam.getTeamName(), fetchedTeam.getTeamName());
        assertTeamsEqual(originalTeam, fetchedTeam);
    }

    @Test
    void redisDbClientStoresAndFetchesListSucceeds() {
        var teamList = new ArrayList<>();
        teamList.add("team1");
        teamList.add("team2");
        try {
            redisDbClient.storeValue("list-test-key", teamList);
        } catch (RuntimeException e) {
            fail();
        }
        var fetchedList = redisDbClient.fetchStringList("list-test-key");
        assertNotNull(fetchedList);
        assertEquals(teamList, fetchedList);
    }

    @Test
    void redisDbClientFetchesNonexistentKeyReturnsNull() {
        var fetchedTeam = redisDbClient.fetchTeamSchema("nonexistant-key1");
        var fetchedList = redisDbClient.fetchStringList("nonexistant-key2");
        assertNull(fetchedTeam);
        assertNull(fetchedList);
    }

    @Test
    void redisDbClientStoreFailsWhenTransactionInterrupted() {
        var testKey = "transaction-test-key";
        var redisDbClient2 = new RedisDbClient(redisUrl, objectMapper);
        var newTeam1 = new TeamSchemaImpl("team1", TeamSchema.ROOT_TEAM, "test_organization");
        var newTeam2 = new TeamSchemaImpl("team2", TeamSchema.ROOT_TEAM, "test_organization");

        var fetchedTeam = redisDbClient.fetchTeamSchema(testKey);
        try {
            redisDbClient2.storeValue(testKey, newTeam1);
        } catch (RuntimeException e) {
            fail();
        }

        try {
            redisDbClient.storeValue(testKey, newTeam2);
        } catch (RuntimeException e) {
            //Should be throwing an error.
        }

        fetchedTeam = redisDbClient.fetchTeamSchema(testKey);
        assertTeamsEqual(fetchedTeam, newTeam1);
    }

    @Test
    void redisDbClientFetchAndStoreWithDifferentKeysSucceeds() {
        //Because transactions are essentially processed in an async way, if a fetch happens on key x and a store happens on key y,
        //but there was a write on key x (from a different client before the store on key y), the store on key y should not fail.
        var testKey1 = "transaction-test-key1";
        var testKey2 = "transaction-test-key2";
        var redisDbClient2 = new RedisDbClient(redisUrl, objectMapper);
        var newTeam1 = new TeamSchemaImpl("team1", TeamSchema.ROOT_TEAM, "test_organization");
        var newTeam2 = new TeamSchemaImpl("team2", TeamSchema.ROOT_TEAM, "test_organization");

        var fetchedTeam = redisDbClient.fetchTeamSchema(testKey1);
        try {
            redisDbClient2.storeValue(testKey1, newTeam1);
        } catch (RuntimeException e) {
            fail();
        }

        try {
            redisDbClient.storeValue(testKey2, newTeam2);
        } catch (RuntimeException e) {
            fail();
        }

        fetchedTeam = redisDbClient.fetchTeamSchema(testKey2);
        assertTeamsEqual(fetchedTeam, newTeam2);
    }

    void assertTeamsEqual(TeamSchema teamSchema1, TeamSchema teamSchema2) {
        assertEquals(teamSchema1.getTeamName(), teamSchema2.getTeamName());
        assertEquals(teamSchema1.getParentTeam(), teamSchema2.getParentTeam());
        assertEquals(teamSchema1.getOrgName(), teamSchema2.getOrgName());
        var teamSchema1Pipelines = teamSchema1.getPipelineSchemas();
        var teamSchema2Pipelines = teamSchema2.getPipelineSchemas();
        assertEquals(teamSchema1Pipelines.size(), teamSchema2Pipelines.size());
        for (int i = 0; i < teamSchema1Pipelines.size(); i++) {
            assertPipelinesEqual(teamSchema1Pipelines.get(i), teamSchema2Pipelines.get(i));
        }
    }

    void assertPipelinesEqual(PipelineSchema pipelineSchema1, PipelineSchema pipelineSchema2) {
        assertEquals(pipelineSchema1.getPipelineName(), pipelineSchema2.getPipelineName());
        var pipelineSchema1GitRepo = pipelineSchema1.getGitRepoSchema();
        var pipelineSchema2GitRepo = pipelineSchema2.getGitRepoSchema();
        assertEquals(pipelineSchema1GitRepo.getGitRepo(), pipelineSchema2GitRepo.getGitRepo());
        assertEquals(pipelineSchema1GitRepo.getPathToRoot(), pipelineSchema2GitRepo.getPathToRoot());
        //This does not test GitCred. For complete serialization/deserialization testing, see the ObjectMapper test class.
    }
}
