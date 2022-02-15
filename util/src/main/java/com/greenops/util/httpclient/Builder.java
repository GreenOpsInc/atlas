package com.greenops.util.httpclient;

import lombok.extern.slf4j.Slf4j;
import org.apache.http.client.HttpClient;
import org.apache.http.conn.ssl.NoopHostnameVerifier;
import org.apache.http.conn.ssl.SSLConnectionSocketFactory;
import org.apache.http.impl.client.HttpClientBuilder;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.ssl.SSLContextBuilder;
import org.apache.http.ssl.SSLContexts;
import org.springframework.stereotype.Component;

import javax.net.ssl.SSLContext;
import javax.net.ssl.TrustManager;
import javax.net.ssl.X509TrustManager;
import java.io.File;
import java.security.SecureRandom;

@Slf4j
@Component
public class Builder {
    private static final String KEYSTORE_CERT_ALIAS = "reposerver.atlas.svc.cluster.local";
    private SSLConnectionSocketFactory sslConSocFactory;
    private HttpClientBuilder builder;

    private Builder() {
        this.builder = HttpClients.custom();
    }

    public static Builder create() {
        return new Builder();
    }

    public Builder withCustomTls(String keystorePath, String keystorePassword, String derCertPath, String caCertsPath, String caCertsKeystorePassword) throws Exception {
        if (keystorePath.equals("") || derCertPath.equals("") || caCertsPath.equals("")) {
            return this;
        }
        File keystoreFile = new File(keystorePath);
        File derCertFile = new File(derCertPath);
        File caCertsFile = new File(caCertsPath);
        if (!keystoreFile.exists() || !derCertFile.exists() || !caCertsFile.exists()) {
            return this;
        }

        SSLContextBuilder SSLBuilder = SSLContexts.custom();
        File file = new File(keystorePath);
        SSLBuilder = SSLBuilder.loadTrustMaterial(file, keystorePassword.toCharArray());
        SSLContext sslcontext = SSLBuilder.build();
        this.sslConSocFactory = new SSLConnectionSocketFactory(sslcontext, new NoopHostnameVerifier());
        this.addCertToJDK(derCertPath, caCertsPath, caCertsKeystorePassword);
        return this;
    }

    public HttpClient build() throws Exception {
        if (this.sslConSocFactory == null) {
            SSLContext context = SSLContext.getInstance("TLS");
            context.init(null, trustAllCerts, new SecureRandom());
            this.builder.setSSLHostnameVerifier((s1, s2) -> true);
            return this.builder.setSSLContext(context).build();
        }
        return this.builder.setSSLSocketFactory(sslConSocFactory).build();
    }

    private static final TrustManager[] trustAllCerts = new TrustManager[]{
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

    private void addCertToJDK(String derCertPath, String caCertsPath, String caCertsKeystorePassword) throws Exception {
        String cmd = "keytool -trustcacerts -import -alias " + KEYSTORE_CERT_ALIAS + " -keystore " + caCertsPath + " -file " + derCertPath + " -storepass " + caCertsKeystorePassword + " -noprompt";
        System.out.println("Running command to add der cert to ca certs: " + cmd);
        Runtime run = Runtime.getRuntime();
        Process pr = run.exec(cmd);
        pr.waitFor();
    }

}
