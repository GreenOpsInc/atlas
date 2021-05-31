package com.greenops.workflowtrigger.api.model.git;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import org.json.JSONException;
import org.json.JSONObject;

@JsonDeserialize(as = GitCredOpen.class)
public class GitCredOpen implements GitCred {
    @Override
    public JSONObject convertToJson() {
        try {
            return new JSONObject().put("Type", "Unauthenticated");
        } catch (JSONException e) {
            throw new RuntimeException("Unauthenticated credentials could not be converted to JSON object.");
        }
    }
}
