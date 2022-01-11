package com.greenops.util.tslmanager;

import com.greenops.util.kubernetesclient.KubernetesClient;
import io.kubernetes.client.models.V1Secret;
import lombok.extern.slf4j.Slf4j;
import org.apache.commons.codec.binary.Base64;
import org.apache.tomcat.util.net.SSLHostConfig;
import org.apache.tomcat.util.net.SSLHostConfigCertificate;
import org.bouncycastle.asn1.x500.X500Name;
import org.bouncycastle.asn1.x509.Extension;
import org.bouncycastle.asn1.x509.GeneralName;
import org.bouncycastle.asn1.x509.GeneralNames;
import org.bouncycastle.cert.X509CertificateHolder;
import org.bouncycastle.cert.X509v3CertificateBuilder;
import org.bouncycastle.cert.jcajce.JcaX509CertificateConverter;
import org.bouncycastle.cert.jcajce.JcaX509v3CertificateBuilder;
import org.bouncycastle.jce.provider.BouncyCastleProvider;
import org.bouncycastle.operator.ContentSigner;
import org.bouncycastle.operator.jcajce.JcaContentSignerBuilder;
import org.bouncycastle.pqc.jcajce.provider.BouncyCastlePQCProvider;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.math.BigInteger;
import java.security.*;
import java.security.cert.CertificateEncodingException;
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
    private static final String KEYSTORE_PASSWORD = "SS28qmtOJH4OFLUP";
    private static final Date NOT_BEFORE = new Date(System.currentTimeMillis() - 86400000L * 365);
    private static final Date NOT_AFTER = new Date(253402300799000L);

    static {
        if (Security.getProvider(BouncyCastleProvider.PROVIDER_NAME) == null) {
            Security.insertProviderAt(new BouncyCastleProvider(), 1);
        }
        if (Security.getProvider(BouncyCastlePQCProvider.PROVIDER_NAME) == null) {
            Security.insertProviderAt(new BouncyCastlePQCProvider(), 2);
        }
    }

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
    public boolean updateKafkaKeystore(String keystoreLocation, String trueStoreLocation) throws Exception {
        File keyStoreFile = new File(keystoreLocation);
        if (keyStoreFile.exists()) {
            keyStoreFile.delete();
        }
        File trueStoreFile = new File(trueStoreLocation);
        if (trueStoreFile.exists()) {
            trueStoreFile.delete();
        }

        ClientSecretName secretName = secretNameFromClientName(ClientName.CLIENT_KAFKA);
        System.out.println("in updateKafkaKeystore keystoreLocation = " + keystoreLocation + " trueStoreLocation = " + trueStoreLocation + " secretName = " + secretName.toString() + " namespace = " + NAMESPACE);
        V1Secret secret = this.kclient.fetchSecretData(NAMESPACE, secretName.toString());
        if (secret == null) {
            return false;
        }

        KeyPair kp = createKeyPair(secret.getData().get(SECRET_CERT_NAME), secret.getData().get(SECRET_KEY_NAME));
        X509Certificate cert = createCertificate(kp);

        KeyStore keystore = KeyStore.getInstance(KeyStore.getDefaultType());
        keystore.setCertificateEntry("kafka.client.keystore", cert);
        keyStoreFile = new File(keystoreLocation);
        keyStoreFile.createNewFile();
        keystore.store(new FileOutputStream(keyStoreFile), null);

        KeyStore truestore = KeyStore.getInstance(KeyStore.getDefaultType());
        truestore.setCertificateEntry("kafka.client.truestore", cert);
        trueStoreFile = new File(keystoreLocation);
        trueStoreFile.createNewFile();
        keystore.store(new FileOutputStream(trueStoreFile), null);

        return true;
    }

    @Override
    public void watchKafkaKeystore(String keystoreLocation, String trueStoreLocation) {
        ClientSecretName secretName = secretNameFromClientName(ClientName.CLIENT_KAFKA);
        this.kclient.watchSecretData(secretName.toString(), NAMESPACE, data -> {
            System.out.println("Exiting due to Kafka certificate change.");
            System.exit(0);
        });
    }

    private SSLHostConfig getTLSConfFromSecrets(ClientName serverName) throws Exception {
        ClientSecretName secretName = secretNameFromClientName(serverName);
        System.out.println("in getTLSConfFromSecrets server name = " + serverName + " secret name = " + secretName.toString() + " namespace = " + NAMESPACE);
        V1Secret secret = this.kclient.fetchSecretData(NAMESPACE, secretName.toString());
        if (secret == null) {
            return null;
        }
        return generateSSLHostConfFromKeyPair(secret.getData().get(SECRET_CERT_NAME), secret.getData().get(SECRET_KEY_NAME));
    }

    private SSLHostConfig generateSSLHostConfFromKeyPair(byte[] pub, byte[] key) throws Exception {
        System.out.println("in generateSSLHostConfFromKeyPair");
        KeyStore keyStore = getKeyStore();
        System.out.println("in generateSSLHostConfFromKeyPair keyStore = " + keyStore);

        KeyPair keyPair = createKeyPair(pub, key);
        System.out.println("in generateSSLHostConfFromKeyPair keyPair = " + keyPair);
        X509Certificate cert = createCertificate(keyPair);
        System.out.println("in generateSSLHostConfFromKeyPair cert = " + cert);
        saveCert(cert, keyPair.getPrivate(), keyStore);
        System.out.println("in generateSSLHostConfFromKeyPair after saveCert");

        SSLHostConfig sslHostConfig = new SSLHostConfig();
        System.out.println("in generateSSLHostConfFromKeyPair sslHostConfig = " + sslHostConfig);
        SSLHostConfigCertificate certificate = new SSLHostConfigCertificate(
                sslHostConfig, SSLHostConfigCertificate.Type.RSA);
        System.out.println("in generateSSLHostConfFromKeyPair certificate = " + certificate);
        certificate.setCertificateKeystore(keyStore);
        System.out.println("in generateSSLHostConfFromKeyPair after certificate.setCertificateKeystore");
        sslHostConfig.addCertificate(certificate);
        System.out.println("in generateSSLHostConfFromKeyPair after sslHostConfig.addCertificate");
        return sslHostConfig;
    }

    private SSLHostConfig getSelfSignedSSLHostConf() throws Exception {
        System.out.println("in getSelfSignedSSLHostConf");
        KeyStore keyStore = getKeyStore();
        System.out.println("in getSelfSignedSSLHostConf keyStore = " + keyStore);

        KeyPair keyPair = generateKeyPair();
        System.out.println("in getSelfSignedSSLHostConf keyPair = " + keyPair);
        X509Certificate cert = createCertificate(keyPair);
        System.out.println("in getSelfSignedSSLHostConf cert = " + cert);
        saveCert(cert, keyPair.getPrivate(), keyStore);
        System.out.println("in getSelfSignedSSLHostConf after saveCert");

        SSLHostConfig sslHostConfig = new SSLHostConfig();
        System.out.println("in getSelfSignedSSLHostConf sslHostConfig = " + sslHostConfig);
        SSLHostConfigCertificate certificate = new SSLHostConfigCertificate(
                sslHostConfig, SSLHostConfigCertificate.Type.RSA);
        System.out.println("in getSelfSignedSSLHostConf certificate = " + certificate);
        certificate.setCertificateKeystore(keyStore);
        System.out.println("in getSelfSignedSSLHostConf after certificate.setCertificateKeystore");
        sslHostConfig.addCertificate(certificate);
        System.out.println("in getSelfSignedSSLHostConf after sslHostConfig.addCertificate");
        sslHostConfig.setInsecureRenegotiation(true);
        System.out.println("in getSelfSignedSSLHostConf after sslHostConfig.setInsecureRenegotiation");
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
//        KeyPairGenerator keyPairGenerator = KeyPairGenerator.getInstance(CERTIFICATE_ALGORITHM);
//        keyPairGenerator.initialize(CERTIFICATE_BITS, new SecureRandom());
//        return keyPairGenerator.generateKeyPair();
        // GENERATE THE PUBLIC/PRIVATE RSA KEY PAIR
        KeyPairGenerator keyPairGenerator = KeyPairGenerator.getInstance("RSA", "BC");
        keyPairGenerator.initialize(1024, new SecureRandom());
        return keyPairGenerator.generateKeyPair();
    }

    // TODO: certificate generation method N1
    @SuppressWarnings("deprecation")
