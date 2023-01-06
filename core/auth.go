package core

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// 生成rsa公私钥
func GetKeyPair() (prvKey, pubKey []byte) {
	// 生成私钥文件
	//tmp :=  //io.Reader
	//i:=rand.Reader
	//j:=strings.NewReader(seed)
	//l:=new(Reader)

	//l.prov=seed
	//fmt.Println(l)
	//l.Read()
	//strings.Reader{seed}
	privateKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		panic(err)
	}
	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derStream,
	}
	prvKey = pem.EncodeToMemory(block)
	publicKey := &privateKey.PublicKey
	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		panic(err)
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derPkix,
	}
	pubKey = pem.EncodeToMemory(block)

	return
}

//以文件形式存储密钥对
func GenRsaKeys() {
	RsaPv, RsaPb := GetKeyPair()
	//privFileName := "RSA_PIV"
	file1, err := os.OpenFile("RSA_PIV", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer file1.Close()
	_, _ = file1.Write(RsaPv)
	//pubFileName := "RSA_PUB"
	file2, err := os.OpenFile("RSA_PUB", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer file2.Close()
	_, _ = file2.Write(RsaPb)
	fmt.Println("KEYGEN completed.")
}

//获取对应的公钥
func GetPubKey(id ...string) []byte {
	var key []byte
	var err error
	if len(id) == 0 {
		key, err = ioutil.ReadFile("RSA_PUB")
	} else {
		if isExist("../ki/" + id[0] + "PUB") {
			key, err = ioutil.ReadFile("../ki/" + id[0] + "PUB")
		} else {
			return nil
		}
	}
	//fmt.Println("./:"+id[0]+"PUB")

	if err != nil {
		log.Panic(err)
	}
	return key
}

//获取对应的私钥
func GetPrivKey() []byte {
	key, err := ioutil.ReadFile("RSA_PIV")
	if err != nil {
		log.Panic(err)
	}
	return key
}

// 数字签名
func RsaSignWithSha256(data []byte, keyBytes []byte) []byte {
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)
	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic(errors.New("private key error"))
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Println("ParsePKCS8PrivateKey err", err)
		panic(err)
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		fmt.Printf("Error from signing: %s\n", err)
		panic(err)
	}

	return signature
}

// 签名验证
func RsaVerySignWithSha256(data, signData, keyBytes []byte) bool {

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic(errors.New("public key error"))
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	hashed := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(pubKey.(*rsa.PublicKey), crypto.SHA256, hashed[:], signData)
	if err != nil {
		panic(err)
	}

	return true
}

// 判断文件或文件夹是否存在, isExist
func isExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		fmt.Println(err)
		return false
	}
	return true
}
