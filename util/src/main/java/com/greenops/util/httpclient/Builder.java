package com.greenops.util.httpclient;

import com.greenops.util.kubernetesclient.KubernetesClient;
import io.kubernetes.client.models.V1Secret;
import lombok.extern.slf4j.Slf4j;
import org.apache.http.client.HttpClient;
import org.apache.http.impl.client.HttpClientBuilder;
import org.springframework.stereotype.Component;

import javax.net.ssl.*;
import java.io.ByteArrayInputStream;
import java.security.KeyStore;
import java.security.SecureRandom;
import java.security.cert.CertificateException;
import java.security.cert.CertificateFactory;
import java.security.cert.X509Certificate;


@Slf4j
@Component
public class Builder {

    private static final String ATLAS_NAMESPACE = "atlas";
    private static final String TLS_CERT_SECRET_FIELD_NAME = "tls.crt";

    private final HttpClientBuilder builder;
    private final KubernetesClient kubernetesClient;
    private SSLContext sslContext;

    private Builder(KubernetesClient kubernetesClient) {
        this.kubernetesClient=kubernetesClient;
        this.builder = HttpClientBuilder.create();
    }

    public static Builder create(KubernetesClient kubernetesClient) {
        return new Builder(kubernetesClient);
    }

    public Builder withCustomTls(String clientTlsSecretName) throws Exception {
        SSLContext context = SSLContext.getInstance("TLS");
        V1Secret tlsSecret = this.kubernetesClient.fetchSecretData(ATLAS_NAMESPACE, clientTlsSecretName);
        if (tlsSecret == null) {
            return this;
        }

        byte[] certBytes = tlsSecret.getData().get(TLS_CERT_SECRET_FIELD_NAME);
        X509Certificate cert = generateCertificateFromDER(certBytes);

        KeyStore keystore = KeyStore.getInstance("JKS");
        keystore.load(null);
        keystore.setCertificateEntry("alias", cert);

        KeyManagerFactory kmf = KeyManagerFactory.getInstance("SunX509");
        kmf.init(keystore, "password".toCharArray());
        KeyManager[] km = kmf.getKeyManagers();
        context.init(km, null, null);
        sslContext = context;
        return this;
    }

    public HttpClient build() throws Exception {
        if (this.sslContext == null) {
            this.sslContext = SSLContext.getInstance("TLS");
            this.sslContext.init(null, trustAllCerts, new SecureRandom());
            this.builder.setSSLHostnameVerifier((s1, s2) -> true);
        }
        return this.builder.setSSLContext(this.sslContext).build();
    }

    private static X509Certificate generateCertificateFromDER(byte[] certBytes) throws CertificateException {
        final CertificateFactory factory = CertificateFactory.getInstance("X.509");
        return (X509Certificate) factory.generateCertificate(new ByteArrayInputStream(certBytes));
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

}
