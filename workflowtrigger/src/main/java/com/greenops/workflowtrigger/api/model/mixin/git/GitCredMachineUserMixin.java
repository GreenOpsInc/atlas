package com.greenops.workflowtrigger.api.model.mixin.git;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

public abstract class GitCredMachineUserMixin {
    @JsonProperty(value = "username")
    String username;
    @JsonProperty(value = "password")
    String password;

    @JsonCreator
    public GitCredMachineUserMixin(@JsonProperty(value = "username") String username, @JsonProperty(value = "password") String password) {}
}
