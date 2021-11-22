package com.greenops.workflowtrigger.api.argoauthenticator;

import account.AccountOuterClass;
import account.AccountServiceGrpc;
import com.greenops.util.error.AtlasAuthenticationError;
import io.grpc.*;
import io.grpc.netty.shaded.io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.shaded.io.grpc.netty.NegotiationType;
import io.grpc.netty.shaded.io.grpc.netty.NettyChannelBuilder;
import io.grpc.netty.shaded.io.netty.handler.ssl.ApplicationProtocolConfig;
import io.grpc.netty.shaded.io.netty.handler.ssl.SslContextBuilder;
import io.grpc.netty.shaded.io.netty.handler.ssl.util.InsecureTrustManagerFactory;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import javax.net.ssl.SSLException;
import javax.net.ssl.TrustManager;
import javax.net.ssl.X509TrustManager;
import java.nio.file.Files;
import java.nio.file.Paths;
import java.security.KeyManagementException;
import java.security.NoSuchAlgorithmException;
import java.util.Arrays;
import java.util.Collections;

@Slf4j
@Component
public class ArgoAuthenticatorApiImpl implements ArgoAuthenticatorApi {

    private final AccountServiceGrpc.AccountServiceBlockingStub blockingStub;
    private final AccountServiceGrpc.AccountServiceStub asyncStub;
    private Channel channel;

    private static TrustManager[] trustAllCerts = new TrustManager[]{
            new X509TrustManager() {
                public java.security.cert.X509Certificate[] getAcceptedIssuers() {
                    return null;
                }
                public void checkClientTrusted(
                        java.security.cert.X509Certificate[] certs, String authType) {
                }
                public void checkServerTrusted(
                        java.security.cert.X509Certificate[] certs, String authType) {
                }
            }
    };

    @Autowired
    public ArgoAuthenticatorApiImpl(@Value("${application.argocd-server-url}") String serverEndpoint) throws SSLException, NoSuchAlgorithmException, KeyManagementException {
//        SSLContext sslContext = SslContext..getInstance("TLS");
//        GrpcSslContexts.forClient().trustManager(trustAllCerts[0]).build()
//        sslContext.init(null, trustAllCerts, new SecureRandom());
        this(Grpc.newChannelBuilderForAddress(serverEndpoint, 8080, TlsChannelCredentials.newBuilder().trustManager(InsecureTrustManagerFactory.INSTANCE.getTrustManagers()).build()));
//
//                        forAddress(serverEndpoint, 8080)
////                .negotiationType(NegotiationType.TLS)
//                .overrideAuthority("localhost")
//                .sslContext(
//                        GrpcSslContexts.forClient()
////                                .applicationProtocolConfig(new ApplicationProtocolConfig(
////                                        ApplicationProtocolConfig.Protocol.ALPN,
////                                        ApplicationProtocolConfig.SelectorFailureBehavior.NO_ADVERTISE,
////                                        ApplicationProtocolConfig.SelectedListenerFailureBehavior.ACCEPT,
////                                        Collections.unmodifiableList(Arrays.asList("grpc-exp", "h2"))))
////                                .startTls(true)
//                                .trustManager(InsecureTrustManagerFactory.INSTANCE).build()
//                )
//        );// GrpcSslContexts.forClient().trustManager(InsecureTrustManagerFactory.INSTANCE).build()));
////                .sslContext(
////                        GrpcSslContexts
////                                .forClient()
//////                                .trustManager(InsecureTrustManagerFactory.INSTANCE)
////                                .trustManager(TlsTesting.loadCert("/Users/mihirpandya/Desktop/Mihir/GreenOps/atlas/workflowtrigger/src/test/java/com/greenops/workflowtrigger/cert.pem")) // public key
////                                .build())
////                .overrideAuthority("localhost"));
    }

    public void getFile() {
        log.info("{}", Files.exists(Paths.get("/Users/mihirpandya/Desktop/Mihir/GreenOps/atlas/workflowtrigger/src/test/java/com/greenops/workflowtrigger/cert.pem")));
    }

    /**
     * Construct client for accessing RouteGuide server using the existing channel.
     */
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
