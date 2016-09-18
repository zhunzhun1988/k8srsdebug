package animation

import (
	"image/color"
	"k8srsdraw/drawapi"
	"sync"
	"time"
)

type DrawerShape interface {
	Hide()
	IsHide() bool
	GetDrawer() *drawapi.Drawer
	SetStartPoint(point drawapi.DrawPoint)
	GetStop() chan int
	GetColor() color.Color
	Draw()
	DrawByColor(c color.Color)
}

type MoveListener interface {
	MoveStatue(s int)
}

func MoveTo(shape DrawerShape, from, to drawapi.DrawPoint, duration time.Duration, ml MoveListener) {
	go func() {
		dx := to.X - from.X
		dy := to.Y - from.Y
		steps := int(int64(duration.Nanoseconds()) / (100 * int64(time.Millisecond)))
	ExitFor:
		for i := 1; i <= steps; i++ {
			x := from.X + i*dx/steps
			y := from.Y + i*dy/steps
			if ((dx >= 0 && x >= to.X) || (dx <= 0 && x <= to.X)) &&
				((dy >= 0 && y >= to.Y) || (dy <= 0 && y <= to.Y)) {
				break
			}
			shape.DrawByColor(shape.GetDrawer().GetBackGround())
			shape.SetStartPoint(drawapi.DrawPoint{x, y})
			if shape.IsHide() {
				break ExitFor
			}
			if ml != nil {
				ml.MoveStatue(0)
			}
			shape.Draw()
			if ml != nil {
				ml.MoveStatue(1)
			}
			select {
			case <-time.After(100 * time.Millisecond):
			case <-shape.GetStop():
				break ExitFor
			}
		}
		if ml != nil {
			ml.MoveStatue(2)
		}
	}()
}
func StartFlicker(shape DrawerShape, t time.Duration) {
	go func() {
		isShow := true
		count := 0
	ExitFor:
		for {
			select {
			case <-time.After(300 * time.Millisecond):
				if shape.IsHide() {
					break ExitFor
				}
				if isShow {
					shape.DrawByColor(shape.GetDrawer().GetBackGround())
					isShow = false
				} else {
					shape.Draw()
					isShow = true
				}
				count++
				if time.Duration(count*300)*time.Millisecond >= t && isShow {
					break ExitFor
				}
			case <-shape.GetStop():
				shape.Draw()
				break ExitFor
			}
		}

	}()
}

type Animation interface {
	MoveTo(from, to drawapi.DrawPoint, duration time.Duration, ml MoveListener)
	StartFlicker(speed int, t time.Duration)
	StopAnimation()
}

type Rect struct {
	Drawer        *drawapi.Drawer
	StartPoint    drawapi.DrawPoint
	Width         int
	Height        int
	Color         color.Color
	IsFill        bool
	stopAnimation chan int
	mutux         sync.Mutex
	isHided       bool
}

func NewRect(d *drawapi.Drawer, startPoint drawapi.DrawPoint, w, h int, c color.Color, isFill bool) *Rect {
	return &Rect{
		Drawer:        d,
		StartPoint:    startPoint,
		Width:         w,
		Height:        h,
		Color:         c,
		IsFill:        isFill,
		mutux:         sync.Mutex{},
		stopAnimation: make(chan int),
		isHided:       true,
	}
}
func (r *Rect) GetColor() color.Color {
	return r.Color
}
func (r *Rect) Draw() {
	r.DrawByColor(r.Color)
	r.isHided = false
}
func (r *Rect) Hide() {
	r.DrawByColor(r.GetDrawer().GetBackGround())
	r.isHided = true
}
func (r *Rect) IsHide() bool {
	return r.isHided
}
func (r *Rect) GetStop() chan int {
	return r.stopAnimation
}
func (r *Rect) GetDrawer() *drawapi.Drawer {
	return r.Drawer
}
func (r *Rect) SetStartPoint(point drawapi.DrawPoint) {
	r.StartPoint = point
}
func (r *Rect) DrawByColor(c color.Color) {
	r.mutux.Lock()
	defer r.mutux.Unlock()
	if r.IsFill {
		r.Drawer.FillRect(r.StartPoint, r.Width, r.Height, c)
	} else {
		r.Drawer.DrawRect(r.StartPoint, r.Width, r.Height, c)
	}
}

func (r *Rect) StopAnimation() {
	r.stopAnimation <- 1
}
func (r *Rect) StartFlicker(t time.Duration) {
	StartFlicker(r, t)
}

func (r *Rect) MoveTo(from, to drawapi.DrawPoint, duration time.Duration, ml MoveListener) {
	MoveTo(r, from, to, duration, ml)
}

type Line struct {
	Drawer        *drawapi.Drawer
	StartPoint    drawapi.DrawPoint
	EndPoint      drawapi.DrawPoint
	Color         color.Color
	mutux         sync.Mutex
	stopAnimation chan int
	isHide        bool
}

func NewLine(d *drawapi.Drawer, startPoint, endPoint drawapi.DrawPoint, c color.Color) *Line {
	return &Line{
		Drawer:        d,
		StartPoint:    startPoint,
		EndPoint:      endPoint,
		Color:         c,
		mutux:         sync.Mutex{},
		stopAnimation: make(chan int),
		isHide:        true,
	}
}

