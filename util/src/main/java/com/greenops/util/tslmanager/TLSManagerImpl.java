package com.greenops.util.tslmanager;

import com.greenops.util.kubernetesclient.KubernetesClient;
import io.kubernetes.client.models.V1Secret;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.codec.binary.Base64;
import org.apache.kafka.clients.CommonClientConfigs;
import org.apache.kafka.common.config.SslConfigs;
import org.apache.tomcat.util.net.SSLHostConfig;
import org.apache.tomcat.util.net.SSLHostConfigCertificate;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.security.*;
import java.security.cert.CertificateEncodingException;
import java.security.cert.CertificateException;
import java.security.cert.X509Certificate;
import java.security.spec.InvalidKeySpecException;
import java.security.spec.X509EncodedKeySpec;
import java.util.Arrays;
import java.util.Date;
import java.util.HashMap;
import java.util.Map;

@Slf4j
@Component
public class TLSManagerImpl implements TLSManager {

    private static final String CERTIFICATE_ALGORITHM = "RSA";
    private static final String CERTIFICATE_DN = "CN=Atlas, O=Atlas, ST=SF, C=US";
    private static final int CERTIFICATE_BITS = 1024;
    private static final String NAMESPACE = "default";
    private static final String SECRET_CERT_NAME = "tls.crt";
    private static final String SECRET_KEY_NAME = "tls.key";
    private static final String SECRET_KAFKA_KEYSTORE_NAME = "kafka.keystore.jks";
    private static final String SECRET_KAFKA_TRUSTSTORE_NAME = "kafka.truststore.jks";
    private static final String SECRET_KAFKA_KEY_CREDENTIALS = "kafka.key.credentials";
    private static final String SECRET_KAFKA_KEYSTORE_CREDENTIALS = "kafka.keystore.credentials";
    private static final String SECRET_KAFKA_TRUSTSTORE_CREDENTIALS = "kafka.truststore.credentials";
    private static final String KEYSTORE_PASSWORD = "SS28qmtOJH4OFLUP";
    private static final Date NOT_BEFORE = new Date(System.currentTimeMillis() - 86400000L * 365);
    private static final Date NOT_AFTER = new Date(253402300799000L);

    private final String serverCertificateAlias;
    private final String serverCertificateName;
    private final KubernetesClient kclient;

    @Autowired
    public TLSManagerImpl(KubernetesClient kclient, String serverCertificateAlias, String serverCertificateName) {
        this.kclient = kclient;
        this.serverCertificateAlias = serverCertificateAlias;
        this.serverCertificateName = serverCertificateName;
    }

    protected static String convertToPem(X509Certificate cert) throws CertificateEncodingException {
        Base64 encoder = new Base64(64);
        String cert_begin = "-----BEGIN CERTIFICATE-----\n";
        String end_cert = "-----END CERTIFICATE-----";

        byte[] derCert = cert.getEncoded();
        String pemCertPre = new String(encoder.encode(derCert));
        String pemCert = cert_begin + pemCertPre + end_cert;
        return pemCert;
    }

    @Override
    public SSLHostConfig getSSLHostConfig(ClientName serverName) throws Exception {
        SSLHostConfig conf = getTLSConfFromSecrets(serverName);
        if (conf != null) {
            System.out.println("in getSSLHostConfig ssl host config found in secrets, returning: " + conf.getCertificateFile());
            return conf;
        }
        System.out.println("in getSSLHostConfig ssl host config no found in secrets, generating self signed...");
        return getSelfSignedSSLHostConf();
    }

    @Override
    public void watchHostSSLConfig(ClientName serverName) {
        ClientSecretName secretName = secretNameFromClientName(serverName);
        this.kclient.watchSecretData(secretName.toString(), NAMESPACE, data -> {
            System.out.println("Exiting due to Server certificate change.");
            System.exit(0);
        });
    }

