package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"encoding/base64"

	"github.com/golang/freetype"
	"golang.org/x/image/font"
)

var (
	dpi      = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile = flag.String("fontfile", "./Martel-Bold.ttf", "filename of the ttf font")
	hinting  = flag.String("hinting", "none", "none | full")
	size     = flag.Float64("size", 12, "font size in points")
	spacing  = flag.Float64("spacing", 1.5, "line spacing (e. g. 2 means double spaced)")
	wonb     = flag.Bool("whiteonblack", false, "white ext on a black background")
)

var text = []string{
	"HORDALAND!",
}

func handler(w http.ResponseWriter, r *http.Request) {
	flag.Parse()

	// Read the font data.
	fontBytes, err := ioutil.ReadFile(*fontfile)

	if err != nil {
		log.Println(err)
		return
	}

	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}

	// Initialize the context
	fg, bg := image.Black, image.White
	ruler := color.RGBA{0xdd, 0xdd, 0xdd, 0xff}

	if *wonb {
		fg, bg = image.White, image.Black
		ruler = color.RGBA{0x22, 0x22, 0x22, 0xff}
	}
	rgba := image.NewRGBA(image.Rect(0, 0, 400, 300))
	draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
	c := freetype.NewContext()
	c.SetDPI(*dpi)
	c.SetFont(f)
	c.SetFontSize(*size)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(fg)
	switch *hinting {
	default:
		c.SetHinting(font.HintingNone)
	case "full":
		c.SetHinting(font.HintingFull)
	}

	// Draw the guidelines.
	for i := 0; i < 200; i++ {
		rgba.Set(10, 10+i, ruler)
		rgba.Set(10+i, 10, ruler)
	}

	// Draw the text.
	pt := freetype.Pt(10, 10+int(c.PointToFixed(*size)>>6))
	for _, s := range text {
		_, err = c.DrawString(s, pt)
		if err != nil {
			log.Println(err)
			return
		}

		pt.Y += c.PointToFixed(*size * *spacing)
	}

	// Save that RGBA image to disk.
	outFile, err := os.Create("out.png")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer outFile.Close()

	// Create a new buffer base on file size.
	fInfo, _ := outFile.Stat()
	var size int64 = fInfo.Size()
	buf := make([]byte, size)

	// Read file content into buffer.
	fReader := bufio.NewReader(outFile)
	fReader.Read(buf)

	png.Encode(&buf, outFile)

	// Convert the buffer bytes to base64 string
	imgBase64Str := base64.StdEncoding.EncodeToString(buf)

	img2html := "<html><body><img src=\"data:image/png;base64," + imgBase64Str + "\" /></body></html>"
	w.Write([]byte(fmt.Sprintf(img2html)))

	// b := bufio.NewWriter(outFile)
	// err = png.Encode(b, rgba)
	// if err != nil {
	// 	log.Println(err)
	// 	os.Exit(1)
	// }
	// err = b.Flush()
	// if err != nil {
	// 	log.Println(err)
	// 	os.Exit(1)
	// }

	// fmt.Println("Wrote out.png OK.")
}

// func drawString(t string) {
// 	dst := image.NewRGBA(image.Rect(0, 0, 400, 300))
// 	draw.Draw(dst, dst.Bounds(), image.White, image.ZP, draw.Src)
// }

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	// http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", mux)
}
