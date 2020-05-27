
#ifndef libenctools_h
#define libenctools_h

#ifdef __cplusplus
extern "C" {
#endif

// 1. RSA公钥结构体
typedef struct {
    const char* pubKey;
    int pubKeyLen;
} RSAPublicKey;

// 2. RSA私钥结构体
typedef struct {
	const char* priKey;
    int priKeyLen;
} RSAPrivateKey;

// 3. RSA公私钥合体结构体
typedef struct {
    RSAPublicKey* pubKey;
    RSAPrivateKey* priKey;
} RSAKeys;

// 4. AES密钥结构体
typedef struct {
    unsigned char* aesKey;
    int aesKeyLen;
} AESKey;

// 5. 数据结构体
typedef struct {
    unsigned char* data;
    int dataLen;
} EncData;

// x) 生成一对RSA公私钥
extern RSAKeys* GenRSAKeys(void);

// a) 用RSA公钥加密
extern EncData* RSAEncrypt(EncData* srcData, RSAPublicKey* rsaPubKey);

// b) 用RSA私钥解密
extern EncData* RSADecrypt(EncData* srcData, RSAPrivateKey* rsaPriKey);

// c) 生成一个AES-128密钥
extern AESKey* GenAESKey(void);

// d) 用AES密钥加密
extern EncData* AESEncrypt(EncData* srcData, AESKey* aesKey);

// e) 用AES密钥解密
extern EncData* AESDecrypt(EncData* srcData, AESKey* aesKey);

// f) 销毁一批数据
extern void DestroyData(EncData* encData);

// g_1) 销毁一个RSA公钥
extern void DestroyRSAPubKey(RSAPublicKey* pubKey);

// g_2) 销毁一个RSA私钥
extern void DestroyRSAPriKey(RSAPrivateKey* priKey);

// g_3) 销毁一个RSA公私钥
extern void DestroyRSAKeys(RSAKeys* keys);

// h) 销毁一个AES密钥
extern void DestroyAESKey(AESKey* pubKey);

#ifdef __cplusplus
}
#endif

#endif /* libenctools_h */