    @Override
    public Map<String, Object> getKafkaSSLConfProps(String keystoreLocation, String truststoreLocation) throws Exception {
        File keyStoreFile = new File(keystoreLocation);
        File truststoreFile = new File(truststoreLocation);
        File parent = keyStoreFile.getParentFile();
        if (parent != null && !parent.exists() && !parent.mkdirs()) {
            throw new IllegalStateException("Couldn't create .atlas config directory: " + parent);
        }

        if (keyStoreFile.exists()) {
            keyStoreFile.delete();
        }
        if (truststoreFile.exists()) {
            truststoreFile.delete();
        }

        ClientSecretName secretName = secretNameFromClientName(ClientName.CLIENT_KAFKA);
        System.out.println("in updateKafkaKeystore keystoreLocation = " + keystoreLocation + " truststoreLocation = " + truststoreLocation + " secretName = " + secretName.toString() + " namespace = " + NAMESPACE);
        V1Secret secret = this.kclient.fetchSecretData(NAMESPACE, secretName.toString());
        if (secret == null) {
            return null;
        }

        System.out.println("secret = ");
        for (String key : secret.getData().keySet()) {
            System.out.println("key : " + key);
            System.out.println("value : " + Arrays.toString(secret.getData().get(key)));
        }

        byte[] keystoreData = secret.getData().get(SECRET_KAFKA_KEYSTORE_NAME);
        byte[] truststoreData = secret.getData().get(SECRET_KAFKA_TRUSTSTORE_NAME);
        String kafkaKeystoreCreds = new String(secret.getData().get(SECRET_KAFKA_KEYSTORE_CREDENTIALS), StandardCharsets.UTF_8);
        String kafkaTruststoreCreds = new String(secret.getData().get(SECRET_KAFKA_TRUSTSTORE_CREDENTIALS), StandardCharsets.UTF_8);
        String kafkaKeyCreds = new String(secret.getData().get(SECRET_KAFKA_KEY_CREDENTIALS), StandardCharsets.UTF_8);

        System.out.println("keystore data = " + new String(keystoreData, StandardCharsets.UTF_8));
        System.out.println("truststore data = " + new String(truststoreData, StandardCharsets.UTF_8));
        System.out.println("keystore creds = '" + kafkaKeystoreCreds + "'");
        System.out.println("truststore creds = '" + kafkaTruststoreCreds + "'");
        System.out.println("key creds = '" + kafkaKeyCreds + "'");

        keyStoreFile = new File(keystoreLocation);
        keyStoreFile.createNewFile();
        Files.write(keyStoreFile.toPath(), keystoreData);

        truststoreFile = new File(truststoreLocation);
        truststoreFile.createNewFile();
        Files.write(truststoreFile.toPath(), truststoreData);

        System.out.println("trying to read keystore and truststore file contents...");

        // TODO: delete
        byte[] keystoreContents = Files.readAllBytes(keyStoreFile.toPath());
        byte[] truststoreContents = Files.readAllBytes(truststoreFile.toPath());
        System.out.println("saved keystore file contents = " + new String(keystoreContents));
        System.out.println("saved truststore file contents = " + new String(truststoreContents));
        // TODO: end delete

        Map<String, Object> configProps = new HashMap<>();
        configProps.put(CommonClientConfigs.SECURITY_PROTOCOL_CONFIG, "SSL");
        configProps.put(SslConfigs.SSL_KEYSTORE_LOCATION_CONFIG, keystoreLocation);
        configProps.put(SslConfigs.SSL_TRUSTSTORE_LOCATION_CONFIG, truststoreLocation);
        configProps.put(SslConfigs.SSL_KEYSTORE_PASSWORD_CONFIG, kafkaKeystoreCreds);
        configProps.put(SslConfigs.SSL_TRUSTSTORE_PASSWORD_CONFIG, kafkaTruststoreCreds);
        configProps.put(SslConfigs.SSL_KEY_PASSWORD_CONFIG, kafkaKeyCreds);
        return configProps;
    }

    @Override
    public void watchKafkaKeystore(String keystoreLocation, String trueStoreLocation) {
        ClientSecretName secretName = secretNameFromClientName(ClientName.CLIENT_KAFKA);
        this.kclient.watchSecretData(secretName.toString(), NAMESPACE, secret -> {
            if (secret == null) {
                System.out.println("Exiting due to Kafka tls certificate change.");
                System.exit(0);
            }

            byte[] keystoreContents;
            File keyStoreFile = new File(keystoreLocation);
            try {
                keystoreContents  = Files.readAllBytes(keyStoreFile.toPath());
            } catch(Exception e) {
                throw new RuntimeException("cannot read keystore file in the kubernetes secret change handler");
            }
            byte[] keystoreData = secret.getData().get(SECRET_KAFKA_KEYSTORE_NAME);

            if (Arrays.equals(keystoreData, keystoreContents)) {
                System.out.println("Exiting due to Kafka tls certificate change.");
                System.exit(0);
            }
        });
    }

    private SSLHostConfig getTLSConfFromSecrets(ClientName serverName) throws Exception {
        ClientSecretName secretName = secretNameFromClientName(serverName);
        V1Secret secret = this.kclient.fetchSecretData(NAMESPACE, secretName.toString());
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
        KeyPairGenerator keyPairGenerator = KeyPairGenerator.getInstance("RSA", "BC");
        keyPairGenerator.initialize(1024, new SecureRandom());
        return keyPairGenerator.generateKeyPair();
    }

    private X509Certificate createCertificate(KeyPair keyPair) throws Exception {

//        // Prepare the information required for generating an X.509 certificate.
//        X500Name owner = new X500Name(CERTIFICATE_DN);
//        X509v3CertificateBuilder builder = new JcaX509v3CertificateBuilder(owner, new BigInteger(64, new SecureRandom()), NOT_BEFORE, NOT_AFTER, owner, keyPair.getPublic());
//
//        // Subject alternative name (part of SNI extension, used for hostname verification)
//        GeneralNames subjectAlternativeName = new GeneralNames(new GeneralName(GeneralName.dNSName, "localhost"));
//        builder.addExtension(Extension.subjectAlternativeName, false, subjectAlternativeName);
//
//        PrivateKey privateKey = keyPair.getPrivate();
//        ContentSigner signer = new JcaContentSignerBuilder("SHA512WithRSAEncryption").build(privateKey);
//        X509CertificateHolder certHolder = builder.build(signer);
//        X509Certificate cert = new JcaX509CertificateConverter().setProvider(new BouncyCastleProvider()).getCertificate(certHolder);
//
//        String pem = convertToPem(cert);
//        System.out.println("PEM CERT: ");
//        System.out.println(pem);
//
//        //check so that cert is valid
//        cert.verify(keyPair.getPublic());
//        return cert;
        return null;
    }

    private void saveCert(X509Certificate cert, PrivateKey key, KeyStore keyStore) throws KeyStoreException, NoSuchAlgorithmException, IOException, CertificateException {
        char[] keystorePassword = KEYSTORE_PASSWORD.toCharArray();
        keyStore.setKeyEntry(serverCertificateAlias, key, keystorePassword, new java.security.cert.Certificate[]{cert});
        File file = new File(".", serverCertificateName);
        keyStore.store(new FileOutputStream(file), keystorePassword);
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
            case CLIENT_KAFKA:
                return ClientSecretName.KAFKA_SECRET_NAME;
            default:
                return ClientSecretName.NOT_VALID_SECRET_NAME;
        }
    }
}
