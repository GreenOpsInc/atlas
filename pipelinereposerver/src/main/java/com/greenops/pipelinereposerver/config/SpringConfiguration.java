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
        System.out.println("in getHttpConnector");
        var connector = new Connector(TomcatServletWebServerFactory.DEFAULT_PROTOCOL);
        connector.setScheme("https");
        connector.setPort(8080);
        connector.setSecure(true);
        System.out.println("in getHttpConnector before creating k client");

        KubernetesClient kclient;
        try {
            kclient = new KubernetesClientImpl(objectMapper);
            System.out.println("in getHttpConnector created k client");
        } catch (IOException exc) {
            System.out.println("in getHttpConnector k client creation exception");
            throw new RuntimeException("Could not initialize Kubernetes Client", exc);
        }

        System.out.println("in getHttpConnector before initializing tls manager");
        SSLHostConfig conf;
        try {
            TLSManager tlsManager = new TLSManagerImpl(kclient, "pipelinereposerver_tls_cert", "keystore.pipelinereposerver_tls_cert");
            System.out.println("in getHttpConnector initialized tls manager");
            conf = tlsManager.getSSLHostConfig(ClientName.CLIENT_CLIENT_WRAPPER);
            System.out.println("in getHttpConnector retrieved tls config, conf certificate = " + conf.getCaCertificateFile());
            tlsManager.watchHostSSLConfig(ClientName.CLIENT_CLIENT_WRAPPER);
            System.out.println("in getHttpConnector tls manager watcher started");
        } catch (Exception exc) {
            System.out.println("in getHttpConnector in catch block, exc = " + exc);
            throw new RuntimeException("Could not configure server with TLS configuration", exc);
        }

        connector.addSslHostConfig(conf);
        return connector;
    }
}
