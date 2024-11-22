package atom3D

import (
	"fmt"
	"os"
	"sort"

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

func NewRender(width, height, depth float64) *Render {
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

func (render Render) Background(dc *gg.Context, rgba []float64) {
	dc.SetRGBA(rgba[0], rgba[1], rgba[2], rgba[3])
	dc.DrawRectangle(0, 0, 10.*render.Width, 10.*render.Height)
	dc.Fill()
}

func (render Render) DrawAtom(dc *gg.Context, pos Vector, radius float64, rgba []float64) {
	render_pos := RenderSO3(render.Angle).DotV(pos)
	ratio := render.FocusFactor * render.Depth / (render_pos.Y + render.Depth)
	dc.SetRGBA(rgba[0], rgba[1], rgba[2], rgba[3])
	dc.DrawCircle(5*render.Width+10.*render_pos.X, 5*render.Height-10.*render_pos.Z, 5*ratio*radius)
	dc.Fill()
}

func (render Render) DrawText(dc *gg.Context, pos Vector, text string, font_size float64, font string, rgba []float64) {
	render_pos := RenderSO3(render.Angle).DotV(pos)
	ratio := render.FocusFactor * render.Depth / (pos.Y + render.Depth)
	dc.SetRGBA(rgba[0], rgba[1], rgba[2], rgba[3])
	fontPath, _ := findfont.Find(font)
	if err := dc.LoadFontFace(fontPath, 10*ratio*font_size); err != nil {
		panic(err)
	}
	dc.DrawString(text, 5*render.Width+10.*render_pos.X, 5*render.Height-10.*render_pos.Z)
}

func (render Render) DrawPlaneText(dc *gg.Context, x float64, y float64, text string, font_size float64, font string, rgba []float64) {
	dc.SetRGBA(rgba[0], rgba[1], rgba[2], rgba[3])
	fontPath, _ := findfont.Find(font)
	if err := dc.LoadFontFace(fontPath, 10*font_size); err != nil {
		panic(err)
	}
	dc.DrawString(text, 5*render.Width+10.*x, 5*render.Height-10.*y)
}

func (render Render) DrawLine(dc *gg.Context, pos1 Vector, pos2 Vector, width float64, rgba []float64) {
	render_pos1 := RenderSO3(render.Angle).DotV(pos1)
	render_pos2 := RenderSO3(render.Angle).DotV(pos2)
	dc.SetRGBA(rgba[0], rgba[1], rgba[2], rgba[3])
	dc.DrawLine(5*render.Width+10.*render_pos1.X, 5*render.Height-10.*render_pos1.Z, 5*render.Width+10.*render_pos2.X, 5*render.Height-10.*render_pos2.Z)
	dc.SetLineWidth(width)
	dc.Stroke()
}

func (render Render) DrawAxis(dc *gg.Context, length float64, width float64, font_size float64, font string) {
	e := []Vector{Vector{length, 0, 0}, Vector{0, length, 0}, Vector{0, 0, length}}
	rgb := [][]float64{{1, 0, 0, 1}, {0, 1, 0, 1}, {0, 0, 1, 1}}
	text := []string{"X", "Y", "Z"}

	indices := render.GetSortedIndices(e)

	for _, i := range indices {
		render.DrawLine(dc, Vector{0, 0, 0}, e[i], width, rgb[i])
		render.DrawText(dc, e[i], text[i], font_size, font, rgb[i])
	}
}

func (render Render) GetSortedIndices(pos []Vector) []int {
	indices := make([]int, len(pos))
    for i := range indices {
        indices[i] = i
    }
    
	veiw_pos := RenderSO3(render.Angle).DotV(Vector{0, 1, 0})

    // 인덱스를 데이터 값에 따라 정렬
    sort.Slice(indices, func(i, j int) bool {
        return veiw_pos.Dot(pos[indices[i]]) > veiw_pos.Dot(pos[indices[j]])
    })

	return indices
}

func (render Render) Save(dc *gg.Context, directory string, count int) {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, os.ModeDir|0755)
	}
	dc.SavePNG(fmt.Sprintf("%s/render_%010d.png", directory, count))
}
