package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/disintegration/imaging"
)

func getFiles(dir string) []string {
	files := make([]string, 0)

	filepath.Walk(dir, func(v string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		if path.Ext(f.Name()) == ".png" {
			files = append(files, path.Join(dir, f.Name()))
		}

		return nil
	})

	sort.Strings(files)

	return files
}

func W(f *os.File, format string, s ...interface{}) {
	ss := fmt.Sprintf(format, s...)
	// fmt.Println(ss)
	f.WriteString(ss + "\n")
}

type ImagePair struct {
	Name     rune
	FileName string
	Image    image.Image
}

var dir = flag.String("d", ".", "dir")
var name = flag.String("n", "", "name")
var skip = flag.String("s", "", "remove")

func NameDetect(imageName string) rune {

	ext := filepath.Ext(imageName)

	imageName = imageName[:len(imageName)-len(ext)]

	if *skip != "" {
		imageName = strings.Replace(imageName, *skip, "", -1)
	}

	return []rune(imageName)[0]
}

//http://www.angelcode.com/products/bmfont/doc/file_format.html
func main() {
	flag.Parse()

	absPath, err := filepath.Abs(*dir)
	if err != nil {
		panic(err)
	}

	if *name == "" {
		*name = filepath.Base(absPath)
	}

	files := getFiles(*dir)

	images := make([]ImagePair, len(files))
	totalW := 0
	totalH := 0
	maxH := 0
	for i, name := range files {
		k := NameDetect(filepath.Base(name))

		img, err := imaging.Open(name)

		if err != nil {
			fmt.Println("图片" + name)
			panic(err)
		}

		ip := ImagePair{
			Name:     k,
			Image:    img,
			FileName: name,
		}

		images[i] = ip

		w := img.Bounds().Max.X
		h := img.Bounds().Max.Y

		totalW += w
		totalH += h
		if h > maxH {
			maxH = h
		}
	}

	avgW := totalW / len(files)
	avgH := totalH / len(files)

	dest := imaging.New(totalW, maxH, color.Black)
	f, _ := os.Create(*name + ".fnt")
	//
	W(f, "info face=\"Arial\" size=%d bold=0 italic=0 charset=\"\" unicode=1 stretchH=100 smooth=1 aa=1 padding=0,0,0,0 spacing=1,1 outline=0", avgH)
	W(f, "common lineHeight=%d base=%d scaleW=%d scaleH=%d pages=1 packed=0 alphaChnl=1 redChnl=0 greenChnl=0 blueChnl=0", avgH, avgH, totalW, maxH)
	W(f, "page id=0 file=\"%s.png\"", *name)
	W(f, "chars count=%d", len(files))

	ofx := 0
	for _, pair := range images {
		img := pair.Image
		k := pair.Name
		w := img.Bounds().Max.X
		h := img.Bounds().Max.Y
		x := ofx
		y := 0

		dest = imaging.Paste(dest, img, image.Pt(x, y))
		ofx += w

		fmt.Println(fmt.Sprintf("%s => %s => %d", pair.FileName, string(pair.Name), int(pair.Name)))

		W(f, "char id=%d x=%d y=%d width=%d height=%d xoffset=%d yoffset=%d xadvance=%d page=0  chnl=15", int(k), x, y, w, h, (avgW-w)/2, -h/2+maxH/2, w+(avgW-w)/2)
	}

	if err := imaging.Save(dest, *name+".png"); err != nil {
		panic(err)
	}
}
