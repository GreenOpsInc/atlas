package com.greenops.util.tslmanager;

import org.apache.tomcat.util.net.SSLHostConfig;

import java.util.Map;

public interface TLSManager {
    SSLHostConfig getSSLHostConfig(ClientName serverName) throws Exception;

    void watchHostSSLConfig(ClientName serverName);

    Map<String, Object> getKafkaSSLConfProps(String keystoreLocation, String trueStoreLocation) throws Exception;

    void watchKafkaKeystore(String keystoreLocation, String trueStoreLocation);
}
