package com.greenops.workflowtrigger.api.argoauthenticator;

public interface ArgoAuthenticatorApi {

    static final String CLUSTER_RESOURCE = "clusters";
    static final String APPLICATION_RESOURCE = "applications";

    static final String SYNC_ACTION = "sync";
    static final String CREATE_ACTION = "create";
    static final String GET_ACTION = "get";
    static final String DELETE_ACTION = "delete";
    static final String ACTION_ACTION = "action";

    public boolean checkRbacPermissions(String action, String resource, String subresource);

}
