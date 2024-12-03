package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"log"
	"os"

	"github.com/fullsailor/pkcs7"
)

func CreateEmbeddedPKCS7Signature(data []byte, certPEM []byte, keyPEM []byte) ([]byte, error) {

	block, _ := pem.Decode(keyPEM)

	if block == nil || block.Type != "PRIVATE KEY" && block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("invalid private key format")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return nil, errors.New("failed to parse RSA private key")
	}

	block, _ = pem.Decode(certPEM)

	if block == nil || block.Type != "CERTIFICATE" {
		return nil, errors.New("invalid certificate format")
	}

	cert, err := x509.ParseCertificate(block.Bytes)

	if err != nil {
		return nil, errors.New("failed to parse X509 certificate")
	}

	signature, err := pkcs7.NewSignedData(data)

	if err != nil {
		return nil, errors.New("failed to create signed data")
	}

	err = signature.AddSigner(cert, privateKey, pkcs7.SignerInfoConfig{})

	if err != nil {
		return nil, errors.New("failed to add signer to PKCS7")
	}

	signature.Detach()

	return signature.Finish()
}

func CreateKey(file *os.File) error {

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		return err
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	_, err = file.Write(privateKeyPEM)

	return err
}

func CreateCSR(
	keyFile *os.File,
	organizationName, commonName, serialNumber string,
	outFile *os.File,
) error {

	keyData := make([]byte, 2048)

	n, err := keyFile.Read(keyData)

	if err != nil {
		return err
	}

	block, _ := pem.Decode(keyData[:n])

	if block == nil || block.Type != "PRIVATE KEY" && block.Type != "RSA PRIVATE KEY" {
		return errors.New("invalid private key format")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	if err != nil {
		return err
	}

	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			Organization: []string{organizationName},
			CommonName:   commonName,
			SerialNumber: serialNumber,
		},
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, template, privateKey)

	if err != nil {
		return err
	}

	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	})

	_, err = outFile.Write(csrPEM)

	return err
}

func GenerateKey() {

	data := []byte("Example data to sign")

	certPEM := []byte("...") // Provide your certificate PEM content here

	keyPEM := []byte("...") // Provide your private key PEM content here

	signedData, err := CreateEmbeddedPKCS7Signature(data, certPEM, keyPEM)

	if err != nil {
		log.Fatalf("Error creating PKCS7 signature: %v", err)
	}

	log.Printf("PKCS7 Signature created successfully: %x\n", signedData)

	keyFile, err := os.Create("private_key.pem")

	if err != nil {
		log.Fatalf("Error creating private key file: %v", err)
	}

	defer keyFile.Close()

	err = CreateKey(keyFile)

	if err != nil {
		log.Fatalf("Error generating private key: %v", err)
	}

	log.Println("Private key generated successfully")

	outFile, err := os.Create("csr.pem")

	if err != nil {
		log.Fatalf("Error creating CSR file: %v", err)
	}

	defer outFile.Close()

	err = CreateCSR(keyFile, "My Organization", "example.com", "123456", outFile)

	if err != nil {
		log.Fatalf("Error generating CSR: %v", err)
	}

	log.Println("CSR created successfully")
}
