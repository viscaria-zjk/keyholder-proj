#ifndef CLIENTSERVER_H
#define CLIENTSERVER_H

#include <iostream>
#include <thread>
#include "udpconn.hpp"


class ClientServer {
	UDPClientConn conn;
	UDPServerConn localConn;					// �Ϳͻ���ͨѶ��conn
	const int recvTimeout = 20 * 1000;		// ���ý�����Ϣ��ʱʱ��Ϊ20s
	char* recvBuffer;
	char* sendBuffer;
	bool  isAlive;							// ��־�������Ƿ񻹻���

public:
	ClientServer(const char serverAddr[], int serverPort, int timeout, const char monitorAddr[], int monitorPort) :
		conn(serverAddr, serverPort), localConn(monitorAddr, monitorPort) {
		recvBuffer = new char[1024];
		sendBuffer = new char[1024];
		isAlive = true;
	}

	~ClientServer() {
		// ����ݴ�
		WSACleanup();
		closesocket(conn.getso());
		closesocket(localConn.getso());

		delete[] recvBuffer;
		delete[] sendBuffer;
	}
	
	// �����߳̿����ͻ���ʵ��
	int start();
};

#endif