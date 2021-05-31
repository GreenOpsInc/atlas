package com.greenops.workflowtrigger.api.model.git;

import com.fasterxml.jackson.databind.annotation.JsonDeserialize;

@JsonDeserialize(as = GitCredOpen.class)
public class GitCredOpen implements GitCred {}
