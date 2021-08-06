package com.greenops.util.datamodel.git;

import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo;

@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, include = JsonTypeInfo.As.PROPERTY, property = "type")
@JsonSubTypes(
        {
                @JsonSubTypes.Type(value = GitCredMachineUser.class, name = "machineuser"),
                @JsonSubTypes.Type(value = GitCredOpen.class, name = "open"),
                @JsonSubTypes.Type(value = GitCredToken.class, name = "oauth")
        }
)
public interface GitCredAccessible {
    static String SECURE_GIT_URL_PREFIX = "https://";

    //This is supposed to turn it into a string for executing a command via CLI
    public String convertGitCredToString(String gitRepoLink);

    public void hide();
}
