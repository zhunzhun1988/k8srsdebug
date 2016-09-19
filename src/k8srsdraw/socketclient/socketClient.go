package socketclient

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type EventHandle interface {
	Init(infos Infos)
	AddNode(nodeName string)
	DeleteNode(nodeName string)
	AddPod(nodeName, podNamespace, podName string)
	DeletePod(nodeName, podNamespace, podName string)
	ReschedulePod(fromNodeName, toNodeName, podNamespace, podName string)
}

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

type SClient struct {
	host        string
	port        string
	conn        net.Conn
	eventHandle EventHandle
	infos       Infos
}

func NewSClient(host, port string, eventHandle EventHandle) *SClient {
	return &SClient{
		host:        host,
		port:        port,
		eventHandle: eventHandle,
		infos:       nil,
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
func (sc *SClient) CompareInfo(newInfos Infos) {
	fmt.Printf("enter CompareInfo size =%d\n", len(newInfos))
	for _, newInfo := range newInfos {
		fmt.Printf("newInfo:%s %d, %d\n", newInfo.NodeName, len(newInfo.PodInfos), len(newInfo.PodInfos))
		if oldInfo, ok := sc.infos[newInfo.NodeName]; !ok {
			sc.eventHandle.AddNode(newInfo.NodeName)
			for _, podInfo := range newInfo.PodInfos {
				fmt.Printf("add pod %s %s:%s\n", newInfo.NodeName, podInfo.Namespace, podInfo.Name)
				sc.eventHandle.AddPod(newInfo.NodeName, podInfo.Namespace, podInfo.Name)
			}
		} else {
			t1, t2, _ := getCommonPodInfos(newInfo.PodInfos, oldInfo.PodInfos)
			for _, p1 := range t1 {
				sc.eventHandle.AddPod(newInfo.NodeName, p1.Namespace, p1.Name)
			}
			for _, p2 := range t2 {
				sc.eventHandle.DeletePod(newInfo.NodeName, p2.Namespace, p2.Name)
			}
		}

	}
	for _, node := range sc.infos {
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
			if sc.infos == nil {
				sc.infos = infos
				sc.eventHandle.Init(sc.infos)
			} else {
				sc.CompareInfo(infos)
				sc.infos = infos
			}
		case INFOTYPE_RESCHEDULE_OK:
			names := strings.Split(strs[1], ":")
			node1, find1 := sc.infos[names[0]]
			node2, find2 := sc.infos[names[1]]
			if find1 && find2 {
				for i, pod := range node1.PodInfos {
					if pod.Namespace == names[2] && pod.Name == names[3] {
						fmt.Printf("befor delete pod %s:%s: %v\v", pod.Namespace, pod.Name, node1.PodInfos)
						if i == 0 {
							node1.PodInfos = node1.PodInfos[1:]
						} else if i == len(node1.PodInfos)-1 {
							node1.PodInfos = node1.PodInfos[0:i]
						} else {
							node1.PodInfos = append(node1.PodInfos[0:i], node1.PodInfos[i+1:]...)
						}
						fmt.Printf("after delete pod %s:%s: %v\v", pod.Namespace, pod.Name, node1.PodInfos)
						break
					}
				}
				node2.PodInfos = append(node2.PodInfos, PodInfos{Namespace: names[2], Name: names[3]})
			}
			sc.eventHandle.ReschedulePod(names[2], names[3], names[0], names[1])
			fmt.Printf("reschedule pod %s:%s from %s to %s Success %s\n", names[0], names[1], names[2], names[3], names[4])
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
	//go func() {
	con, err := net.Dial("tcp", sc.host+":"+sc.port)
	defer con.Close()

	if err != nil {
		fmt.Println("Server not found.")
		os.Exit(-1)
	}
	fmt.Println("Connection OK.")

	msg := make([]byte, 1024)
	for {

		length, err := con.Read(msg)
		if err != nil {
			fmt.Printf("Error when read from server. err=%v\n", err)
			os.Exit(0)
		}
		strs := strings.Split(string(msg[0:length]), "#")
		for _, str := range strs {
			if str != "" {
				sc.handleMessage(str)
			}
		}
	}
}
