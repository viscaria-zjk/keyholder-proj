#include "stdafx.h"
#include <iostream>
#include <WinSock2.h>
#include "clientserver.h"
#pragma comment(lib, "ws2_32.lib")

int main(int argc, const char* argv[]) {
	if (argc < 3) {
		std::cout << "Error: Invalid arguments. Usage: cs [server-ip] [server-port] (e.g. cs 112.126.86.188 8957)" << std::endl;
		return -1;
	}

	// 初始化
	WORD socketVersion = MAKEWORD(2, 2);
	WSADATA wsaData;
	if (WSAStartup(socketVersion, &wsaData) != 0) {
		std::cout << "Error: Initialise socket error.\n";
		return -1;
	}
	ClientServer cs(argv[1], atoi(argv[2]), 5000, "127.0.0.1", 9898);
	return cs.start();
}

