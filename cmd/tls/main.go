package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"
)

func main() {
	subject := pkix.Name{
		Country:            []string{"Country"},
		Organization:       []string{"Organization"},
		OrganizationalUnit: []string{"OrganizationalUnit"},
		Locality:           []string{"Locality"},
		Province:           []string{"Province"},
	}
	// CA配置
	ca := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),             // 证书序列号，对同一CA所颁发的证书，序列号唯一标识证书
		Issuer:       pkix.Name{Country: []string{"IssuerCountry"}}, // 证书发行者名称(颁发者)
		Subject:      subject,                                       // 证书使用者
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, // 指定了这份证书包含的公钥可以执行的密码操作，例如只能用于签名，但不能用来加密。​
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth}, // 典型用法是指定叶子证书中的公钥的使用目的。它包括一系列的OID，每一个都指定一种用途。例如{id pkix 31}表示用于服务器端的TLS/SSL连接；{id pkix 34}表示密钥可以用于保护电子邮件。​ ​通常情况下，当一份证书有多个限制用途的扩展时，所有限制条件都应该满足才可以使用。RFC 5280有一个例子，该证书同时含有keyUsage和extendedKeyUsage，这样的证书只能用在被这两个扩展指定的用途，例如网络安全服务决定证书用途时，会同时对这个扩展进行判断。
		BasicConstraintsValid: true, // 用于指示一份证书是不是CA证书
		IsCA:                  true,
		MaxPathLen:            0,   // 0保证在中间CA下面不能有其他证书颁发机构
		SignatureAlgorithm:    0,   // 签名算法
		PublicKeyAlgorithm:    0,   // 公钥算法
		PublicKey:             nil, // 主题公钥
		Version:               0,   // 版本号
		Extensions:            nil, // 证书的扩展项（可选）
	}
	// 生成CA私钥
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Printf("GenerateKey err:%v", err)
		return
	}
	// 创建ca证书
	caBytes, err := x509.CreateCertificate(rand.Reader, &ca, &ca, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		log.Printf("CreateCertificate err:%v", err)
		return
	}

	//	编码证书文件
	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		log.Printf("pem.Encode err:%v", err)
		return
	}
	caCert, err := x509.ParseCertificate(caBytes)
	if err != nil {
		log.Printf("x507 ParseCertificate ca err:%v", err)
		return
	}
	caKeyPEM := new(bytes.Buffer)
	err = pem.Encode(caKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivateKey),
	})
	if err != nil {
		log.Printf("ca key pem.Encode err:%v", err)
		return
	}

	//	服务端证书证书配置
	serverCert := x509.Certificate{
		DNSNames: []string{
			"pod-dmission",
			"pod-dmission.default",
			"pod-dmission.default.svc",
			"pod-dmission.default.svc.cluster.local",
		},
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "commonNae", Organization: []string{"China Go"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}

	//	生成服务端私钥
	serverPriKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Printf("serverPriKey GenerateKey err:%v", err)
		return
	}

	serverCertBytes, err := x509.CreateCertificate(rand.Reader, &serverCert, caCert, &serverPriKey.PublicKey, caPrivateKey)
	if err != nil {
		log.Printf("servercert GenerateKey err:%v", err)
		return
	}
	serverCertPEM := new(bytes.Buffer)
	err = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})
	if err != nil {
		log.Printf("servercert Encode err:%v", err)
		return
	}
	serverCertKeyPEM := new(bytes.Buffer)
	err = pem.Encode(serverCertKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPriKey),
	})
	if err != nil {
		log.Printf("servercert key Encode err:%v", err)
		return
	}
	err = os.MkdirAll("goCert", 0755)
	if err != nil {
		log.Printf("MkdirAll err:%v", err)
	}
	err = os.WriteFile("goCert/ca.crt", caPEM.Bytes(), 0600)
	if err != nil {
		log.Printf("ca.crt file err:%v", err)
		return
	}
	err = os.WriteFile("goCert/ca-key.crt", caKeyPEM.Bytes(), 0600)
	if err != nil {
		log.Printf("ca-key.crt file err:%v", err)
		return
	}
	err = os.WriteFile("goCert/server.crt", serverCertPEM.Bytes(), 0600)
	if err != nil {
		log.Printf("server.crt file err:%v", err)
		return
	}
	err = os.WriteFile("goCert/server-key.crt", serverCertKeyPEM.Bytes(), 0600)
	if err != nil {
		log.Printf("server-key.crt file err:%v", err)
		return
	}
}
