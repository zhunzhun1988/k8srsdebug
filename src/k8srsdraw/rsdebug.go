package main

import (
	"fmt"
	"k8srsdraw/eventhandler"
	"k8srsdraw/socketclient"
	"os"
	"time"
)

func main() {
	var sip = "10.19.132.220"
	if len(os.Args) == 2 {
		sip = os.Args[1]
	} else if len(os.Args) != 1 {
		fmt.Printf("%s serverip\nFor Example: %s 127.0.0.1\n", os.Args[0], os.Args[0])
		os.Exit(-1)
	}
	deh := eventhandler.NewDrawEventHandle(800, 400)
	for {
		sc := socketclient.NewSClient(sip, "8888", deh)
		sc.Run()
		time.Sleep(3 * time.Second)
	}
}
