package main

import (
    "keyholder/server"
    "keyholder/utils"
    "os"
    "os/signal"
)

func main()  {
    if len(os.Args) < 1 {
        utils.ReportError("usage:", os.Args[0], "[monitor-ip]")
    }

    // 识别Ctrl+C退出
    c := make(chan os.Signal,1)
    var sv *server.Server
    var err error
    signal.Notify(c,os.Interrupt)
    go func() {
        for _ = range c {
            cleanser(sv)
        }
    }()

    // 运行服务器
    sv, err = server.New(os.Args[1], 8957)
    if err != nil {
        utils.ReportError("Server could not be opened at this time:", err)
    }
    err = sv.Run()
    if err != nil {
        utils.ReportError("Server exited with error:", err)
    }
}

func cleanser(sv *server.Server) {
    utils.ReportWarning("You are interrupting server.")
    // 关闭服务器
    sv.Close()
    os.Exit(1)
}