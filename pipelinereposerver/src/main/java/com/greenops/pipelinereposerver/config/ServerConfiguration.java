package com.greenops.pipelinereposerver.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.greenops.pipelinereposerver.kubernetesclient.KubernetesClient;
import com.greenops.pipelinereposerver.kubernetesclient.KubernetesClientImpl;
import com.greenops.pipelinereposerver.tslmanager.TLSManager;
import com.greenops.pipelinereposerver.tslmanager.TLSManagerImpl;
import org.apache.catalina.Context;
import org.apache.catalina.connector.Connector;
import org.apache.tomcat.util.descriptor.web.SecurityCollection;
import org.apache.tomcat.util.descriptor.web.SecurityConstraint;
import org.apache.tomcat.util.net.SSLHostConfig;
import org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory;
import org.springframework.boot.web.servlet.server.ServletWebServerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.io.IOException;

// TODO: server configuration
//      1. add bean to:
//          a. read config from secrets
//          b. if not available create self signed
//          c. create keystore
//          d. use keystore for tomcat
//      2. create watcher to watch kubernetes secret updates
//          a. on create/update/delete reload the server
//          b. after server is reloaded it will pull new config from secrets and create the server
//          c. the same could be done for kafka (I guess)

@Configuration
public class ServerConfiguration {

    @Bean
    public ServletWebServerFactory servletContainer() {
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
        tomcat.addAdditionalTomcatConnectors(getHttpConnector());
        return tomcat;
    }

    // TODO: call tlsmanager here to get tls cert and apply it to the server
    private Connector getHttpConnector() {
        var connector = new Connector(TomcatServletWebServerFactory.DEFAULT_PROTOCOL);
        connector.setScheme("https");
        connector.setPort(8080);
        connector.setSecure(true);

        // TODO: somehow get kubernetes client from beans
        KubernetesClient kclient;
        try {
            // TODO: somehow get object mapper from beans
            ObjectMapper objectMapper = new ObjectMapper();
            kclient = new KubernetesClientImpl(objectMapper);
        } catch (IOException exc) {
            // TODO: log and stop server
            throw new RuntimeException("Could not initialize Kubernetes Client", exc);
        }

        SSLHostConfig conf;
        try {
            TLSManager tlsManager = new TLSManagerImpl(kclient);
            conf = tlsManager.getSSLHostConfig();
            tlsManager.watchHostSSLConfig();
        } catch (Exception exc) {
            // TODO: log and stop server
            throw new RuntimeException("Could not configure server with TLS configuration", exc);
        }

        connector.addSslHostConfig(conf);
        return connector;
    }
}
