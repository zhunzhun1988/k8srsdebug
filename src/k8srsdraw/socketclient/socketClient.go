package socketclient

import (
	"encoding/json"
	"fmt"
	"k8srsdraw/workqueue"
	"net"
	"strings"
	"time"
)

const (
	INFOTYPE_NODEINFO                      string = "1"
	INFOTYPE_MESSAGE                       string = "2"
	INFOTYPE_RESCHEDULE_OK                 string = "3"
	INFOTYPE_RESCHEDULE_FAIL               string = "4"
	INFOTYPE_RESCHEDULE_STARTONERESCHEDULE string = "5"
	INFOTYPE_RESCHEDULE_STOPONERESCHEDULE  string = "6"
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
	ReschedulePod(fromNodeName, toNodeName, podNamespace, fromPodName, toPodName string)
	GetCurNodeInfos() Infos
}

type SClientWorkItem struct {
	workqueue.WorkItem
	cmd      string
	id       string
	scClient *SClient
}

func NewSClientWorkItem(id, str string, c *SClient) *SClientWorkItem {
	return &SClientWorkItem{
		cmd:      str,
		scClient: c,
		id:       id,
	}
}
func (wi *SClientWorkItem) GetID() string {
	return wi.id
}
func (wi *SClientWorkItem) Run() {
	wi.scClient.handleMessage(wi.id, wi.cmd)
}

type SClient struct {
	host        string
	port        string
	conn        net.Conn
	eventHandle EventHandle
	isFirstRun  bool
	workQueue   *workqueue.WorkQueue
	//infos       Infos
}

func NewSClient(host, port string, eventHandle EventHandle) *SClient {
	return &SClient{
		host:        host,
		port:        port,
		eventHandle: eventHandle,
		isFirstRun:  true,
		workQueue:   workqueue.NewWorQueue(),
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

func (sc *SClient) CompareInfo(newInfos Infos) {
	oldInfos := sc.eventHandle.GetCurNodeInfos()

	//ret, _ := json.Marshal(oldInfos)
	//fmt.Printf("CompareInfo:%s\n", ret)
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
func (sc *SClient) handleMessage(id, msg string) {
	fmt.Printf("id=%s, msg:%s\n", id, msg)
	switch id {
	case INFOTYPE_NODEINFO:
		infos := make(map[string]*NodeInfos)
		//fmt.Printf("msg type INFOTYPE_NODEINFO %s\n", msg)
		json.Unmarshal([]byte(msg), &infos)
		if sc.isFirstRun {
			sc.isFirstRun = false
			sc.eventHandle.Init(infos)
		} else {
			sc.CompareInfo(infos)
		}
	case INFOTYPE_RESCHEDULE_OK:
		names := strings.Split(msg, ":")
		podNs, fromPodName, toPodName, fromNode, toNode := names[0], names[1], names[2], names[3], names[4]
		fmt.Printf("reschedule pod %s:%s from %s to %s %s Success %s\n", names[0], names[1], names[2], names[3], toPodName, names[4])
		sc.eventHandle.ReschedulePod(fromNode, toNode, podNs, fromPodName, toPodName)

		fmt.Printf("INFOTYPE_RESCHEDULE_OK stop %v\n", time.Now())
		/*ret := sc.workQueue.RemoveItemByID(id)
		if len(ret) > 0 {
			sc.workQueue.AsyncRun(ret[len(ret)-1])
		}*/
	case INFOTYPE_RESCHEDULE_FAIL:
		names := strings.Split(msg, ":")
		fmt.Printf("reschedule pod %s:%s from %s to %s fail %s\n", names[0], names[1], names[2], names[3], names[4])
	case INFOTYPE_MESSAGE:
		//fmt.Printf("msg type INFOTYPE_MESSAGE:%s\n", msgStr)
	}
}
func (sc *SClient) Run() {
	con, err := net.Dial("tcp", sc.host+":"+sc.port)
	if err != nil {
		return
	}
	defer con.Close()

	if err != nil {
		fmt.Println("Server not found.")
		return
	}
	fmt.Println("Connection OK.")

	//msg := make([]byte, 80960)
	for {
		fmt.Printf("before read \n")
		lastIndex := 0
		strBuf := ""
		for {
			bufBytes := make([]byte, 80960)
			length, err := con.Read(bufBytes)
			if err != nil {
				fmt.Printf("Error when read from server. err=%v\n", err)
				return
			}
			strBuf += string(bufBytes[:length])
			if bufBytes[length-1] == '#' {
				length = lastIndex + length
				break
			} else {
				lastIndex = lastIndex + length
			}
		}

		strs := strings.Split(strBuf, "#")
		for _, str := range strs {
			if str != "" {
				//sc.handleMessage(str)
				strs := strings.Split(str, "->")
				if len(strs) == 2 {
					sc.workQueue.AsyncRun(NewSClientWorkItem(strs[0], strs[1], sc))
				}
			}
		}
		con.Write([]byte("1"))
	}

}
