package com.greenops.pipelinereposerver.config;

import com.fasterxml.jackson.databind.ObjectMapper;
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
import org.springframework.boot.web.embedded.tomcat.TomcatServletWebServerFactory;
import org.springframework.boot.web.servlet.server.ServletWebServerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

import java.io.IOException;

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
            TLSManager tlsManager = new TLSManagerImpl(kclient, "pipelinereposerver_tls_cert", "keystore.pipelinereposerver_tls_cert");
            conf = tlsManager.getSSLHostConfig(ClientName.CLIENT_CLIENT_WRAPPER);
            tlsManager.watchHostSSLConfig(ClientName.CLIENT_CLIENT_WRAPPER);
        } catch (Exception exc) {
            // TODO: log and stop server
            throw new RuntimeException("Could not configure server with TLS configuration", exc);
        }

        connector.addSslHostConfig(conf);
        return connector;
    }
}
