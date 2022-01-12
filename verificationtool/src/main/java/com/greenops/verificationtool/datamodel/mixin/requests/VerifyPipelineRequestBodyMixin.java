package com.greenops.verificationtool.datamodel.mixin.requests;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.greenops.verificationtool.datamodel.rules.RuleData;

import java.util.List;

public abstract class VerifyPipelineRequestBodyMixin {
    @JsonProperty(value = "stepName")
    String stepName;
    @JsonProperty(value = "gitRepoUrl")
    String gitRepoUrl;
    @JsonProperty(value = "rules")
    List<RuleData> rules;

    @JsonCreator
    public VerifyPipelineRequestBodyMixin(@JsonProperty("gitRepoUrl") String gitRepoUrl,
                                          @JsonProperty("stepName") String stepName,
                                          @JsonProperty("rules") List<RuleData> rules) {
    }
}
