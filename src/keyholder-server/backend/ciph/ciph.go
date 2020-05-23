package ciph

import (
    "bytes"
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "encoding/pem"
    "errors"
)

// 生成RSA公私钥（输入：密钥的比特数、保存密钥至的目录）
func GenRSAKeys(bits int) (pubKey []byte, priKey []byte, err error) {
    keys, err := rsa.GenerateKey(rand.Reader, bits)
    if err != nil {
        return nil, nil, err
    }
    // 保存私钥（pem x509格式）
    derStream := x509.MarshalPKCS1PrivateKey(keys)
    block := &pem.Block {
        Type:  "RSA PRIVATE KEY",
        Bytes: derStream,
    }
    priKey = pem.EncodeToMemory(block)
    // 保存公钥（pem x509）
    publicKey := &keys.PublicKey
    derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
    if err != nil {
        return nil, nil, err
    }
    block = &pem.Block{
        Type:  "PUBLIC KEY",
        Bytes: derPkix,
    }
    pubKey = pem.EncodeToMemory(block)
    if pubKey == nil || priKey == nil {
        return nil, nil, errors.New("generate keys error")
    }
    return pubKey, priKey, err
}

// RSA解密（输入：私钥文件存储的目录、密文，输出：明文）
func RSADecrypt(encodedBits []byte, priKey []byte) (decoded []byte) {
    // pem及X509解码解码
    block, _ := pem.Decode(priKey)
    privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
    if err!=nil{
        panic(err)
    }
    // 对密文进行解密
    plainText,_:=rsa.DecryptPKCS1v15(rand.Reader, privateKey, encodedBits)
    // 返回明文
    return plainText
}

// AES解密
func AESDecrypt(orig, key[] byte) []byte {
    block, _ := aes.NewCipher(key)
    blockSize := block.BlockSize()
    blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
    origData := make([]byte, len(orig))
    blockMode.CryptBlocks(origData, orig)
    // 去补码（AES解密用）
    unPadding := int(origData[len(origData) - 1])
    return origData[:len(origData) - unPadding]
}

// AES加密
func AESEncrypt(origData, key []byte) []byte {
    // 新建加密器及补码
    block, _ :=aes.NewCipher(key)
    blockSize := block.BlockSize()
    // 计算需要补几位数
    padding := blockSize - len(origData) % blockSize
    //fmt.Println(padding)
    // 在切片后面追加char数量的byte(char) 	切片后面的第一个字符是增加的padding数量
    origData = append(origData, bytes.Repeat([]byte { byte(padding) }, padding)...)
    // 设置加密格式
    blockMode := cipher.NewCBCEncrypter(block,key[:blockSize])
    // 创建缓冲区
    encrypted := make([]byte, len(origData))
    // 开始加密
    blockMode.CryptBlocks(encrypted, origData)
    return encrypted
}