func (l *Line) GetDrawer() *drawapi.Drawer {
	return l.Drawer
}
func (l *Line) SetStartPoint(point drawapi.DrawPoint) {
	l.EndPoint.X += point.X - l.StartPoint.X
	l.EndPoint.Y += point.Y - l.StartPoint.Y
	l.StartPoint = point

}
func (l *Line) GetStop() chan int {
	return l.stopAnimation
}
func (l *Line) GetColor() color.Color {
	return l.Color
}
func (l *Line) Draw() {
	l.DrawByColor(l.Color)
	l.isHide = false
}
func (l *Line) Hide() {
	l.DrawByColor(l.GetDrawer().GetBackGround())
	l.isHide = true
}
func (l *Line) IsHide() bool {
	return l.isHide
}
func (l *Line) DrawByColor(c color.Color) {
	l.mutux.Lock()
	defer l.mutux.Unlock()
	l.Drawer.DrawLine(l.StartPoint, l.EndPoint, c)
}

func (l *Line) StopAnimation() {
	l.stopAnimation <- 1
}

func (l *Line) StartFlicker(t time.Duration) {
	StartFlicker(l, t)
}

func (l *Line) MoveTo(from, to drawapi.DrawPoint, duration time.Duration, ml MoveListener) {
	MoveTo(l, from, to, duration, ml)
}

type TextWidgt struct {
	Drawer        *drawapi.Drawer
	StartPoint    drawapi.DrawPoint
	Width         int
	Height        int
	Text          string
	Color         color.Color
	mutux         sync.Mutex
	FontSize      float64
	stopAnimation chan int
	isHide        bool
}

func NewTextWidgt(d *drawapi.Drawer, startPoint drawapi.DrawPoint, width, height int, text string,
	fontSize float64, c color.Color) *TextWidgt {
	if int(fontSize)+10 > height {
		height = int(fontSize) + 10
	}
	return &TextWidgt{
		Drawer:        d,
		StartPoint:    startPoint,
		Text:          text,
		Color:         c,
		mutux:         sync.Mutex{},
		FontSize:      fontSize,
		stopAnimation: make(chan int),
		Width:         width,
		Height:        height,
		isHide:        true,
	}
}
func (t *TextWidgt) GetDrawer() *drawapi.Drawer {
	return t.Drawer
}
func (t *TextWidgt) SetStartPoint(point drawapi.DrawPoint) {
	t.StartPoint = point
}
func (t *TextWidgt) GetStop() chan int {
	return t.stopAnimation
}
func (t *TextWidgt) GetColor() color.Color {
	return t.Color
}
func (t *TextWidgt) Draw() {
	t.DrawByColor(t.Color)
	t.isHide = false
}
func (t *TextWidgt) Hide() {
	//t.DrawByColor(t.GetDrawer().GetBackGround())
	t.GetDrawer().FillRect(t.StartPoint, t.Width, t.Height, t.GetDrawer().GetBackGround())
	t.isHide = true
}
func (t *TextWidgt) IsHide() bool {
	return t.isHide
}
func (t *TextWidgt) DrawByColor(c color.Color) {
	t.mutux.Lock()
	defer t.mutux.Unlock()
	t.Drawer.DrawText(t.StartPoint,
		t.Drawer.GetStrByWidth(t.Text, t.FontSize, t.Width), t.FontSize, t.Color)
}

func (t *TextWidgt) StopAnimation() {
	t.stopAnimation <- 1
}

func (t *TextWidgt) StartFlicker(duration time.Duration) {
	StartFlicker(t, duration)
}

func (t *TextWidgt) MoveTo(from, to drawapi.DrawPoint, duration time.Duration, ml MoveListener) {
	MoveTo(t, from, to, duration, ml)
}

type Circle struct {
	Drawer        *drawapi.Drawer
	StartPoint    drawapi.DrawPoint
	Radius        int
	IsFill        bool
	Color         color.Color
	mutux         sync.Mutex
	stopAnimation chan int
	isHide        bool
}

func NewCircle(d *drawapi.Drawer, startPoint drawapi.DrawPoint, radius int, isFill bool,
	c color.Color) *Circle {
	return &Circle{
		Drawer:        d,
		StartPoint:    startPoint,
		Radius:        radius,
		IsFill:        isFill,
		Color:         c,
		mutux:         sync.Mutex{},
		stopAnimation: make(chan int),
		isHide:        true,
	}
}
func (c *Circle) GetDrawer() *drawapi.Drawer {
	return c.Drawer
}
func (c *Circle) SetStartPoint(point drawapi.DrawPoint) {
	c.StartPoint = point
}
func (c *Circle) GetStop() chan int {
	return c.stopAnimation
}
func (c *Circle) GetColor() color.Color {
	return c.Color
}
func (c *Circle) Draw() {
	c.DrawByColor(c.Color)
	c.isHide = false
}
func (c *Circle) Hide() {
	c.DrawByColor(c.GetDrawer().GetBackGround())
	c.isHide = true
}
func (c *Circle) IsHide() bool {
	return c.isHide
}
func (c *Circle) DrawByColor(color color.Color) {
	c.mutux.Lock()
	defer c.mutux.Unlock()
	c.Drawer.DrawCircle(c.StartPoint, c.Radius, c.IsFill, color)
}

func (c *Circle) StopAnimation() {
	c.stopAnimation <- 1
}

func (c *Circle) StartFlicker(duration time.Duration) {
	StartFlicker(c, duration)
}

func (c *Circle) MoveTo(from, to drawapi.DrawPoint, duration time.Duration, ml MoveListener) {
	MoveTo(c, from, to, duration, ml)
}
