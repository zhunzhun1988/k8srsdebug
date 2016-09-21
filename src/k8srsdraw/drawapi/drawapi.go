// Copyright 2016-09-14 patrickxie

package drawapi

import (
	"image"
	"image/color"
	"io/ioutil"
	"time"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type DrawPoint struct {
	X, Y int
}

type Drawer struct {
	rgba       *image.RGBA
	font       *truetype.Font
	background color.Color
	xu         *xgbutil.XUtil
	xwin       *xwindow.Window
	stop       chan int
	changed    bool
}

func Abs(x int) int {
	if x >= 0 {
		return x
	}
	return -x
}
func Max(x int, y int) int {
	if x >= y {
		return x
	}
	return y
}

func NewDrawer(xu *xgbutil.XUtil, xwin *xwindow.Window, rgba *image.RGBA, bg color.Color) *Drawer {
	fontBytes, _ := ioutil.ReadFile("./luxisr.ttf")
	f, _ := truetype.Parse(fontBytes)
	return &Drawer{
		rgba:       rgba,
		font:       f,
		background: bg,
		xu:         xu,
		xwin:       xwin,
		stop:       make(chan int),
		changed:    true,
	}
}
func (d *Drawer) Show() {
	if d.changed == true {
		d.changed = false
		ximg := xgraphics.NewConvert(d.xu, d.rgba)
		// I want 'ximg' to show on 'xwin'
		ximg.XSurfaceSet(d.xwin.Id)
		// now show it
		ximg.XDraw()
		ximg.XPaint(d.xwin.Id)
	}
}
func (d *Drawer) Run() {
	go func() {
	ExitFor:
		for {
			select {
			case <-d.stop:
				break ExitFor
			case <-time.After(100 * time.Millisecond):
				d.Show()
			}
		}
	}()
}
func (d *Drawer) StopRun() {
	d.stop <- 1
}
func (d *Drawer) GetBackGround() color.Color {
	return d.background
}
func (d *Drawer) DrawLine(startPoint DrawPoint, endPoint DrawPoint, c color.Color) {
	dx := Abs(startPoint.X - endPoint.X)
	dy := Abs(startPoint.Y - endPoint.Y)
	maxD := Max(dx, dy)
	for i := 0; i <= maxD; i++ {
		x := int(float64(startPoint.X) + (float64(endPoint.X-startPoint.X))*(float64(i)/float64(maxD)))
		y := int(float64(startPoint.Y) + (float64(endPoint.Y-startPoint.Y))*(float64(i)/float64(maxD)))
		d.rgba.Set(x, y, c)
	}
	d.changed = true
}
func (d *Drawer) DrawLineWithAnimation(startPoint DrawPoint, endPoint DrawPoint, c color.Color, duration time.Duration) {
	steps := int(int64(duration.Nanoseconds()) / (100 * int64(time.Millisecond)))
	dx := Abs(startPoint.X - endPoint.X)
	dy := Abs(startPoint.Y - endPoint.Y)
	maxD := Max(dx, dy)
	dd := (maxD / steps) + 1
	for i := 0; i <= maxD; i++ {
		if i > 0 && i%dd == 0 {
			time.Sleep(100 * time.Millisecond)
		}
		x := int(float64(startPoint.X) + (float64(endPoint.X-startPoint.X))*(float64(i)/float64(maxD)))
		y := int(float64(startPoint.Y) + (float64(endPoint.Y-startPoint.Y))*(float64(i)/float64(maxD)))
		d.rgba.Set(x, y, c)
		d.changed = true
	}
}
func (d *Drawer) DrawCircle(startPoint DrawPoint, radius int, isFill bool, c color.Color) {
	if isFill {
		for x := startPoint.X - radius; x <= startPoint.X+radius; x++ {
			for y := startPoint.Y - radius; y <= startPoint.Y+radius; y++ {
				tmp := (startPoint.X-x)*(startPoint.X-x) + (startPoint.Y-y)*(startPoint.Y-y)
				if tmp <= radius*radius {
					d.rgba.Set(x, y, c)
				}
			}
		}
	} else {
		for x := startPoint.X - radius; x <= startPoint.X+radius; x++ {
			for y := startPoint.Y - radius; y <= startPoint.Y+radius; y++ {
				tmp := (startPoint.X-x)*(startPoint.X-x) + (startPoint.Y-y)*(startPoint.Y-y)
				if radius*radius <= tmp && tmp <= (radius+1)*(radius+1) {
					d.rgba.Set(x, y, c)
				}
			}
		}
	}
	d.changed = true
}
func (d *Drawer) DrawRect(startPoint DrawPoint, width, height int, c color.Color) {
	d.DrawLine(startPoint, DrawPoint{startPoint.X + width, startPoint.Y}, c)
	d.DrawLine(startPoint, DrawPoint{startPoint.X, startPoint.Y + height}, c)
	d.DrawLine(DrawPoint{startPoint.X + width, startPoint.Y + height},
		DrawPoint{startPoint.X + width, startPoint.Y}, c)
	d.DrawLine(DrawPoint{startPoint.X + width, startPoint.Y + height},
		DrawPoint{startPoint.X, startPoint.Y + height}, c)
}
func (d *Drawer) FillRect(startPoint DrawPoint, width, height int, c color.Color) {
	for x := 0; x <= width; x++ {
		for y := 0; y <= height; y++ {
			d.rgba.Set(startPoint.X+x, startPoint.Y+y, c)
		}
	}
	d.changed = true
}
func (d *Drawer) GetStrByWidth(text string, fontSize float64, width int) string {
	f := truetype.NewFace(d.font, &truetype.Options{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	l := font.MeasureString(f, text).Floor()
	if l <= width {
		return text
	}
	bytes := []byte(text)
	i := int(float64(len(bytes)) / float64(l) * float64(width))
	for {
		t := font.MeasureString(f, string(bytes[:i])+"...").Floor()
		if t <= width {
			if i+1 >= len(bytes) {
				return string(bytes[:i]) + "..."
			} else if font.MeasureString(f, string(bytes[:i+1])+"...").Floor() >= width {
				return string(bytes[:i]) + "..."
			}
		}
		if t > width {
			i -= 1
		} else {
			i += 1
		}
	}
}

func (d *Drawer) DrawText(startPoint DrawPoint, text string,
	fontSize float64, c color.Color) int {
	drawer := &font.Drawer{
		Dst: d.rgba,
		Src: image.NewUniform(c),
		Face: truetype.NewFace(d.font, &truetype.Options{
			Size:    fontSize,
			DPI:     72,
			Hinting: font.HintingNone,
		}),
	}
	drawer.Dot = fixed.Point26_6{
		X: fixed.I(startPoint.X),
		Y: fixed.I(startPoint.Y + int(fontSize)),
	}
	drawer.DrawString(text)
	d.changed = true
	return drawer.MeasureString(text).Floor()
}
