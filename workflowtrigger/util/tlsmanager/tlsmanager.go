package tlsmanager

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"time"

	kclient "greenops.io/workflowtrigger/kubernetesclient"
	corev1 "k8s.io/api/core/v1"
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

const (
	secretCrtName = "tls.crt"
	secretKeyName = "tls.key"
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
		log.Println("CERT FOUND IN SECRETS")
		return cert, nil
	}

	log.Println("in GetTLSConf, before getSelfSignedTLSConf")
	return m.getSelfSignedTLSConf()
}

// TODO: to add a tls cert\key to secrets run:
//		kubectl create secret tls atlas-server-tls --cert ./cert.pem --key ./key.pem
// TODO: check that this wather is not trigger server reloading if cert is not changed
//		it could possibly happend on server start when cert is available and we also receiving secret change event
func (m *tlsManager) WatchTLSConf(handler func(conf *tls.Config, err error)) {
	log.Println("in WatchTLSConf")
	m.k.WatchSecretData(atlasTLSSecretName, atlasNamespace, func(t kclient.SecretChangeType, secret *corev1.Secret) {
		log.Printf("in WatchTLSConf, event %v. data = %s\n", t, secret)
		switch t {
		case kclient.SecretChangeTypeAdd:
			fallthrough
		case kclient.SecretChangeTypeUpdate:
			log.Printf("in WatchTLSConf, secret data = %v\n", secret.Data)
			conf, err := m.generateTLSConfFromKeyPair(secret.Data[secretCrtName], secret.Data[secretKeyName])
			log.Printf("in WatchTLSConf, conf = %v\n", conf)
			if err != nil {
				handler(nil, err)
			}
			handler(conf, nil)
		case kclient.SecretChangeTypeDelete:
			conf, err := m.getSelfSignedTLSConf()
			if err != nil {
				handler(nil, err)
			}
			handler(conf, nil)
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

	return m.generateTLSConfFromKeyPair(secret[secretCrtName], secret[secretKeyName])
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

	m.k.StoreTLSCert(certPEM.String(), "atlas-tls-cert", atlasNamespace)

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivateKeyPEM.Bytes())
	if err != nil {
		return nil, err
	}

	rootCAs := bestEffortSystemCertPool()
	rootCAs.AppendCertsFromPEM(certPEM.Bytes())

	// TODO: guess it's a good idea to add certificates to the kubernetes secret
	//		and other servers will have to pull and use this certificate
	// TODO: main certificate should be added by workflowtrigger
	//		other servers should listen to secret and only start servers and clients
	//		when secret is reachable and cert is parsed and added to the pools
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
