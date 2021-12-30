package tlsmanager

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"log"
	"math/big"

	kclient "greenops.io/workflowtrigger/kubernetesclient"
)

/* TODO

1. create a function to generate TLS certs for servers
2. create a function to get certificate from secrets
3. use secrets crt or self signed
4. create a function to update servers on certificate change
	use kubernetesclient/WatchSecretData function
*/

type Manager interface {
	GetTLSConf() (*tls.Config, error)
	WatchTLSConf(handler func(conf *tls.Config, err error))
}

type tlsManager struct {
	k                 kclient.KubernetesClient
	selfSignedTLSConf *tls.Config
}

// TODO: currently those values are hardcoded, fetch them from config or somewhere else
const (
	atlasTLSSecretName = "atlas-server-tls"
	atlasNamespace     = "default"
)

func New(k kclient.KubernetesClient) Manager {
	return &tlsManager{k: k}
}

func (m *tlsManager) GetTLSConf() (*tls.Config, error) {
	log.Println("in GetTLSConf")
	cert, err := m.getTLSConfFromSecrets()
	log.Printf("in GetTLSConf, cert = %v\n", cert)
	if err != nil {
		return nil, err
	}
	if cert != nil {
		return cert, nil
	}

	log.Println("in GetTLSConf, before getSelfSignedTLSConf")
	return m.getSelfSignedTLSConf()
}

func (m *tlsManager) WatchTLSConf(handler func(conf *tls.Config, err error)) {
	log.Println("in WatchTLSConf")
	m.k.WatchSecretData(atlasTLSSecretName, atlasNamespace, func(t kclient.SecretChangeType, obj interface{}) {
		log.Printf("in WatchTLSConf, event %v. data = %s\n", t, obj)
		switch t {
		case kclient.SecretChangeTypeAdd:
		case kclient.SecretChangeTypeUpdate:
			secret := obj.(map[string][]byte)
			log.Printf("in WatchTLSConf, secret = %v\n", secret)
			conf, err := m.generateTLSConfFromKeyPair(secret["cert"], secret["key"])
			log.Printf("in WatchTLSConf, conf = %v\n", conf)
			if err != nil {
				handler(nil, err)
			}
			handler(conf, nil)
		case kclient.SecretChangeTypeDelete:
			handler(m.selfSignedTLSConf, nil)
		}
	})
}

func (m *tlsManager) getSelfSignedTLSConf() (*tls.Config, error) {
	log.Println("in getSelfSignedTLSConf")
	if m.selfSignedTLSConf != nil {
		return m.selfSignedTLSConf, nil
	}

	conf, err := m.generateSelfSignedTLSConf()
	if err != nil {
		return nil, err
	}
	log.Printf("in getSelfSignedTLSConf, conf = %v\n", conf)

	m.selfSignedTLSConf = conf
	return conf, nil
}

func (m *tlsManager) getTLSConfFromSecrets() (*tls.Config, error) {
	log.Println("in getTLSConfFromSecrets")
	secret := m.k.FetchSecretData(atlasTLSSecretName, atlasNamespace)
	log.Println("in getTLSConfFromSecrets, secret: ", secret)
	if secret == nil {
		return nil, nil
	}

	return m.generateTLSConfFromKeyPair(secret["tls.crt"], secret["tls.key"])
}

func (m *tlsManager) generateTLSConfFromKeyPair(cert []byte, key []byte) (*tls.Config, error) {
	log.Printf("in generateTLSConfFromKeyPair cert = %s, key = %s\n", string(cert), string(key))
	c, err := tls.X509KeyPair(cert, key)
	log.Printf("in generateTLSConfFromKeyPair c = %v\n", c)
	if err != nil {
		return nil, err
	}

	rootCAs := bestEffortSystemCertPool()
	return &tls.Config{
		Certificates:             []tls.Certificate{c},
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		RootCAs:                  rootCAs,
	}, nil
}

func bestEffortSystemCertPool() *x509.CertPool {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		log.Println("root ca not found, returning new...")
		return x509.NewCertPool()
	}
	log.Println("root ca found")
	return rootCAs
}

