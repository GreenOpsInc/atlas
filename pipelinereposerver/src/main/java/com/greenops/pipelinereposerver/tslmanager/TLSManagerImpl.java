package com.greenops.pipelinereposerver.tslmanager;

import com.greenops.pipelinereposerver.kubernetesclient.KubernetesClient;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

@Slf4j
@Component
public class TLSManagerImpl implements TLSManager {

    private final KubernetesClient kclient;
    private final String REPO_SERVER_SECRET_NAME = "pipelinereposerver-tls";
    private final String NAMESPACE = "default";

    private String /*tls.Config*/ tlsConf;
    private String /*tls.Config*/ selfSignedConf;
    private String /*map[ClientName]*tls.Config*/ tlsClientConfigs;
    private String /*map[ClientName][]byte*/ tlsClientCertPEM;

    @Autowired
    public TLSManagerImpl(KubernetesClient kclient) {
        this.kclient = kclient;

        this.getServerTLSConf();
    }

    @Override
    public void bestEffortSystemCertPool() {
        rootCAs, _ := x509.SystemCertPool()
        if rootCAs == nil {
            log.Println("root ca not found, returning new...")
            return x509.NewCertPool()
        }
        log.Println("root ca found")
        return rootCAs
    }

    @Override
    public void getServerTLSConf() {
        conf, err := m.getTLSConf()
        if err != nil {
            return nil, err
        }
        m.tlsConf = conf
        return conf, nil
    }

    @Override
    public void watchServerTLSConf(String handler) {
        log.Println("in WatchServerTLSConf")
        m.k.WatchSecretData(string(WorkflowTriggerTLSSecretName), Namespace, func(t kclient.SecretChangeType, secret *corev1.Secret) {
            log.Printf("in WatchServerTLSConf, event %v. data = %s\n", t, secret)
            var (
                    config   *tls.Config
                    err      error
                    insecure bool
            )

            switch t {
                case kclient.SecretChangeTypeAdd:
                    fallthrough
                case kclient.SecretChangeTypeUpdate:
                    log.Printf("in WatchServerTLSConf, secret data = %v\n", secret.Data)
                    config, err = m.generateTLSConfFromKeyPair(secret.Data[TLSSecretCrtName], secret.Data[TLSSecretCrtName])
                    log.Printf("in WatchServerTLSConf, tlsConf = %v\n", config)
                    insecure = false
                case kclient.SecretChangeTypeDelete:
                    config, err = m.getSelfSignedTLSConf()
                    insecure = true
            }

            if err != nil {
                handler(nil, err)
                return
            }

            config.InsecureSkipVerify = insecure
            m.tlsConf = config
            handler(config, nil)
        })
    }

    // TODO: need to create a general method for get to get delegator, repo server and wrapper secrets
    private String /* *tls.Config*/ getTLSConf() {
        log.Println("in GetServerTLSConf")
        if m.tlsConf != nil {
            return m.tlsConf, nil
        }

        conf, err := m.getTLSConfFromSecrets()
        log.Printf("in GetServerTLSConf, tlsConf = %v\n", conf)
        if err != nil {
            return nil, err
        }
        if conf != nil {
            log.Println("CERT FOUND IN SECRETS")
            conf.InsecureSkipVerify = false
            return conf, nil
        }

        log.Println("in GetServerTLSConf, before getSelfSignedTLSConf")
        conf, err = m.getSelfSignedTLSConf()
        if err != nil {
            return nil, err
        }

        conf.InsecureSkipVerify = true
        return conf, nil
    }

    private /* *tls.Config*/ String getSelfSignedTLSConf() {
        log.Println("in getSelfSignedTLSConf")
        if m.selfSignedConf != nil {
            return m.selfSignedConf, nil
        }

        conf, err := m.generateSelfSignedTLSConf()
        if err != nil {
            return nil, err
        }
        log.Printf("in getSelfSignedTLSConf, tlsConf = %v\n", conf)

        m.selfSignedConf = conf
        return conf, nil
    }

    private  /* *tls.Config*/ String getTLSConfFromSecrets() {
        log.Println("in getTLSConfFromSecrets")
        secret := m.k.FetchSecretData(string(WorkflowTriggerTLSSecretName), Namespace)
        log.Println("in getTLSConfFromSecrets, secret: ", secret)
        if secret == nil {
            return nil, nil
        }

        conf, err := m.generateTLSConfFromKeyPair(secret[TLSSecretCrtName], secret[TLSSecretKeyName])
        if err != nil {
            return nil, err
        }
        return conf, nil
    }

