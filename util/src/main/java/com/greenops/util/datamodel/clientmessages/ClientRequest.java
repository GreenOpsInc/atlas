package com.greenops.util.datamodel.clientmessages;

import com.fasterxml.jackson.annotation.JsonSubTypes;
import com.fasterxml.jackson.annotation.JsonTypeInfo;

@JsonTypeInfo(use = JsonTypeInfo.Id.NAME, include = JsonTypeInfo.As.PROPERTY, property = "type")
@JsonSubTypes(
        {
                @JsonSubTypes.Type(value = ClientDeleteByConfigRequest.class, name = "del_config"),
                @JsonSubTypes.Type(value = ClientDeleteByGvkRequest.class, name = "del_gvk"),
                @JsonSubTypes.Type(value = ClientDeployAndWatchRequest.class, name = "deploy_watch"),
                @JsonSubTypes.Type(value = ClientDeployNamedArgoAppAndWatchRequest.class, name = "deploy_namedargo_watch"),
                @JsonSubTypes.Type(value = ClientDeployNamedArgoApplicationRequest.class, name = "deploy_namedargo"),
                @JsonSubTypes.Type(value = ClientDeployRequest.class, name = "deploy"),
                @JsonSubTypes.Type(value = ClientRollbackAndWatchRequest.class, name = "rollback"),
                @JsonSubTypes.Type(value = ClientSelectiveSyncAndWatchRequest.class, name = "sel_sync_watch")
        }
)
public interface ClientRequest {
}
