package atom3D

import (
	"fmt"
	"os"

	findfont "github.com/flopp/go-findfont"
	"github.com/fogleman/gg"
)

func RenderSO3(angle Vector) Tensor {
	var angle_rotaion Tensor
	if angle.X == 0 && angle.Y == 0 && angle.Z == 0 {
		angle_rotaion = Tensor{
			1, 0, 0,
			0, 1, 0,
			0, 0, 1}
	} else {
		angle_rotaion = SO3_x(angle.X).DotT(
			SO3_y(angle.Y).DotT(
				SO3_z(angle.Z)))
	}
	return angle_rotaion
}

type Render struct {
	Width       float64
	Height      float64
	Depth       float64
	Angle       Vector
	FocusFactor float64
}

func NewRender(width, height, depth float64) Render {
	return &Render{
		Width:       width,
		Height:      height,
		Depth:       depth,
		Angle:       Vector{0, 0, 0},
		FocusFactor: 1.}
}

func (render Render) Figure() *gg.Context {
	w, h := 10.*render.Width, 10.*render.Height
	return gg.NewContext(int(w), int(h))
}

func (render Render) Background(dc *gg.Context, color []float64) {
	dc.SetRGB(color[0], color[1], color[2])
	dc.DrawRectangle(0, 0, 10.*render.Width, 10.*render.Height)
	dc.Fill()
}

func (render Render) DrawAtom(dc *gg.Context, pos Vector, radius float64, color []float64) {
	render_pos := RenderSO3(render.Angle).DotV(pos)
	ratio := render.FocusFactor * render.Depth / (render_pos.Y + render.Depth)
	dc.SetRGB(color[0], color[1], color[2])
	dc.DrawCircle(5*render.Width+10.*render_pos.X, 5*render.Height-10.*render_pos.Z, 5*ratio*radius)
	dc.Fill()
}

func (render Render) DrawText(dc *gg.Context, pos Vector, text string, font_size float64, font string, color []float64) {
	render_pos := RenderSO3(render.Angle).DotV(pos)
	ratio := render.FocusFactor * render.Depth / (pos.Y + render.Depth)
	dc.SetRGB(color[0], color[1], color[2])
	fontPath, _ := findfont.Find(font)
	if err := dc.LoadFontFace(fontPath, ratio*font_size); err != nil {
		panic(err)
	}
	dc.DrawString(text, 5*render.Width+10.*render_pos.X, 5*render.Height-10.*render_pos.Z)
}

func (render Render) Save(dc *gg.Context, directory string, count int) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, os.ModeDir|0755)
	}
	dc.SavePNG(fmt.Sprintf("%s/render_%010d.png", directory, count))
}
