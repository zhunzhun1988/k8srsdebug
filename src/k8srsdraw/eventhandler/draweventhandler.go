package eventhandler

import (
	"image/color"
	"k8srsdraw/socketclient"
	"k8srsdraw/window"
)

type DrawEventHandle struct {
	w *window.Window
}

func NewDrawEventHandle(w, h int) *DrawEventHandle {
	return &DrawEventHandle{
		w: window.NewWindow(w, h, color.RGBA{0x00, 0x00, 0x00, 0xff}),
	}
}

func (deh *DrawEventHandle) Init(infos socketclient.Infos) {
	for _, nodeInfo := range infos {
		deh.w.AddNode(nodeInfo.NodeName)

		for _, podInfo := range nodeInfo.PodInfos {
			deh.w.AddPod(nodeInfo.NodeName, podInfo.Namespace, podInfo.Name)
		}
	}
}
func (deh *DrawEventHandle) AddNode(nodeName string) {
	deh.w.AddNode(nodeName)
}
func (deh *DrawEventHandle) DeleteNode(nodeName string) {
	deh.w.DeleteNode(nodeName)
}
func (deh *DrawEventHandle) AddPod(nodeName, podNamespace, podName string) {
	deh.w.AddPod(nodeName, podNamespace, podName)
}
func (deh *DrawEventHandle) DeletePod(nodeName, podNamespace, podName string) {
	deh.w.DeletePod(nodeName, podNamespace, podName)
}

func (deh *DrawEventHandle) ReschedulePod(fromNodeName, toNodeName, podNamespace, podName string) {
	deh.w.MovePodFromTo(fromNodeName, toNodeName, podNamespace, podName)
}

func (deh *DrawEventHandle) GetCurNodeInfos() socketclient.Infos {
	ret := make(map[string]*socketclient.NodeInfos)
	nodes := deh.w.GetNodeList()
	for _, node := range nodes {
		t := &socketclient.NodeInfos{NodeName: node.Name,
			PodInfos: make([]socketclient.PodInfos, 0),
		}
		ret[node.Name] = t
		for _, p := range node.Pods {
			for name, _ := range p.Names {
				t.PodInfos = append(t.PodInfos,
					socketclient.PodInfos{Name: name, Namespace: p.Namespace})
			}

		}
	}
	return ret
}
