package com.greenops.util.tslmanager;

import java.util.Map;

public interface TLSManager {
    Map<String, Object> getKafkaSSLConfProps(String keystoreLocation, String trueStoreLocation) throws Exception;

    void watchKafkaKeystore(String keystoreLocation, String trueStoreLocation);
}
