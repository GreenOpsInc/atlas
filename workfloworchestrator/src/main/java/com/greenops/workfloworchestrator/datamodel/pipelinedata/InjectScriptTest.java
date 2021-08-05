package com.greenops.workfloworchestrator.datamodel.pipelinedata;

import com.greenops.workfloworchestrator.datamodel.requests.KubernetesCreationRequest;
import com.greenops.workfloworchestrator.ingest.handling.testautomation.CommandBuilder;

import java.util.List;
import java.util.Map;

import static com.greenops.workfloworchestrator.ingest.handling.ClientKey.makeTestKey;
import static com.greenops.workfloworchestrator.ingest.handling.util.deployment.SchemaHandlingUtil.escapeFile;
import static com.greenops.workfloworchestrator.ingest.handling.util.deployment.SchemaHandlingUtil.getFileName;

public class InjectScriptTest implements Test {

    private String path;
    private boolean executeInApplicationPod;
    private boolean executeBeforeDeployment;
    private Map<String, String> variables;
    private String image;

    InjectScriptTest(String path, String image, boolean executeInApplicationPod, boolean executeBeforeDeployment, Map<String, String> variables) {
        this.path = path;
        this.image = image;
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
            // TODO Add logic for image auto detection based on shebang here
            imageName = "";
        }
        return new KubernetesCreationRequest(
                "Job",
                makeTestKey(testNumber),
                "",
                imageName,
                List.of("/bin/sh", "-c"),
                new CommandBuilder().createFile(filename, escapeFile(testConfig)).compile(filename).executeExistingFile(filename).build(),
                getVariables()
        );
    }

    private boolean shouldExecuteInPod() {
        return executeInApplicationPod;
    }
}
