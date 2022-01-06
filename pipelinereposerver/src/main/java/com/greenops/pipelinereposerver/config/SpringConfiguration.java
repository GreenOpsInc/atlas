package com.greenops.pipelinereposerver.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.util.datamodel.git.GitCredMachineUser;
import com.greenops.util.datamodel.git.GitCredToken;
import com.greenops.util.datamodel.git.GitRepoSchema;
import com.greenops.util.datamodel.mixin.git.GitCredMachineUserMixin;
import com.greenops.util.datamodel.mixin.git.GitCredTokenMixin;
import com.greenops.util.datamodel.mixin.git.GitRepoSchemaMixin;
import com.greenops.util.datamodel.mixin.pipeline.PipelineSchemaMixin;
import com.greenops.util.datamodel.mixin.pipeline.TeamSchemaMixin;
import com.greenops.util.datamodel.mixin.request.GetFileRequestMixin;
import com.greenops.util.datamodel.pipeline.PipelineSchemaImpl;
import com.greenops.util.datamodel.pipeline.TeamSchemaImpl;
import com.greenops.util.datamodel.request.GetFileRequest;
import com.greenops.util.dbclient.DbClient;
import com.greenops.util.dbclient.redis.RedisDbClient;
import com.greenops.util.kubernetesclient.KubernetesClient;
import com.greenops.util.kubernetesclient.KubernetesClientImpl;
import com.greenops.util.tslmanager.ClientName;
import com.greenops.util.tslmanager.TLSManager;
import com.greenops.util.tslmanager.TLSManagerImpl;
import org.apache.catalina.Context;
import org.apache.catalina.connector.Connector;
import org.apache.tomcat.util.descriptor.web.SecurityCollection;
import org.apache.tomcat.util.descriptor.web.SecurityConstraint;
import org.apache.tomcat.util.net.SSLHostConfig;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory;
import org.springframework.boot.web.servlet.server.ServletWebServerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.io.IOException;

@Configuration
public class SpringConfiguration {

    @Bean
    ObjectMapper objectMapper() {
        return new ObjectMapper()
                .addMixIn(TeamSchemaImpl.class, TeamSchemaMixin.class)
                .addMixIn(PipelineSchemaImpl.class, PipelineSchemaMixin.class)
                .addMixIn(GitRepoSchema.class, GitRepoSchemaMixin.class)
                .addMixIn(GitCredMachineUser.class, GitCredMachineUserMixin.class)
                .addMixIn(GitCredToken.class, GitCredTokenMixin.class)
                .addMixIn(GetFileRequest.class, GetFileRequestMixin.class);
    }

    @Bean
    DbClient dbClient(@Value("${application.redis-url}") String redisUrl, ObjectMapper objectMapper) {
        return new RedisDbClient(redisUrl, objectMapper);
    }

    @Bean
    public ServletWebServerFactory servletContainer(ObjectMapper objectMapper) {
        TomcatServletWebServerFactory tomcat = new TomcatServletWebServerFactory() {
            @Override
            protected void postProcessContext(Context context) {
                var securityConstraint = new SecurityConstraint();
                securityConstraint.setUserConstraint("CONFIDENTIAL");
                var collection = new SecurityCollection();
                collection.addPattern("/*");
                securityConstraint.addCollection(collection);
                context.addConstraint(securityConstraint);
            }
        };
        tomcat.addAdditionalTomcatConnectors(getHttpConnector(objectMapper));
        return tomcat;
    }

    private Connector getHttpConnector(ObjectMapper objectMapper) {
        var connector = new Connector(TomcatServletWebServerFactory.DEFAULT_PROTOCOL);
        connector.setScheme("https");
        connector.setPort(8080);
        connector.setSecure(true);

        KubernetesClient kclient;
        try {
            kclient = new KubernetesClientImpl(objectMapper);
        } catch (IOException exc) {
            throw new RuntimeException("Could not initialize Kubernetes Client", exc);
        }

        SSLHostConfig conf;
        try {
            TLSManager tlsManager = new TLSManagerImpl(kclient, "pipelinereposerver_tls_cert", "keystore.pipelinereposerver_tls_cert");
            conf = tlsManager.getSSLHostConfig(ClientName.CLIENT_CLIENT_WRAPPER);
            tlsManager.watchHostSSLConfig(ClientName.CLIENT_CLIENT_WRAPPER);
        } catch (Exception exc) {
            throw new RuntimeException("Could not configure server with TLS configuration", exc);
        }

        connector.addSslHostConfig(conf);
        return connector;
    }
}