    private  /* *tls.Config*/ String generateTLSConfFromKeyPair(byte[] cert , byte[] key) {
        log.Printf("in generateTLSConfFromKeyPair cert = %s, key = %s\n", string(cert), string(key))
        c, err := tls.X509KeyPair(cert, key)
        log.Printf("in generateTLSConfFromKeyPair c = %v\n", c)
        if err != nil {
            return nil, err
        }

        rootCAs := m.BestEffortSystemCertPool()
        return &tls.Config{
            Certificates:             []tls.Certificate{c},
            MinVersion:               tls.VersionTLS13,
                    PreferServerCipherSuites: true,
                    RootCAs:                  rootCAs,
        }, nil
    }

    private  /* *tls.Config*/ String generateSelfSignedTLSConf() {
        // JAVA
        X509Certificate cert = null;
        KeyPairGenerator keyPairGenerator = KeyPairGenerator.getInstance(CERTIFICATE_ALGORITHM);
        keyPairGenerator.initialize(CERTIFICATE_BITS, new SecureRandom());
        KeyPair keyPair = keyPairGenerator.generateKeyPair();

        // GENERATE THE X509 CERTIFICATE
        X509V3CertificateGenerator v3CertGen =  new X509V3CertificateGenerator();
        v3CertGen.setSerialNumber(BigInteger.valueOf(System.currentTimeMillis()));
        v3CertGen.setIssuerDN(new X509Principal(CERTIFICATE_DN));
        v3CertGen.setNotBefore(new Date(System.currentTimeMillis() - 1000L * 60 * 60 * 24));
        v3CertGen.setNotAfter(new Date(System.currentTimeMillis() + (1000L * 60 * 60 * 24 * 365*10)));
        v3CertGen.setSubjectDN(new X509Principal(CERTIFICATE_DN));
        v3CertGen.setPublicKey(keyPair.getPublic());
        v3CertGen.setSignatureAlgorithm("SHA256WithRSAEncryption");
        cert = v3CertGen.generateX509Certificate(keyPair.getPrivate());
        saveCert(cert,keyPair.getPrivate());
        return cert;

        // GO
        certSerialNumber, err := generateCertificateSerialNumber()
        if err != nil {
            return nil, err
        }

        cert := &x509.Certificate{
            SerialNumber: certSerialNumber,
                    Subject: pkix.Name{
                Organization: []string{"GreenOps, INC."},
                Country:      []string{"US"},
            },
            DNSNames:     []string{"localhost"},
            IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
            NotBefore:    time.Now(),
                    NotAfter:     time.Now().AddDate(10, 0, 0),
                    SubjectKeyId: []byte{1, 2, 3, 4, 6},
            ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
            KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
        }
        log.Printf("in generateSelfSignedTLSConf, cert = %v\n", cert)

        certPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
        if err != nil {
            return nil, err
        }
        log.Printf("in generateSelfSignedTLSConf, certPrivateKey = %v\n", certPrivateKey)

        certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &certPrivateKey.PublicKey, certPrivateKey)
        if err != nil {
            return nil, err
        }
        log.Printf("in generateSelfSignedTLSConf, certBytes = %v\n", certBytes)

        certPEM := new(bytes.Buffer)
                err = pem.Encode(certPEM, &pem.Block{
            Type:  "CERTIFICATE",
                    Bytes: certBytes,
        })
        if err != nil {
            return nil, err
        }
        log.Printf("in generateSelfSignedTLSConf, certPEM = %v\n", certPEM)

        certPrivateKeyPEM := new(bytes.Buffer)
                err = pem.Encode(certPrivateKeyPEM, &pem.Block{
            Type:  "RSA PRIVATE KEY",
                    Bytes: x509.MarshalPKCS1PrivateKey(certPrivateKey),
        })
        if err != nil {
            return nil, err
        }
        log.Printf("in generateSelfSignedTLSConf, certPrivateKeyPEM = %v\n", certPrivateKeyPEM)

        log.Printf("cert PEM = %s\n", certPEM.String())
        log.Printf("key PEM = %s\n", certPrivateKeyPEM.String())

        serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivateKeyPEM.Bytes())
        if err != nil {
            return nil, err
        }

        rootCAs := m.BestEffortSystemCertPool()
        rootCAs.AppendCertsFromPEM(certPEM.Bytes())

        log.Printf("in generateSelfSignedTLSConf, serverCert = %v\n", serverCert)
        return &tls.Config{
            Certificates:             []tls.Certificate{serverCert},
            MinVersion:               tls.VersionTLS13,
                    PreferServerCipherSuites: true,
                    RootCAs:                  rootCAs,
        }, nil
    }

    private int generateCertificateSerialNumber(){
        serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
        return rand.Int(rand.Reader, serialNumberLimit)
    }
}
