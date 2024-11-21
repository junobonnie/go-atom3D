package atom3D

import (
	"fmt"
	"os"

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

func (render Render) Figure() *gg.Context {
	w, h := 10.*render.Width, 10.*render.Height
	return gg.NewContext(int(w), int(h))
}

func (render Render) Background(dc *gg.Context) {
	dc.SetRGB(1, 1, 1)
	dc.DrawRectangle(0, 0, 10.*render.Width, 10.*render.Height)
	dc.Fill()
}

func (render Render) DrawAtom(dc *gg.Context, pos Vector, radius float64) {
	render_pos := RenderSO3(render.Angle).DotV(pos)
	ratio := render.FocusFactor * render.Depth / (render_pos.Z + render.Depth)
	dc.SetRGB(0, 0, 1)
	dc.DrawCircle(5*render.Width+10.*render_pos.X, 5*render.Height-10.*render_pos.Y, 5*ratio*radius)
	dc.Fill()
}

func (render Render) DrawText(dc *gg.Context, pos Vector, font_size float64, text string) {
	render_pos := RenderSO3(render.Angle).DotV(pos)
	ratio := render.FocusFactor * render.Depth / (pos.Z + render.Depth)
	dc.SetRGB(1, 0, 0)
	if err := dc.LoadFontFace("D2CodingNerd.ttf", ratio*font_size); err != nil {
		panic(err)
	}
	dc.DrawString(text, 5*render.Width+10.*render_pos.X, 5*render.Height-10.*render_pos.Y)
}

func (render Render) Save(dc *gg.Context, filename string, directory string, count int) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, os.ModeDir|0755)
	}
	dc.SavePNG(fmt.Sprintf("%s/%s_%010d.png", directory, filename, count))
}
