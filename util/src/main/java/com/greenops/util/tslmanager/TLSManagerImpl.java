package com.greenops.util.tslmanager;

import com.greenops.util.kubernetesclient.KubernetesClient;
import io.kubernetes.client.models.V1Secret;
import lombok.extern.slf4j.Slf4j;
import org.apache.tomcat.util.net.SSLHostConfig;
import org.apache.tomcat.util.net.SSLHostConfigCertificate;
import org.bouncycastle.jce.X509Principal;
import org.bouncycastle.x509.X509V3CertificateGenerator;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.math.BigInteger;
import java.security.*;
import java.security.cert.CertificateException;
import java.security.cert.X509Certificate;
import java.security.spec.InvalidKeySpecException;
import java.security.spec.X509EncodedKeySpec;
import java.util.Date;

@Slf4j
@Component
public class TLSManagerImpl implements TLSManager {

    private static final String CERTIFICATE_ALGORITHM = "RSA";
    private static final String CERTIFICATE_DN = "CN=Atlas, O=Atlas, ST=SF, C=US";
    private static final int CERTIFICATE_BITS = 1024;
    private static final String NAMESPACE = "default";
    private static final String SECRET_CERT_NAME = "tls.crt";
    private static final String SECRET_KEY_NAME = "tls.key";

    private final String serverCertificateAlias;
    private final String serverCertificateName;
    private final KubernetesClient kclient;

    @Autowired
    public TLSManagerImpl(KubernetesClient kclient, String serverCertificateAlias, String serverCertificateName) {
        this.kclient = kclient;
        this.serverCertificateAlias = serverCertificateAlias;
        this.serverCertificateName = serverCertificateName;
    }

    @Override
    public SSLHostConfig getSSLHostConfig(ClientName serverName) throws Exception {
        SSLHostConfig conf = getTLSConfFromSecrets(serverName);
        if (conf != null) {
            return conf;
        }
        return getSelfSignedSSLHostConf();
    }

    @Override
    public void watchHostSSLConfig(ClientName serverName) {
        ClientSecretName secretName = secretNameFromClientName(serverName);
        this.kclient.watchSecretData(secretName.toString(), NAMESPACE, data -> {
            // TODO: log message about certificate changes and app reloading
            System.exit(0);
        });
    }

    private SSLHostConfig getTLSConfFromSecrets(ClientName serverName) throws Exception {
        ClientSecretName secretName = secretNameFromClientName(serverName);
        V1Secret secret = this.kclient.fetchSecretData(secretName.toString(), NAMESPACE);
        if (secret == null) {
            return null;
        }
        return generateSSLHostConfFromKeyPair(secret.getData().get(SECRET_CERT_NAME), secret.getData().get(SECRET_KEY_NAME));
    }

    private SSLHostConfig generateSSLHostConfFromKeyPair(byte[] pub, byte[] key) throws Exception {
        KeyStore keyStore = getKeyStore();

        KeyPair keyPair = createKeyPair(pub, key);
        X509Certificate cert = createCertificate(keyPair);
        saveCert(cert, keyPair.getPrivate(), keyStore);

        SSLHostConfig sslHostConfig = new SSLHostConfig();
        SSLHostConfigCertificate certificate = new SSLHostConfigCertificate(
                sslHostConfig, SSLHostConfigCertificate.Type.RSA);
        certificate.setCertificateKeystore(keyStore);
        sslHostConfig.addCertificate(certificate);
        return sslHostConfig;
    }

    private SSLHostConfig getSelfSignedSSLHostConf() throws Exception {
        KeyStore keyStore = getKeyStore();

        KeyPair keyPair = generateKeyPair();
        X509Certificate cert = createCertificate(keyPair);
        saveCert(cert, keyPair.getPrivate(), keyStore);

        SSLHostConfig sslHostConfig = new SSLHostConfig();
        SSLHostConfigCertificate certificate = new SSLHostConfigCertificate(
                sslHostConfig, SSLHostConfigCertificate.Type.RSA);
        certificate.setCertificateKeystore(keyStore);
        sslHostConfig.addCertificate(certificate);
        sslHostConfig.setInsecureRenegotiation(true);
        return sslHostConfig;
    }

