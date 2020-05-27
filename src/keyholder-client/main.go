package main

import (
    "key-client/ui"
    "log"
    "os"
    "strconv"
)

func main() {
    port := 9898
    var err error
    if len(os.Args) > 1 {
        port, err = strconv.Atoi(os.Args[1])
        if err != nil {
            log.Fatal("You must enter a valid port in digital format, or not provide, which uses 9898 by default.")
        }
    }
    ui.Start(port)
}

