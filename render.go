package atom3D

import (
	"fmt"

	"github.com/fogleman/gg"
)

func Rendering(width *int, height *int, depth float64, filename string, directory string) {
	w, h := float64(*width)*2, float64(*height)*2
	dc := gg.NewContext(int(w), int(h))
	dc.SetRGB(1, 1, 1)
	dc.DrawRectangle(0, 0, w, h)
	dc.Fill()

	_, t, count, N, _, pos, _ := Read(filename)

	for i := 0; i < int(N); i++ {
		dc.DrawCircle(pos[i].X*2, h-pos[i].Y*2, 5)
		dc.SetRGB(0, 0, 255)
		dc.Fill()
	}
	/*
		dc.SetRGB(0, 0, 1)
		if err := dc.LoadFontFace("../polarity_image/D2CodingNerd.ttf", w/10); err != nil {
			panic(err)
		}
		dc.DrawString(fmt.Sprintf("time:%f", t), 0, h-10)
	*/

	dc.SavePNG(fmt.Sprintf("%s/render_%10d.png", directory, count))
	fmt.Println(int(t), "is done")
}
