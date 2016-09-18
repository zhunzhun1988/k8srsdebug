package window

import (
	"fmt"
	"image"
	"image/color"
	"k8srsdraw/animation"
	"k8srsdraw/drawapi"
	"sort"
	"sync"
	"time"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xwindow"
)

var (
	PodColor        = color.RGBA{0x00, 0xff, 0x00, 0xff}
	NodeColor       = color.RGBA{0xff, 0x00, 0x00, 0xff}
	NodeOKColor     = color.RGBA{0x00, 0xff, 0x00, 0xff}
	LineColor       = color.RGBA{0xff, 0xff, 0xff, 0xff}
	NodeRowSpace    = 10
	NodeColumSpace  = 10
	NodeTopPadding  = 10
	NodeLeftPadding = 10
	PodPadding      = 10
	PodHeight       = 22
	TextLeftPadding = 5
)

func getRowColum(num int) (r, c int) {
	if num <= 4 {
		return 1, num
	} else if num <= 8 {
		return 2, 4
	} else if num <= 12 {
		return 3, 4
	} else if num <= 16 {
		return 4, 4
	} else if num <= 20 {
		return 4, 5
	} else if num <= 25 {
		return 5, 5
	} else if num <= 30 {
		return 5, 6
	} else {
		if num%6 == 0 {
			return 6, num / 6
		} else {
			return 6, num/6 + 1
		}
	}
}
func getNextPos(rNum, cNum int, curR, curC int) (r, c int) {
	if curC+1 < cNum {
		return curR, curC + 1
	} else if curR+1 < rNum {
		return curR + 1, 0
	} else {
		return rNum - 1, cNum - 1
	}
}

type PodShowStatue struct {
	ShowString     string
	ShowStartPoint drawapi.DrawPoint
	ShowWidth      int
	ShowHeight     int
}
type Pod struct {
	StartPoint drawapi.DrawPoint
	Width      int
	Height     int
	Namespace  string
	Names      map[string]bool
	Count      int
	rect       *animation.Rect
	text       *animation.TextWidgt
	showStatus PodShowStatue
}
type PodList []*Pod

func (pl PodList) Len() int {
	return len(pl)
}
func (pl PodList) Less(i, j int) bool {

	if pl[i].Count != pl[j].Count {
		return pl[i].Count > pl[j].Count
	} else {
		return pl[i].Namespace < pl[j].Namespace
	}
}
func (pl PodList) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}
func NewPod(startPoint drawapi.DrawPoint, w, h int, ns string, c int) *Pod {
	return &Pod{
		StartPoint: startPoint,
		Width:      w,
		Height:     h,
		Namespace:  ns,
		Names:      make(map[string]bool),
		Count:      c,
		rect:       nil,
		text:       nil,
		showStatus: PodShowStatue{},
	}
}
func (p *Pod) MoveTo(d *drawapi.Drawer, toPoint drawapi.DrawPoint, ml animation.MoveListener) {
	if p.Count == 1 {
		p.text.MoveTo(p.rect.StartPoint, toPoint, 5*time.Second, ml)
		p.rect.MoveTo(p.rect.StartPoint, toPoint, 5*time.Second, ml)
	} else {
		tmpPod := NewPod(p.StartPoint, p.Width, p.Height, p.Namespace, 1)
		tmpPod.Show(d)
		tmpPod.MoveTo(d, toPoint, ml)
	}

}
func (p *Pod) IsShowStatusChanged() bool {
	showStr := p.GetShowStr()
	if p.showStatus.ShowHeight != p.Height ||
		p.showStatus.ShowWidth != p.Width ||
		p.showStatus.ShowStartPoint != p.StartPoint ||
		p.showStatus.ShowString != showStr {
		return true
	}
	return false
}
func (p *Pod) GetShowStr() string {
	return fmt.Sprintf("%s:%d", p.Namespace, p.Count)
}
func (p *Pod) Show(d *drawapi.Drawer) {
	showStr := p.GetShowStr()
	p.rect = animation.NewRect(d, p.StartPoint, p.Width, p.Height, PodColor, false)
	p.text = animation.NewTextWidgt(d, drawapi.DrawPoint{p.StartPoint.X + TextLeftPadding,
		p.StartPoint.Y}, p.Width-TextLeftPadding, p.Height, showStr, 15, PodColor)
	p.text.Draw()
	p.rect.Draw()
	p.showStatus.ShowString = showStr
	p.showStatus.ShowStartPoint = p.StartPoint
	p.showStatus.ShowHeight = p.Height
	p.showStatus.ShowWidth = p.Width
}
func (p *Pod) Hide() {
	if p.text != nil {
		p.text.Hide()
	}
	if p.rect != nil {
		p.rect.Hide()
	}

}
func (p *Pod) Flicker(duration time.Duration) {
	if p.rect != nil {
		p.rect.StartFlicker(duration)
	}
}

