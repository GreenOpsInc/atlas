package com.greenops.workflowtrigger.api.argoauthenticator;

import account.AccountOuterClass;
import account.AccountServiceGrpc;
import com.greenops.util.error.AtlasAuthenticationError;
import io.grpc.Channel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.StatusRuntimeException;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

@Slf4j
@Component
public class ArgoAuthenticatorApiImpl implements ArgoAuthenticatorApi {

    private final AccountServiceGrpc.AccountServiceBlockingStub blockingStub;
    private final AccountServiceGrpc.AccountServiceStub asyncStub;
    private Channel channel;

    @Autowired
    public ArgoAuthenticatorApiImpl(@Value("${application.argocd-server-url}") String serverEndpoint) {
        this(ManagedChannelBuilder.forTarget(serverEndpoint).usePlaintext());
    }

    /** Construct client for accessing RouteGuide server using the existing channel. */
    public ArgoAuthenticatorApiImpl(ManagedChannelBuilder<?> channelBuilder) {
        channel = channelBuilder.build();
        blockingStub = AccountServiceGrpc.newBlockingStub(channel);
        asyncStub = AccountServiceGrpc.newStub(channel);
    }

    @Override
    public boolean checkRbacPermissions(String action, String resource, String subresource) {
        var canIRequest = AccountOuterClass.CanIRequest.newBuilder()
                .setAction(action)
                .setResource(resource)
                .setSubresource(subresource)
                .build();

        try {
            var canIResponse = blockingStub.canI(canIRequest);
            return canIResponse.getValue().equals("yes");
        } catch (StatusRuntimeException e) {
            log.info("RPC failed: {}", e.getStatus());
            if (e.getStatus().getCode().value() == 16) {
                throw new AtlasAuthenticationError(e.getMessage());
            }
            throw new RuntimeException(e);
        }
    }
}