// TODO: add certificate to the global registry or create a new registry
// TODO: try to use most secure configuration for ca and certificate conf
func (m *tlsManager) generateSelfSignedTLSConf() (*tls.Config, error) {
	rootCAs := bestEffortSystemCertPool()

	// TODO: try to delete ca creation if it works without it
	//log.Printf("in generateSelfSignedTLSConf")
	//
	//caSerialNumber, err := generateCertificateSerialNumber()
	//if err != nil {
	//	return nil, err
	//}
	//
	//// create a CA certificate
	//ca := &x509.Certificate{
	//	SerialNumber: caSerialNumber,
	//	Subject: pkix.Name{
	//		Organization: []string{"Atlas"},
	//		Country:      []string{"US"},
	//	},
	//	NotBefore:             time.Now(),
	//	NotAfter:              time.Now().AddDate(10, 0, 0),
	//	IsCA:                  true,
	//	DNSNames:              []string{"localhost"},
	//	ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	//	KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment,
	//	BasicConstraintsValid: true,
	//}
	//rootCAs.AddCert(ca)
	//log.Printf("in generateSelfSignedTLSConf, ca = %v\n", ca)
	//
	//caPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	//if err != nil {
	//	return nil, err
	//}
	//log.Printf("in generateSelfSignedTLSConf, caPrivateKey = %v\n", caPrivateKey)
	//
	//caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivateKey.PublicKey, caPrivateKey)
	//if err != nil {
	//	return nil, err
	//}
	//log.Printf("in generateSelfSignedTLSConf, caBytes = %v\n", caBytes)
	//
	//caPEM := new(bytes.Buffer)
	//err = pem.Encode(caPEM, &pem.Block{
	//	Type:  "CERTIFICATE",
	//	Bytes: caBytes,
	//})
	//if err != nil {
	//	return nil, err
	//}
	//log.Printf("in generateSelfSignedTLSConf, caPEM = %v\n", caPEM)
	//
	//caPrivateKeyPEM := new(bytes.Buffer)
	//err = pem.Encode(caPrivateKeyPEM, &pem.Block{
	//	Type:  "RSA PRIVATE KEY",
	//	Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey),
	//})
	//if err != nil {
	//	return nil, err
	//}
	//log.Printf("in generateSelfSignedTLSConf, caPrivateKeyPEM = %v\n", caPrivateKeyPEM)
	//

	// TODO: uncomment
	//certSerialNumber, err := generateCertificateSerialNumber()
	//if err != nil {
	//	return nil, err
	//}

	//// create a server certificate
	//cert := &x509.Certificate{
	//	SerialNumber: certSerialNumber,
	//	Subject: pkix.Name{
	//		Organization: []string{"GreenOps, INC."},
	//		Country:      []string{"US"},
	//	},
	//	DNSNames:     []string{"localhost"},
	//	IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	//	NotBefore:    time.Now(),
	//	NotAfter:     time.Now().AddDate(10, 0, 0),
	//	SubjectKeyId: []byte{1, 2, 3, 4, 6},
	//	ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	//	KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	//}
	//log.Printf("in generateSelfSignedTLSConf, cert = %v\n", cert)
	//
	//certPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	//if err != nil {
	//	return nil, err
	//}
	//log.Printf("in generateSelfSignedTLSConf, certPrivateKey = %v\n", certPrivateKey)
	//
	//certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &certPrivateKey.PublicKey, certPrivateKey)
	//if err != nil {
	//	return nil, err
	//}
	//log.Printf("in generateSelfSignedTLSConf, certBytes = %v\n", certBytes)
	//
	//certPEM := new(bytes.Buffer)
	//err = pem.Encode(certPEM, &pem.Block{
	//	Type:  "CERTIFICATE",
	//	Bytes: certBytes,
	//})
	//if err != nil {
	//	return nil, err
	//}
	//log.Printf("in generateSelfSignedTLSConf, certPEM = %v\n", certPEM)
	//
	//certPrivateKeyPEM := new(bytes.Buffer)
	//err = pem.Encode(certPrivateKeyPEM, &pem.Block{
	//	Type:  "RSA PRIVATE KEY",
	//	Bytes: x509.MarshalPKCS1PrivateKey(certPrivateKey),
	//})
	//if err != nil {
	//	return nil, err
	//}
	//log.Printf("in generateSelfSignedTLSConf, certPrivateKeyPEM = %v\n", certPrivateKeyPEM)
	//
	//log.Printf("cert PEM = %s\n", certPEM.String())
	//log.Printf("key PEM = %s\n", certPrivateKeyPEM.String())

	certPEM := `-----BEGIN CERTIFICATE-----
MIIFOzCCAyOgAwIBAgIQRxTE4jdMRNoFXqG1Z6u0uzANBgkqhkiG9w0BAQsFADAm
MQswCQYDVQQGEwJVUzEXMBUGA1UEChMOR3JlZW5PcHMsIElOQy4wHhcNMjExMjMw
MjEwNTU5WhcNMzExMjMwMjEwNTU5WjAmMQswCQYDVQQGEwJVUzEXMBUGA1UEChMO
R3JlZW5PcHMsIElOQy4wggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQDJ
7cg1TSP/czsAq2Ca5ItndyaB6azQn8l0oVkXbT858E5XnFxtedzsmgu6+lNcOEon
1kKf3Z8HvMkw0rGlNI0MHIIB/FunB79LIonUmgBCNb+PXJcaCmsFVoRFXNKMjh89
4QpjGhy1XFw1yLfTUQpYsScx9OdhgZFihkga+/RD6Nr3oadr0n6omZmn+eAfvaGm
99xkY+sy47vtCRDFGJyCb37fDZZC1R914TBhcIAMeVMCUnIQe49xTLwDF4CSGVB7
RE8rwZ3KtPM0tso+ZpyJXjVyvPboDIRYn7DJld7sVV1Atf039DLfhmviANw3BO6m
r3iVcGOyhF+D9nfkqW3jHSHdgpBEMjbKs+kqzlnVERb/8bFH0GGA5oK5f5nb8urP
aBr11T6OEZRrfMAD4BzAP6XR+GGWISQsOs9KvZDSRKmYHvPlhvJKeoREqVHjHisE
lHctxumLJf1FS5oPFfDVd0Gg447Okc0Ne9eE0ajBM9o/Yr2xaAOvyfyQa/nyj0GV
cgoTFLCDUJmV9AADEH48dA3w8N0j2MuvCtkOALdL6v4aaHE/wL3cd77p1hDgrgkh
ZfHfKHtsUCIUivc+pWJknPiHczGRQzgdkbQj+mKJygJ4AdY/lWqR2JcneFETbbmr
ZspYR8mm141plfI37Z3eMUnWYAPHw4W+GbKTCUEgTQIDAQABo2UwYzAOBgNVHQ8B
Af8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDgYDVR0OBAcEBQECAwQGMCwG
A1UdEQQlMCOCCWxvY2FsaG9zdIcEfwAAAYcQAAAAAAAAAAAAAAAAAAAAATANBgkq
hkiG9w0BAQsFAAOCAgEANQ1i34RTA9N/ZoKvpGqcTZktOgehBu28ppFhFDkpsHv4
G3Sx/ls9Eg11m+TnyAjDA1G8lEJeO6P+br+BmrAj4cVYcRRfTJrRA8t1zfpKVaks
wnqEuxc1iOpecOAXD5rG/yuibCgsT+ZRE6d+GZVT2Iv4tjoq4KW3UdWRvDOuk9kr
k7q6v84jx20V0LRtKH5pnsdkpNqDkm+3Mgu9MeASGpINuX7OL3sbaHlZxs44wSFW
0mRW8ehroKGms9r8Z9KD//mun9HTgyhEE4pVIG9eghBKimfvibIdZTa39m6EL5E2
xhEejiNJun3HMe1mVW/FEqtR8zCb7vSoyi1uTCaJYmQR4rZ1oj9Ku4fGsgEOx6KS
FwlcCx8/b5BsHeDtMs8ef/mXIvR0GwxxO2U6Dj6i0B0t+Binzp2skw48uqh9Fj6Z
eXjkzKhtS1/Pp6AzADZROgM71BzhOekswO3alNmz9DO3jAGCALu8PB5Y09WHgbtG
XU9MfdIvFj6wB6mUcOCFjTbC7WtfFwE5d/T4wXL5MhF25Dbluq8FcIAVnclack5t
J2q79aptXZekQaWEiGtBjBglf7B4PDwP4kWRrVnRT4A5eJacg26g5AoHLt4Jn8RD
lWiSBervciIiV63WDR5PmLabXPZ5EWK6qu1BD0erG9Z534wYF4XmWb3rnZDKMoY=
-----END CERTIFICATE-----`
	certPrivateKeyPEM := `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`

	serverCert, err := tls.X509KeyPair([]byte(certPEM), []byte(certPrivateKeyPEM))
	if err != nil {
		return nil, err
	}
	rootCAs.AppendCertsFromPEM([]byte(certPEM))

	log.Printf("in generateSelfSignedTLSConf, serverCert = %v\n", serverCert)
	return &tls.Config{
		Certificates:             []tls.Certificate{serverCert},
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		RootCAs:                  rootCAs,
	}, nil
}

func generateCertificateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, serialNumberLimit)
}