    private KeyPair createKeyPair(byte[] pub, byte[] key) throws NoSuchAlgorithmException, InvalidKeySpecException {
        PublicKey pubKey = publicKeyFromString(pub);
        PrivateKey privKey = privateKeyFromString(key);
        return new KeyPair(pubKey, privKey);
    }

    private PublicKey publicKeyFromString(byte[] val) throws NoSuchAlgorithmException, InvalidKeySpecException {
        X509EncodedKeySpec x509publicKey = new X509EncodedKeySpec(val);
        KeyFactory kf = KeyFactory.getInstance("RSA");
        return kf.generatePublic(x509publicKey);
    }

    private PrivateKey privateKeyFromString(byte[] val) throws NoSuchAlgorithmException, InvalidKeySpecException {
        X509EncodedKeySpec x509privateKey = new X509EncodedKeySpec(val);
        KeyFactory kf = KeyFactory.getInstance("RSA");
        return kf.generatePrivate(x509privateKey);
    }

    private KeyPair generateKeyPair() throws Exception {
        KeyPairGenerator keyPairGenerator = KeyPairGenerator.getInstance(CERTIFICATE_ALGORITHM);
        keyPairGenerator.initialize(CERTIFICATE_BITS, new SecureRandom());
        return keyPairGenerator.generateKeyPair();
    }

    @SuppressWarnings("deprecation")
    private X509Certificate createCertificate(KeyPair keyPair) throws InvalidKeyException, SignatureException {
        X509V3CertificateGenerator v3CertGen = new X509V3CertificateGenerator();
        v3CertGen.setSerialNumber(BigInteger.valueOf(System.currentTimeMillis()));
        v3CertGen.setIssuerDN(new X509Principal(CERTIFICATE_DN));
        v3CertGen.setNotBefore(new Date(System.currentTimeMillis() - 1000L * 60 * 60 * 24));
        v3CertGen.setNotAfter(new Date(System.currentTimeMillis() + (1000L * 60 * 60 * 24 * 365 * 10)));
        v3CertGen.setSubjectDN(new X509Principal(CERTIFICATE_DN));
        v3CertGen.setPublicKey(keyPair.getPublic());
        v3CertGen.setSignatureAlgorithm("SHA256WithRSAEncryption");
        return v3CertGen.generateX509Certificate(keyPair.getPrivate());
    }

    private void saveCert(X509Certificate cert, PrivateKey key, KeyStore keyStore) throws KeyStoreException, NoSuchAlgorithmException, IOException, CertificateException {
        keyStore.setKeyEntry(serverCertificateAlias, key, "YOUR_PASSWORD".toCharArray(), new java.security.cert.Certificate[]{cert});
        File file = new File(".", serverCertificateName);
        keyStore.store(new FileOutputStream(file), "YOUR_PASSWORD".toCharArray());
    }

    private KeyStore getKeyStore() throws Exception {
        KeyStore keyStore = KeyStore.getInstance("PKCS12");
        keyStore.load(null, null);
        return keyStore;
    }

    private ClientSecretName secretNameFromClientName(ClientName clientName) {
        switch (clientName) {
            case CLIENT_REPO_SERVER:
                return ClientSecretName.REPO_SERVER_SECRET_NAME;
            case CLIENT_WORKFLOW_TRIGGER:
                return ClientSecretName.WORKFLOW_TRIGGER_SECRET_NAME;
            case CLIENT_CLIENT_WRAPPER:
                return ClientSecretName.CLIENT_WRAPPER_SECRET_NAME;
            case CLIENT_COMMAND_DELEGATOR:
                return ClientSecretName.COMMAND_DELEGATOR_SECRET_NAME;
            case CLIENT_ARGOCD_REPO_SERVER:
                return ClientSecretName.ARGOCD_REPO_SERVER_SECRET_NAME;
            default:
                return ClientSecretName.NOT_VALID_SECRET_NAME;
        }
    }
}
