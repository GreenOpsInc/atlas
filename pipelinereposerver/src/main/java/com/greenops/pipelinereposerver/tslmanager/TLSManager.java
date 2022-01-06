package com.greenops.pipelinereposerver.tslmanager;

import org.apache.tomcat.util.net.SSLHostConfig;

public interface TLSManager {
    SSLHostConfig getSSLHostConfig() throws Exception;

    void watchHostSSLConfig();
}
