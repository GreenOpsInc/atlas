package com.greenops.util.httpclient;

import lombok.extern.slf4j.Slf4j;
import nl.altindag.ssl.SSLFactory;
import nl.altindag.ssl.util.PemUtils;
import org.apache.http.client.HttpClient;
import org.apache.http.impl.client.HttpClientBuilder;
import org.apache.http.impl.client.HttpClients;
import org.springframework.stereotype.Component;

import javax.net.ssl.SSLContext;
import java.io.File;
import java.nio.file.Paths;

@Slf4j
@Component
public class Builder {
    private static final String KEYSTORE_CERT_ALIAS = "reposerver.atlas.svc.cluster.local";
    private SSLContext sslCon;
    private HttpClientBuilder builder;

    private Builder() {
        this.builder = HttpClients.custom();
    }

    public static Builder create() {
        return new Builder();
    }

    public Builder withCustomTls(String certPemPath, String keyPemPath) {
        if (certPemPath.equals("") || keyPemPath.equals("")) {
            return this;
        }
        File certPemFile = new File(certPemPath);
        File keyPemFile = new File(keyPemPath);
        if (!certPemFile.exists() || !keyPemFile.exists()) {
            log.info("Paths are not valid or files do not exist, not adding custom TLS");
            return this;
        }

        var keyManager = PemUtils.loadIdentityMaterial(Paths.get(certPemPath), Paths.get(keyPemPath));
        var trustManager = PemUtils.loadTrustMaterial(Paths.get(certPemPath));
        var sslFactory = SSLFactory.builder()
                .withIdentityMaterial(keyManager)
                .withTrustMaterial(trustManager)
                .build();
        this.sslCon = sslFactory.getSslContext();
        return this;
    }

    public HttpClient build() {
        if (this.sslCon == null) {
            var sslFactory = SSLFactory.builder()
                    .withUnsafeTrustMaterial()
                    .withUnsafeHostnameVerifier()
                    .build();
            this.builder.setSSLHostnameVerifier(sslFactory.getHostnameVerifier());
            return this.builder.setSSLContext(sslFactory.getSslContext()).build();
        }
        return this.builder.setSSLContext(this.sslCon).build();
    }

}
