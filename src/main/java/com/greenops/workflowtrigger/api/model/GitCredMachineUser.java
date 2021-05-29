package com.greenops.workflowtrigger.api.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import org.json.JSONException;
import org.json.JSONObject;

@JsonDeserialize(as = GitCredMachineUser.class)
public class GitCredMachineUser implements GitCred {
    @JsonProperty(value = "username")
    private String username;
    @JsonProperty(value = "password")
    private String password;

    public GitCredMachineUser(String username, String password) {
        this.username = username;
        this.password = password;
    }

    @Override
    public JSONObject convertToJson() {
        try {
            return new JSONObject()
                    .put("Username", username)
                    .put("Password", password);
        } catch (JSONException e) {
            throw new RuntimeException("Machine user's Git credentials could not be converted to JSON object.");
        }
    }
}
