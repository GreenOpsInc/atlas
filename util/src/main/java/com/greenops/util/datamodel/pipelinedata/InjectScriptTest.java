package com.greenops.util.datamodel.pipelinedata;

import com.greenops.util.error.AtlasNonRetryableError;
import com.greenops.util.datamodel.request.KubernetesCreationRequest;

import java.util.List;
import java.util.Map;

import static com.greenops.util.ingest.ClientKey.makeTestKey;
import static com.greenops.util.ingest.deployment.SchemaHandlingUtil.getFileName;

public class InjectScriptTest implements Test {
    public static final String WATCH_TEST_KEY = "WatchTestKey";
    private static final String DEFAULT_NAMESPACE = "default";

    private final String path;
    private final String image;
    private final String namespace;
    private final List<String> commands;
    private final List<String> arguments;
    private final boolean executeInApplicationPod;
    private final boolean executeBeforeDeployment;
    private final Map<String, String> variables;


    InjectScriptTest(String path, String image, String namespace, List<String> commands, List<String> arguments, boolean executeInApplicationPod, boolean executeBeforeDeployment, Map<String, String> variables) {
        this.path = path;
        this.image = image;
        this.namespace = namespace;
        this.commands = commands;
        this.arguments = arguments;
        this.executeInApplicationPod = executeInApplicationPod;
        this.executeBeforeDeployment = executeBeforeDeployment;
        this.variables = variables;
    }

    @Override
    public String getPath() {
        return path;
    }

    public String getImage() {
        return image;
    }

    public String getNamespace() {
        return namespace;
    }

    @Override
    public boolean shouldExecuteBefore() {
        return executeBeforeDeployment;
    }

    @Override
    public Map<String, String> getVariables() {
        return variables;
    }

    @Override
    public Object getPayload(int testNumber, String testConfig) {
        var filename = getFileName(getPath());
        var specifiedImage = getImage();
        var imageName = "";
        if (specifiedImage != null && !specifiedImage.isEmpty()) {
            imageName = specifiedImage;
        } else {
            throw new AtlasNonRetryableError("Image name for InjectScriptTest should not be empty.");
        }
        var specifiedNamespace = getNamespace();
        var jobNamespace = DEFAULT_NAMESPACE;
        if (specifiedNamespace != null && !specifiedNamespace.isBlank()) {
            jobNamespace = specifiedNamespace;
        }
        return new KubernetesCreationRequest(
                INJECT_TASK,
                "Job",
                makeTestKey(testNumber),
                jobNamespace,
                imageName,
                getCommands(),
                getArguments(),
                filename,
                testConfig,
                getVariables()
        );
    }

    @Override
    public String getWatchKey() {
        return WATCH_TEST_KEY;
    }

    private boolean shouldExecuteInPod() {
        return executeInApplicationPod;
    }

    public List<String> getCommands() {
        return commands;
    }

    public List<String> getArguments() {
        return arguments;
    }
}