type Node struct {
	Name       string
	StartPoint drawapi.DrawPoint
	Width      int
	Height     int
	Pods       map[string]*Pod
}
type NodeList []*Node

func (nl NodeList) Less(i, j int) bool {
	return nl[i].Name < nl[j].Name
}
func (nl NodeList) Len() int {
	return len(nl)
}
func (nl NodeList) Swap(i, j int) {
	nl[i], nl[j] = nl[j], nl[i]
}
func NewNode(name string, startPoint drawapi.DrawPoint, w, h int) *Node {
	return &Node{
		Name:       name,
		StartPoint: startPoint,
		Width:      w,
		Height:     h,
		Pods:       make(map[string]*Pod),
	}
}
func (n *Node) GetPodList() PodList {
	pl := make([]*Pod, 0)
	for _, p := range n.Pods {
		pl = append(pl, p)
	}
	return pl
}
func (n *Node) GetPodPos(i int) (startPoint drawapi.DrawPoint, w, h int) {
	startPoint.X = n.StartPoint.X + PodPadding
	startPoint.Y = n.StartPoint.Y + n.Height - (i+1)*(PodHeight+PodPadding)
	w = n.Width - PodPadding*2
	h = PodHeight
	return
}
func (n *Node) GetPod(podNamespace string) *Pod {
	p, _ := n.Pods[podNamespace]
	return p
}
func (n *Node) AddPod(d *drawapi.Drawer, podNamespace, podName string) {
	p, find := n.Pods[podNamespace]
	fmt.Printf("patrick debug %v\n", find)
	if find {
		if _, ok := p.Names[podName]; ok {
			fmt.Println("pod:%s:%s is added return\n", podNamespace, podName)
			return
		}
		p.Names[podName] = true
		p.Count++
		n.Draw(d)
		p.Flicker(1 * time.Second)
		fmt.Printf("patrick debug %d\n", p.Count)
	} else {
		sp, w, h := n.GetPodPos(len(n.Pods))
		p = NewPod(sp, w, h, podNamespace, 1)
		p.Names[podName] = true
		n.Pods[podNamespace] = p
		p.Show(d)
		n.Draw(d)
	}
}
func (n *Node) DeletePod(d *drawapi.Drawer, podNamespace, podName string) {
	p, find := n.Pods[podNamespace]
	if find {
		if _, ok := p.Names[podName]; !ok {
			fmt.Println("pod:%s:%s is not exist return\n", podNamespace, podName)
			return
		}
		p.Count--
		delete(p.Names, podName)
		if p.Count <= 0 {
			p.Hide()
			delete(n.Pods, podNamespace)
			n.Draw(d)
		} else {
			n.Draw(d)
			p.Flicker(1 * time.Second)
		}
	}
}

func (n *Node) Draw(d *drawapi.Drawer) {
	nameHeight := 21
	c := NodeColor
	if len(n.Pods) <= 1 {
		c = NodeOKColor
	}
	r1 := animation.NewRect(d, n.StartPoint, n.Width-20, nameHeight, c, false)
	r2 := animation.NewRect(d, drawapi.DrawPoint{n.StartPoint.X,
		n.StartPoint.Y + nameHeight}, n.Width, n.Height-nameHeight, c, false)
	r1.Draw()
	r2.Draw()
	t := animation.NewTextWidgt(d, drawapi.DrawPoint{n.StartPoint.X + TextLeftPadding,
		n.StartPoint.Y + 1}, n.Width-TextLeftPadding-21, nameHeight, n.Name, float64(nameHeight-4), c)
	t.Hide()
	t.Draw()
	l := animation.NewLine(d, drawapi.DrawPoint{n.StartPoint.X, n.StartPoint.Y + 20},
		drawapi.DrawPoint{n.StartPoint.X + n.Width, n.StartPoint.Y + 20}, c)
	l.Draw()
	for _, p := range n.Pods {
		//if p.IsShowStatusChanged() {
		p.Hide()
		//}
	}

	i := 0
	pl := n.GetPodList()
	sort.Sort(pl)

	for _, p := range pl {
		sp, w, h := n.GetPodPos(i)
		p.StartPoint, p.Width, p.Height = sp, w, h
		p.Show(d)
		i++
	}
}

type Window struct {
	width      int
	height     int
	background color.Color
	drawer     *drawapi.Drawer
	xu         *xgbutil.XUtil
	Nodes      map[string]*Node
	canvas     *image.RGBA
	xwin       *xwindow.Window
	mutex      sync.Mutex
}

