package com.greenops.util.tslmanager;

import org.apache.tomcat.util.net.SSLHostConfig;

public interface TLSManager {
    SSLHostConfig getSSLHostConfig(ClientName serverName) throws Exception;

    void watchHostSSLConfig(ClientName serverName);

    boolean updateKafkaKeystore(String keystoreLocation, String trueStoreLocation) throws Exception;

    void watchKafkaKeystore(String keystoreLocation, String trueStoreLocation);
}
