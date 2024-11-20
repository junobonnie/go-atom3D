package atom3D

import (
	"fmt"
	"os"

	"github.com/fogleman/gg"
)

func RenderSO3(angle Vector, omega Vector, t float64) Tensor {
	var omega_rotaion Tensor
	if omega.X == 0 && omega.Y == 0 && omega.Z == 0 {
		omega_rotaion = Tensor{
			1, 0, 0,
			0, 1, 0,
			0, 0, 1}
	} else {
		omega_rotaion = SO3_x(omega.X * t).DotT(
			SO3_y(omega.Y * t).DotT(
				SO3_z(omega.Z * t)))
	}
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
	return omega_rotaion.DotT(angle_rotaion)
}

func Rendering(width float64, height float64, depth float64, angle Vector, omega Vector, focus_factor float64, filename string, directory string) {
	w, h := 10.*width, 10.*height
	dc := gg.NewContext(int(w), int(h))
	dc.SetRGB(1, 1, 1)
	dc.DrawRectangle(0, 0, w, h)
	dc.Fill()

	_, t, count, N, _, pos, _ := Read(filename)

	for i := 0; i < int(N); i++ {
		render_pos := RenderSO3(angle, omega, t).DotV(pos[i])
		ratio := focus_factor * depth / (render_pos.Z + depth)
		dc.DrawCircle(w/2+10.*render_pos.X, h/2-10*render_pos.Y, 5*ratio)
		dc.SetRGB(0, 0, 1)
		dc.Fill()
	}
	/*
		dc.SetRGB(0, 0, 1)
		if err := dc.LoadFontFace("../polarity_image/D2CodingNerd.ttf", w/10); err != nil {
			panic(err)
		}
		dc.DrawString(fmt.Sprintf("time:%f", t), 0, h-10)
	*/
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, os.ModeDir|0755)
	}
	dc.SavePNG(fmt.Sprintf("%s/render_%010d.png", directory, count))
}