func NewWindow(w, h int, bg color.Color) *Window {
	xu, err := xgbutil.NewConn()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// just create a id for the window
	xwin, err := xwindow.Generate(xu)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	// now, create the window
	err = xwin.CreateChecked(
		xu.RootWin(), // parent window
		0, 0, w, h,   // window size
		0) // related to event, not considered here
	if err != nil {
		fmt.Println(err)
		return nil
	}
	// now we can see the window on the screen
	xwin.Map()

	// 'rstr' calculates the data needed to draw
	// 'painter' draw with the data on 'canvas'
	canvas := image.NewRGBA(image.Rect(0, 0, w, h))

	d := drawapi.NewDrawer(xu, xwin, canvas, color.RGBA{0x00, 0x00, 0x00, 0xff})
	d.Run()
	return &Window{
		width:      w,
		height:     h,
		background: bg,
		drawer:     d,
		xu:         xu,
		Nodes:      make(map[string]*Node),
		canvas:     canvas,
		xwin:       xwin,
		mutex:      sync.Mutex{},
	}
}
func (w *Window) GetDrawer() *drawapi.Drawer {
	return w.drawer
}
func (w *Window) WaitEvent() {
	xevent.Main(w.xu)
}

func (w *Window) AddNode(name string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	_, find := w.Nodes[name]
	if find == false {
		w.Nodes[name] = NewNode(name, drawapi.DrawPoint{0, 0}, 0, 0)
		w.Update(true)
	}
}
func (w *Window) GetNodeList() NodeList {
	nl := make([]*Node, 0)
	for _, n := range w.Nodes {
		nl = append(nl, n)
	}
	return nl
}
func (w *Window) AddPod(nodeName, podNamespace, podName string) {
	n, find := w.Nodes[nodeName]
	if find == true {
		n.AddPod(w.drawer, podNamespace, podName)
	}
}
func (w *Window) DeletePod(nodeName, podNamespace, podName string) {
	n, find := w.Nodes[nodeName]
	if find == true {
		n.DeletePod(w.drawer, podNamespace, podName)
	}
}
func (w *Window) ExistNode(name string) bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	_, find := w.Nodes[name]
	return find
}
func (w *Window) DeleteNode(name string) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	_, find := w.Nodes[name]
	if find == true {
		delete(w.Nodes, name)
		w.Update(true)
	}
}
func (w *Window) MovePodFromTo(fromNode, toNode, podNamespace, podName string) {
	go func() {
		w.mutex.Lock()
		defer w.mutex.Unlock()
		nodeFrom, findFrom := w.Nodes[fromNode]
		if findFrom == false {
			return
		}
		nodeTo, findTo := w.Nodes[toNode]
		if findTo == false {
			return
		}
		pf := nodeFrom.GetPod(podNamespace)
		if pf == nil {
			return
		}
		pt := nodeTo.GetPod(podNamespace)
		var ptPoint drawapi.DrawPoint

		if pt != nil {
			ptPoint = pt.StartPoint
		} else {
			ptPoint, _, _ = nodeTo.GetPodPos(len(nodeTo.Pods))
		}

		startPoint := drawapi.DrawPoint{pf.StartPoint.X + pf.Width, pf.StartPoint.Y + pf.Height}
		endPoint := ptPoint

		w.GetDrawer().DrawLineWithAnimation(startPoint, endPoint, LineColor, 2*time.Second)
		time.Sleep(300 * time.Millisecond)
		w.GetDrawer().DrawLine(startPoint, endPoint, w.GetDrawer().GetBackGround())
		nodeFrom.DeletePod(w.drawer, podNamespace, podName)
		nodeTo.AddPod(w.drawer, podNamespace, podName)

		//time.Sleep(3 * time.Second)
		//w.Update(true)
	}()
}
func (w *Window) MoveStatue(s int) {
	if s == 1 {
		//	w.Update(true)
	} else if s == 0 {
		//	w.Update(false)
	}
}
func (w *Window) Update(force bool) {
	if force {
		w.drawer.StopRun()
		w.canvas = image.NewRGBA(image.Rect(0, 0, w.width, w.height))
		w.drawer = drawapi.NewDrawer(w.xu, w.xwin, w.canvas, color.RGBA{0x00, 0x00, 0x00, 0xff})
		w.drawer.Run()
	}

	if len(w.Nodes) == 0 {
		return
	}
	rNum, cNum := getRowColum(len(w.Nodes))
	nodeWidth := (w.width - NodeLeftPadding*2 - (cNum-1)*NodeColumSpace) / cNum
	nodeHeight := (w.height - NodeTopPadding*2 - (rNum-1)*NodeRowSpace) / rNum
	r, c := 0, 0
	nl := w.GetNodeList()
	sort.Sort(nl)
	for _, node := range nl {
		node.StartPoint = drawapi.DrawPoint{NodeLeftPadding + c*(NodeColumSpace+nodeWidth),
			NodeTopPadding + r*(NodeRowSpace+nodeHeight)}
		node.Width = nodeWidth
		node.Height = nodeHeight
		r, c = getNextPos(rNum, cNum, r, c)
	}
	for _, node := range nl {
		node.Draw(w.drawer)
	}
}