//    private X509Certificate createCertificate(KeyPair keyPair) throws InvalidKeyException, SignatureException {
//        X509V3CertificateGenerator v3CertGen = new X509V3CertificateGenerator();
//        v3CertGen.setSerialNumber(BigInteger.valueOf(System.currentTimeMillis()));
//        v3CertGen.setIssuerDN(new X509Principal(CERTIFICATE_DN));
//        v3CertGen.setNotBefore(new Date(System.currentTimeMillis() - 1000L * 60 * 60 * 24));
//        v3CertGen.setNotAfter(new Date(System.currentTimeMillis() + (1000L * 60 * 60 * 24 * 365 * 10)));
//        v3CertGen.setSubjectDN(new X509Principal(CERTIFICATE_DN));
//        v3CertGen.setPublicKey(keyPair.getPublic());
//        v3CertGen.setSignatureAlgorithm("SHA256WithRSAEncryption");
//        return v3CertGen.generateX509Certificate(keyPair.getPrivate());
//    }

    // TODO: certificate generation method N2
//    private X509Certificate createCertificate(KeyPair keyPair) throws Exception {
//        // yesterday
//        Date validityBeginDate = new Date(System.currentTimeMillis() - 24 * 60 * 60 * 1000);
//        // in 2 years
//        Date validityEndDate = new Date(System.currentTimeMillis() + 2L * 365 * 24 * 60 * 60 * 1000);
//
//        // GENERATE THE X509 CERTIFICATE
//        X509V1CertificateGenerator certGen = new X509V1CertificateGenerator();
//        X500Principal dnName = new X500Principal(CERTIFICATE_DN);
//
//        certGen.setSerialNumber(BigInteger.valueOf(System.currentTimeMillis()));
//        certGen.setSubjectDN(dnName);
//        certGen.setIssuerDN(dnName); // use the same
//        certGen.setNotBefore(validityBeginDate);
//        certGen.setNotAfter(validityEndDate);
//        certGen.setPublicKey(keyPair.getPublic());
//        certGen.setSignatureAlgorithm("SHA256WithRSAEncryption");
//
//        X509Certificate cert = certGen.generate(keyPair.getPrivate(), "BC");
//
//        PEMWriter pemWriter = new PEMWriter(new PrintWriter(System.out));
//        pemWriter.writeObject(cert);
//        pemWriter.flush();
//        pemWriter.writeObject(keyPair.getPrivate());
//        pemWriter.flush();
//        return cert;
//    }

    // TODO: fix error: java.lang.NullPointerException: Cannot invoke "java.security.SecureRandom.nextBytes(byte[])" because "this.random" is null
    // TODO: certificate generation method N3
    private X509Certificate createCertificate(KeyPair keyPair) throws Exception {

        // Prepare the information required for generating an X.509 certificate.
        X500Name owner = new X500Name(CERTIFICATE_DN);
        X509v3CertificateBuilder builder = new JcaX509v3CertificateBuilder(owner, new BigInteger(64, new SecureRandom()), NOT_BEFORE, NOT_AFTER, owner, keyPair.getPublic());

        // Subject alternative name (part of SNI extension, used for hostname verification)
        GeneralNames subjectAlternativeName = new GeneralNames(new GeneralName(GeneralName.dNSName, "localhost"));
        builder.addExtension(Extension.subjectAlternativeName, false, subjectAlternativeName);

        PrivateKey privateKey = keyPair.getPrivate();
        ContentSigner signer = new JcaContentSignerBuilder("SHA512WithRSAEncryption").build(privateKey);
        X509CertificateHolder certHolder = builder.build(signer);
        X509Certificate cert = new JcaX509CertificateConverter().setProvider(new BouncyCastleProvider()).getCertificate(certHolder);

        String pem = convertToPem(cert);
        System.out.println("PEM CERT: ");
        System.out.println(pem);

        //check so that cert is valid
        cert.verify(keyPair.getPublic());
        return cert;
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
