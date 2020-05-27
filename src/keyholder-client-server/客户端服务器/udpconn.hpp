#ifndef UDPCONN_H
#define UDPCONN_H
#include <WinSock2.h>
#include <libenc.h>

// ��װ��UDP�ͻ�����
class UDPClientConn {
	SOCKET so;
	sockaddr_in serverAddr;
	int addrLen;

	UDPClientConn() {
		this->so = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
	}
public:
	UDPClientConn(const char* serverAddr, int port) {
		this->so = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
		this->serverAddr.sin_family = AF_INET;
		this->serverAddr.sin_port = htons(port);
		this->serverAddr.sin_addr.S_un.S_addr = inet_addr(serverAddr);
		this->addrLen = sizeof(this->serverAddr);
	}
	SOCKET getso() {
		return so;
	}

	// ��ȡSocket������UDP
	int read(char* buffer, int maxLength, int timeOut, AESKey* aesKey = NULL) {
		if (timeOut != 0) {		// ����Ҫtimeout
			if (SOCKET_ERROR == setsockopt(so, SOL_SOCKET, SO_RCVTIMEO, (char *)&timeOut, sizeof(int))) {
				return -1;
			}
		}

		int ret = recvfrom(so, buffer, maxLength, 0, (sockaddr *)&serverAddr, &addrLen);
		if (aesKey != NULL) {
			if (ret <= 0) {
				// ��ʱ �����ʧ��
				return -1;
			}
			// ����׼��������
			EncData dec;
			dec.data = (unsigned char*)buffer;
			dec.dataLen = ret;
			EncData* decoded = AESDecrypt(&dec, aesKey);
			if (decoded == NULL) {
				return 0;
			}
			// �������ܺ�����
			memcpy(buffer, decoded->data, decoded->dataLen);
			ret = decoded->dataLen;
			DestroyData(decoded);
		}
		buffer[ret] = 0;
		return ret;
	}

	// ����Socket������UDP
	int write(const char* buffer, int length, AESKey* aesKey = NULL) {
		if (aesKey != NULL) {	// ��Ҫ����
			// ����׼��������
			EncData enc;
			enc.data = (unsigned char*)buffer;
			enc.dataLen = length;
			EncData* aesEnc = AESEncrypt(&enc, aesKey);
			if (aesKey == NULL) {
				return -1;
			}
			// ����
			int ret = sendto(so, (char *)aesEnc->data, aesEnc->dataLen, 0, (sockaddr *)&serverAddr, addrLen);
			// ������ܺ������
			DestroyData(aesEnc);
			return ret;
		}
		else {					// ����Ҫ����
			return sendto(so, (char *)buffer, length, 0, (sockaddr *)&serverAddr, addrLen);
		}
	}

	
	void setServerAddr(const char* addr) {
		this->serverAddr.sin_addr.S_un.S_addr = inet_addr(addr);
		this->addrLen = sizeof(this->serverAddr);
	}

	void setServerPort(int port) {
		this->serverAddr.sin_port = htons(port);
		this->addrLen = sizeof(this->serverAddr);
	}
};

// ��װ��UDP����������
class UDPServerConn {
	friend void clientSession(UDPServerConn* localConn, UDPClientConn* conn, const char* startRequest, int startReqLen, AESKey* aesKey, int timeout, sockaddr_in* clientAddr);

	SOCKET so;
	sockaddr_in serverAddr;
	int serverAddrLen;

	UDPServerConn() {
		this->so = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
	}
public:
	UDPServerConn(const char* monitorAddr, int monitorPort) {
		// ����һ����ַ
		this->so = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP);
		this->serverAddr.sin_family = AF_INET;
		this->serverAddr.sin_port = htons(monitorPort);
		this->serverAddr.sin_addr.S_un.S_addr = inet_addr(monitorAddr);
		this->serverAddrLen = sizeof(this->serverAddr);
		// ��
		bind(this->so, (sockaddr*)&(this->serverAddr), this->serverAddrLen);
	}
	SOCKET getso() {
		return so;
	}

	// ��ȡSocket������UDP
	int read(char* buffer, int maxLength, int timeOut, sockaddr_in& comingAddr) {
		if (timeOut != 0) {		// ����Ҫtimeout
			if (SOCKET_ERROR == setsockopt(so, SOL_SOCKET, SO_RCVTIMEO, (char *)&timeOut, sizeof(int))) {
				return -1;
			}
		}
		int addrLen = sizeof(comingAddr);
		int ret = recvfrom(so, buffer, maxLength, 0, (sockaddr *)&comingAddr, &addrLen);
		if (ret <= 0) {
			// ��ʱ �����ʧ��
			return -1;
		}
		buffer[ret] = 0;
		
		return ret;
	}

	// ����Socket������UDP
	int write(const char* buffer, int length, sockaddr_in& comingAddr) {
		return sendto(so, (char *)buffer, length, 0, (sockaddr *)&comingAddr, sizeof(comingAddr));
	}

	void setServerAddr(const char* addr) {
		this->serverAddr.sin_addr.S_un.S_addr = inet_addr(addr);
		this->serverAddrLen = sizeof(this->serverAddr);
	}

	void setServerPort(int port) {
		this->serverAddr.sin_port = htons(port);
		this->serverAddrLen = sizeof(this->serverAddr);
	}
};

#endif