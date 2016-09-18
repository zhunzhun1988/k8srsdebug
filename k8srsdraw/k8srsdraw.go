package main

import (
	"fmt"
	"image/color"
	"k8srsdraw/window"
	"time"
)

var (
	drawWight  = 1300
	drawHeight = 800
)

func main() {
	w := window.NewWindow(drawWight, drawHeight, color.RGBA{0x00, 0x00, 0x00, 0xff})
	for i := 0; i < 5; i++ {
		w.AddNode(fmt.Sprintf("node %d", +i))
	}
	for n := 0; n < 5; n++ {
		for i := 1; i < 5; i++ {
			for c := 0; c < 1; c++ {
				w.AddPod(fmt.Sprintf("node %d", n), fmt.Sprintf("testnamespace%d", +i), "node")
				//time.Sleep(200 * time.Millisecond)
			}
		}
	}

	time.Sleep(3 * time.Second)
	w.MovePodFromTo("node 4", "node 0", "testnamespace1", "node")
	w.MovePodFromTo("node 4", "node 1", "testnamespace2", "node")
	w.MovePodFromTo("node 4", "node 2", "testnamespace3", "node")
	w.MovePodFromTo("node 4", "node 3", "testnamespace4", "node")
	w.MovePodFromTo("node 3", "node 0", "testnamespace1", "node")
	w.MovePodFromTo("node 3", "node 1", "testnamespace2", "node")
	w.MovePodFromTo("node 3", "node 2", "testnamespace3", "node")
	w.MovePodFromTo("node 3", "node 4", "testnamespace5", "node")
	w.MovePodFromTo("node 2", "node 0", "testnamespace1", "node")
	w.MovePodFromTo("node 2", "node 1", "testnamespace2", "node")
	w.MovePodFromTo("node 2", "node 4", "testnamespace5", "node")
	w.MovePodFromTo("node 2", "node 3", "testnamespace4", "node")
	w.MovePodFromTo("node 1", "node 0", "testnamespace1", "node")
	w.MovePodFromTo("node 1", "node 2", "testnamespace3", "node")
	w.MovePodFromTo("node 1", "node 4", "testnamespace5", "node")
	w.MovePodFromTo("node 1", "node 3", "testnamespace4", "node")
	w.MovePodFromTo("node 0", "node 2", "testnamespace3", "node")
	w.MovePodFromTo("node 0", "node 1", "testnamespace2", "node")
	w.MovePodFromTo("node 0", "node 4", "testnamespace5", "node")
	w.MovePodFromTo("node 0", "node 3", "testnamespace4", "node")
	/*w.MovePodFromTo("node 1", "node 2", "testnamespace1", "node")
	w.MovePodFromTo("node 4", "node 0", "testnamespace1", "node")
	w.MovePodFromTo("node 3", "node 1", "testnamespace1", "node")
	w.MovePodFromTo("node 5", "node 0", "testnamespace0", "node")
	w.MovePodFromTo("node 2", "node 0", "testnamespace2", "node")
	w.MovePodFromTo("node 1", "node 3", "testnamespace3", "node")
	w.MovePodFromTo("node 5", "node 1", "testnamespace3", "node")
	w.MovePodFromTo("node 4", "node 2", "testnamespace2", "node")
	w.MovePodFromTo("node 0", "node 4", "testnamespace3", "node")*/
	w.WaitEvent()
}
