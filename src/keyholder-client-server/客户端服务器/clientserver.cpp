

#include "stdafx.h"
#include "clientserver.h"

void sendnorm(UDPClientConn* conn, AESKey* aesKey, int recvTimeout, bool* isalive) {
	char* normBuff = new char[30];
	while (true) {
		(*conn).write("NORM", 4, aesKey);
		int ret;
		if ((ret = (*conn).read(normBuff, 1024, recvTimeout, aesKey)) < 0) {
			std::cout << "Error: Unknown reply for NORM signal\n";
			(*isalive) = false;
			break;
		}
		Sleep(30 * 1000);
	}
	delete[] normBuff;
}

int ClientServer::start() {
	UINT32 serverNewPort = 0;
	while (true) {
		// 新建会话
		if (isAlive == true) {
			conn.write("NEWS", 4);
			if (conn.read(recvBuffer, 1024, recvTimeout) < 0) {
				std::cout << "Error: Timed out when connecting to server: socket error\n";
				std::cout << "Info: Reconnecting...\n";
				Sleep(2000);
				continue;
			}
			// 将端口号设置为服务器的新端口号
			serverNewPort = *((UINT32 *)(recvBuffer + 5));
			conn.setServerPort((int)serverNewPort);
			std::cout << "Info: Connected." << std::endl;
			conn.write("OKAY", 4);
		}
		else {
			strcpy(sendBuffer, "NEWS ");
			memcpy(sendBuffer + 5, &serverNewPort, 4);
			conn.setServerPort(8957);
			while (isAlive == false) {
				std::cout << "Warning: Reconnecting...\n";
				conn.write(sendBuffer, 9);
				if (conn.read(recvBuffer, 1024, recvTimeout) > 0) {
					serverNewPort = *((UINT32 *)(recvBuffer + 5));
					conn.setServerPort((int)serverNewPort);
					std::cout << "Info: Connected." << std::endl;
					conn.write("OKAY", 4);
					isAlive = true;
					break;
				}
				Sleep(20 * 1000);
			}
		}
		// 向服务器发送OKAY，表示收到了端口号
		// 继续接收公钥
		if (conn.read(recvBuffer, 1024, recvTimeout) < 0) {
			std::cout << "Error: Timed out when connecting to server: waiting RSA key\n";
			isAlive = false;
			continue;
		}
		else if (recvBuffer[4] = 0, strcmp(recvBuffer, "PUBK") != 0) {
			std::cout << "Error: Stopping service because " << recvBuffer << " was received.\n";
			isAlive = false;
			continue;
		}
		conn.write("OKAY", 4);
		RSAPublicKey* pub = new RSAPublicKey; // 用于存储接收的公钥

		UINT32 pubKeyLen = *((UINT32 *)(recvBuffer + 5));
		pub->pubKeyLen = pubKeyLen;
		char *pubKey = new char[pubKeyLen];
		memcpy(pubKey, (char *)(recvBuffer + 10), pubKeyLen);
		pub->pubKey = pubKey;

		// 生成AES密钥
		AESKey* aesKey = GenAESKey();
		if (aesKey == NULL) {
			std::cout << "Error: AES key generator reports an error.\n";
			isAlive = false;
			continue;
		}
		// 用公钥加密密钥
		strcpy((char*)sendBuffer, "AESK ");
		memcpy(sendBuffer + 5, aesKey->aesKey, aesKey->aesKeyLen);
		EncData* enc = new EncData;
		enc->data = (unsigned char *)sendBuffer;
		enc->dataLen = 5 + aesKey->aesKeyLen;
		EncData* pubkAESK = RSAEncrypt(enc, pub);
		if (pubkAESK == NULL) {
			std::cout << "Error: Encrypring using AES key reports an error\n";
			isAlive = false;
			continue;
		}

		// 将加密后的密钥发送给服务器
		int ret;
		ret = conn.write((char *)pubkAESK->data, pubkAESK->dataLen);
		// 接收服务器的OKAY，表示服务器收到了密钥
		if ((ret = conn.read(recvBuffer, 1024, recvTimeout, aesKey)) < 0) {
			std::cout << "Error: Timed out while waiting server's OKAY on AESKey confirmed.\n";
			isAlive = false;
			continue;
		}

		DestroyData(pubkAESK);
		delete enc;
		delete[] pubKey;
		delete pub;

		// 如果服务器返回的不是OKAY，服务器出错，关闭客户端服务器
		recvBuffer[ret] = 0;
		if (strcmp((char *)recvBuffer, "OKAY") != 0) {
			std::cout << "Error: Stopping service because " << recvBuffer << " was received.\n";
			isAlive = false;
			continue;
		}

		// 如果服务器返回的是OKAY，则作为服务器，开始监听客户端的连入
		std::thread thNorm(sendnorm, &conn, aesKey, recvTimeout, &isAlive);
		// 收到APPL信息或者RETU信息
		// 用密钥对其进行加密并传输到服务器
		// 测试APPL

		std::cout << "Info: Awaiting clients's requests...\n";
		std::cout << "Warning: Wait up to 1,000 seconds" << std::endl;
		int count = 1;	// 目前在线之用户数
		sockaddr_in comingAddr;
		bool isAPPL = false;
		bool isRETU = false;
		do {
			ret = localConn.read(recvBuffer, 1024, 1000000, comingAddr);
			if (ret < 1) {
				std::cout << "Warning:Reading client's information throwed an error, or timed out." << std::endl;
				continue;
			}
			recvBuffer[ret] = 0x00;
			std::cout << "Info: Received a request" << std::endl;
			
			if (recvBuffer[4] = 0, strcmp(recvBuffer, "APPL") == 0) {
				isAPPL = true; isRETU = false;
			}
			else if (strcmp(recvBuffer, "RETU") == 0) {
				isRETU = true; isAPPL = false;
			}
			else {
				isAPPL = false;
				isRETU = false;
			}

			// 转发需求
			if ((ret = conn.write(recvBuffer, ret, aesKey)) < 0) {
				std::cout << "Error: Sending APPL not ok" << std::endl;
				return -1;
			}

			if ((ret = conn.read(recvBuffer, 1024, recvTimeout, aesKey)) < 0) {
				std::cout << "Error: Timed out when listening reply on APPL" << std::endl;
				return -1;
			}
			else if (recvBuffer[4] = 0, strcmp(recvBuffer, "OKAY") != 0 && strcmp(recvBuffer, "NOOK") != 0) {
				std::cout << "Error: Stopping service because " << recvBuffer << " was received.\n";
				return -1;
			}

			if (strcmp(recvBuffer, "OKAY") == 0) {
				if (isAPPL == true) {
					count++;
				}
				else if (isRETU == true) {
					count--;
					if (count == 1) {
						// 已经没有客户了，诱导客服自动退出
						count = 0;
					}
				}
				// 回复OKAY
				if ((ret = localConn.write(recvBuffer, ret, comingAddr)) < 0) {
					std::cout << "Error: Sending OKAY back to client not ok" << std::endl;
				}
			}
			else {
				// 回复NOOK，告诉用户及不准用户继续。
				if ((ret = localConn.write(recvBuffer, ret, comingAddr)) < 0) {
					std::cout << "Error: Sending NOOK back to client not ok" << std::endl;
				}
			}
		} while (count > 0);

		// 客服运行结束了
		closesocket(localConn.getso());
		if (isAlive) {
			conn.write("GBYE", 4, aesKey);
			if ((ret = conn.read(recvBuffer, 1024, recvTimeout, aesKey) < 0)) {
				std::cout << "Error: Receiving GBYE reply error; maybe timed out.";
				isAlive = false;
				thNorm.detach();
			}
			thNorm.detach();
			// 清除暂存
			DestroyAESKey(aesKey);
			std::cout << "Exiting..." << std::endl;
		}
		else {
			return -1;
		}
		break;
	}
	return 0;
}