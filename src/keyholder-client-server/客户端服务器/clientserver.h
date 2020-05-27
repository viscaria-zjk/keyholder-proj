#ifndef CLIENTSERVER_H
#define CLIENTSERVER_H

#include <iostream>
#include <thread>
#include "udpconn.hpp"


class ClientServer {
	UDPClientConn conn;
	UDPServerConn localConn;					// 和客户端通讯的conn
	const int recvTimeout = 20 * 1000;		// 设置接收信息超时时间为20s
	char* recvBuffer;
	char* sendBuffer;
	bool  isAlive;							// 标志服务器是否还活着

public:
	ClientServer(const char serverAddr[], int serverPort, int timeout, const char monitorAddr[], int monitorPort) :
		conn(serverAddr, serverPort), localConn(monitorAddr, monitorPort) {
		recvBuffer = new char[1024];
		sendBuffer = new char[1024];
		isAlive = true;
	}

	~ClientServer() {
		// 清除暂存
		WSACleanup();
		closesocket(conn.getso());
		closesocket(localConn.getso());

		delete[] recvBuffer;
		delete[] sendBuffer;
	}
	
	// 在主线程开启客户端实例
	int start();
};

#endif