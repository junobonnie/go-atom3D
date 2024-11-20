package atom3D

import (
	"fmt"
	"os"

	"github.com/fogleman/gg"
)

func Rendering(width float64, height float64, depth float64, filename string, directory string) {
	w, h := 10.*width, 10.*height
	dc := gg.NewContext(int(w), int(h))
	dc.SetRGB(1, 1, 1)
	dc.DrawRectangle(0, 0, w, h)
	dc.Fill()

	_, t, count, N, _, pos, _ := Read(filename)

	for i := 0; i < int(N); i++ {
		dc.DrawCircle(10.*pos[i].X, h-10*pos[i].Y, 5)
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
	fmt.Println(int(t), "is done")
}
