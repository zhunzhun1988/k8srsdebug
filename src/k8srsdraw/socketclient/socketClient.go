package socketclient

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

const (
	INFOTYPE_NODEINFO        string = "1"
	INFOTYPE_MESSAGE         string = "2"
	INFOTYPE_RESCHEDULE_OK   string = "3"
	INFOTYPE_RESCHEDULE_FAIL string = "4"
)

type PodInfos struct {
	Name      string
	Namespace string
}

type NodeInfos struct {
	NodeName string
	PodInfos []PodInfos
}

type Infos map[string]*NodeInfos

type EventHandle interface {
	Init(infos Infos)
	AddNode(nodeName string)
	DeleteNode(nodeName string)
	AddPod(nodeName, podNamespace, podName string)
	DeletePod(nodeName, podNamespace, podName string)
	ReschedulePod(fromNodeName, toNodeName, podNamespace, podName string)
	GetCurNodeInfos() Infos
}

type SClient struct {
	host        string
	port        string
	conn        net.Conn
	eventHandle EventHandle
	isFirstRun  bool
	//infos       Infos
}

func NewSClient(host, port string, eventHandle EventHandle) *SClient {
	return &SClient{
		host:        host,
		port:        port,
		eventHandle: eventHandle,
		isFirstRun:  true,
		//infos:       nil,
	}
}
func getCommonPodInfos(nodeInfos1, nodeInfos2 []PodInfos) (node1Only, node2Only, common []PodInfos) {
	node1Only = make([]PodInfos, 0)
	node2Only = make([]PodInfos, 0)
	common = make([]PodInfos, 0)
	for _, podInfo1 := range nodeInfos1 {
		flag := false
		for _, podInfo2 := range nodeInfos2 {
			if podInfo1.Namespace == podInfo2.Namespace && podInfo1.Name == podInfo2.Name {
				common = append(common, podInfo1)
				flag = true
				break
			}
		}
		if flag == false {
			node1Only = append(node1Only, podInfo1)
		}
	}
	for _, podInfo2 := range nodeInfos2 {
		flag := false
		for _, podInfo1 := range nodeInfos1 {
			if podInfo1.Namespace == podInfo2.Namespace && podInfo1.Name == podInfo2.Name {
				flag = true
				break
			}
		}
		if flag == false {
			node2Only = append(node2Only, podInfo2)
		}
	}
	return
}

/*
func (sc *SClient) AddPod(nodename, podNamespace, podName string) {
	node, find := sc.infos[nodename]
	if find {
		for _, pod := range node.PodInfos {
			if pod.Name == podName && pod.Namespace == podNamespace {
				return
			}
		}
		node.PodInfos = append(node.PodInfos, PodInfos{Name: podName, Namespace: podNamespace})
	}
}
func (sc *SClient) DeletePod(nodename, podNamespace, podName string) {
	node, find := sc.infos[nodename]
	if find {
		for i, pod := range node.PodInfos {
			if pod.Name == podName && pod.Namespace == podNamespace {
				if i == 0 {
					node.PodInfos = node.PodInfos[1:]
				} else if i == len(node.PodInfos)-1 {
					node.PodInfos = node.PodInfos[0 : len(node.PodInfos)-1]
				} else {
					node.PodInfos = append(node.PodInfos[0:i], node.PodInfos[i+1:]...)
				}
				return
			}
		}

	}
}*/
func (sc *SClient) CompareInfo(newInfos Infos) {
	oldInfos := sc.eventHandle.GetCurNodeInfos()
	for _, newInfo := range newInfos {
		if oldInfo, ok := oldInfos[newInfo.NodeName]; !ok {
			sc.eventHandle.AddNode(newInfo.NodeName)
			for _, podInfo := range newInfo.PodInfos {
				sc.eventHandle.AddPod(newInfo.NodeName, podInfo.Namespace, podInfo.Name)
			}
		} else {
			t1, t2, _ := getCommonPodInfos(newInfo.PodInfos, oldInfo.PodInfos)
			for _, p2 := range t2 {
				sc.eventHandle.DeletePod(newInfo.NodeName, p2.Namespace, p2.Name)
			}
			for _, p1 := range t1 {
				sc.eventHandle.AddPod(newInfo.NodeName, p1.Namespace, p1.Name)

			}
		}
	}
	for _, node := range oldInfos {
		if _, ok := newInfos[node.NodeName]; !ok {
			sc.eventHandle.DeleteNode(node.NodeName)
		}
	}
}
func (sc *SClient) handleMessage(msg string) {
	strs := strings.Split(msg, "->")
	if len(strs) == 2 {
		msgStr := strs[1]
		switch strs[0] {
		case INFOTYPE_NODEINFO:
			infos := make(map[string]*NodeInfos)
			fmt.Printf("msg type INFOTYPE_NODEINFO %s\n", msgStr)
			json.Unmarshal([]byte(strs[1]), &infos)
			if sc.isFirstRun {
				sc.isFirstRun = false
				sc.eventHandle.Init(infos)
			} else {
				sc.CompareInfo(infos)
			}
		case INFOTYPE_RESCHEDULE_OK:
			names := strings.Split(strs[1], ":")

			sc.eventHandle.ReschedulePod(names[2], names[3], names[0], names[1])
			//sc.AddPod(names[3], names[0], names[1])
			//sc.DeletePod(names[2], names[0], names[1])
			fmt.Println("--------------")
			fmt.Printf("reschedule pod %s:%s from %s to %s Success %s\n", names[0], names[1], names[2], names[3], names[4])
			/*n1, find1 := sc.infos[names[3]]
			n2, find2 := sc.infos[names[2]]

			fmt.Println("node ", names[3])
			fmt.Println(n1.PodInfos)
			fmt.Println("node ", names[2])
			fmt.Println(n2.PodInfos)*/
			fmt.Println("++++++++++++++")
		case INFOTYPE_RESCHEDULE_FAIL:
			names := strings.Split(strs[1], ":")
			fmt.Printf("reschedule pod %s:%s from %s to %s fail %s\n", names[0], names[1], names[2], names[3], names[4])
		case INFOTYPE_MESSAGE:
			//fmt.Printf("msg type INFOTYPE_MESSAGE:%s\n", msgStr)
		}
	} else {
		fmt.Printf("unknow msg %d from server\n", strs[0])
	}
}
func (sc *SClient) Run() {
	con, err := net.Dial("tcp", sc.host+":"+sc.port)
	defer con.Close()

	if err != nil {
		fmt.Println("Server not found.")
		return
	}
	fmt.Println("Connection OK.")

	msg := make([]byte, 80960)
	for {

		length, err := con.Read(msg)
		if err != nil {
			fmt.Printf("Error when read from server. err=%v\n", err)
			return
		}
		strs := strings.Split(string(msg[0:length]), "#")
		for _, str := range strs {
			if str != "" {
				sc.handleMessage(str)
			}
		}
	}

}
