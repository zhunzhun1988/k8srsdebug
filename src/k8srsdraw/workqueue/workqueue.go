package workqueue

import (
	"sync"
)

type WorkItem interface {
	Run()
	GetID() string
}

const (
	WORKQUEUE_IDLE = iota
	WORKQUEUE_NOSTARTED
	WORKQUEUE_STOPED
	WORKQUEUE_RUNNING
)

type WorkQueuRunStatue int
type WorkQueue struct {
	mutex         sync.Mutex
	wait          chan int
	workItemlist  []WorkItem
	runStatuMutex sync.Mutex
	runStatues    WorkQueuRunStatue
}

func NewWorQueue() *WorkQueue {
	ret := &WorkQueue{
		mutex:         sync.Mutex{},
		runStatuMutex: sync.Mutex{},
		wait:          make(chan int),
		workItemlist:  make([]WorkItem, 0),
		runStatues:    WORKQUEUE_NOSTARTED,
	}
	ret.Start()
	return ret
}
func (wq *WorkQueue) IsRunning() bool {
	wq.runStatuMutex.Lock()
	defer wq.runStatuMutex.Unlock()
	return wq.runStatues == WORKQUEUE_RUNNING
}
func (wq *WorkQueue) setRunStatues(s WorkQueuRunStatue) {
	wq.runStatuMutex.Lock()
	defer wq.runStatuMutex.Unlock()
	wq.runStatues = s
}
func (wq *WorkQueue) getRunStatues() WorkQueuRunStatue {
	wq.runStatuMutex.Lock()
	defer wq.runStatuMutex.Unlock()
	return wq.runStatues
}
func (wq *WorkQueue) AddWorkItem(item WorkItem) {
	wq.mutex.Lock()
	defer wq.mutex.Unlock()
	wq.workItemlist = append(wq.workItemlist, item)
}
func (wq *WorkQueue) PopWorkItem() WorkItem {
	wq.mutex.Lock()
	defer wq.mutex.Unlock()
	if len(wq.workItemlist) == 0 {
		return nil
	} else {
		ret := wq.workItemlist[0]
		wq.workItemlist = wq.workItemlist[1:]
		return ret
	}
}
func (wq *WorkQueue) RemoveAllItem() {
	wq.mutex.Lock()
	defer wq.mutex.Unlock()
	wq.workItemlist = make([]WorkItem, 0)
}
func (wq *WorkQueue) RemoveItemByID(id string) []WorkItem {
	wq.mutex.Lock()
	defer wq.mutex.Unlock()
	ret := make([]WorkItem, 0)
	tmp := make([]WorkItem, 0)
	for _, item := range wq.workItemlist {
		if item.GetID() == id {
			ret = append(ret, item)
		} else {
			tmp = append(tmp, item)
		}
	}
	wq.workItemlist = tmp
	return ret
}
func (wq *WorkQueue) AsyncRun(workItem WorkItem) {
	wq.AddWorkItem(workItem)
	if wq.getRunStatues() == WORKQUEUE_IDLE {
		wq.wait <- 1
	}
}
func (wq *WorkQueue) Start() {
	go func() {
		for {
			wq.setRunStatues(WORKQUEUE_IDLE)
			select {
			case <-wq.wait:
				wq.setRunStatues(WORKQUEUE_RUNNING)
				for {
					workItem := wq.PopWorkItem()
					if workItem != nil {
						workItem.Run()
					} else {
						wq.setRunStatues(WORKQUEUE_IDLE)
						break
					}
				}
			}
		}
	}()
}
