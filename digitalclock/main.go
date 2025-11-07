// //go:build !solution

// package main

// import (
// 	"fmt"
// 	"image"
// 	"image/color"
// 	"image/draw"
// 	"image/png"
// 	"net/http"
// 	"os"
// 	"strconv"
// 	"strings"
// 	"time"
// )

// const (
// 	defaultPort   = "8889"
// 	defaultHeight = 56
// 	defaultWidht  = 12
// )

// // func parseTimeString(t string) (h, m, s int) {
// // 	parts := strings.Split(t, ":")
// // 	h = strconv.Atoi(parts[])
// // 	h, m, s = parts[0], parts[1], parts[2]
// // 	return h, m, s
// // }

// func getFormatedTime(t string) string {
// 	if _, err := time.Parse(time.TimeOnly, t); err != nil {
// 		now := time.Now()
// 		t = now.Format(time.TimeOnly)
// 	}
// 	return t
// }

// func validateScale(kstr string) (int, error) {
// 	if kstr == "" {
// 		return 1, nil
// 	}

// 	k, err := strconv.Atoi(kstr)

// 	if err != nil {
// 		return -1, fmt.Errorf("invalid k param")
// 	}

// 	if k < 1 || k > 30 {
// 		return -1, fmt.Errorf("k have to be from 1 to 30")
// 	}

// 	return k, nil
// }

// func TimeHandler(w http.ResponseWriter, r *http.Request) {
// 	q := r.URL.Query()
// 	t := getFormatedTime(q.Get("time"))
// 	k, err := validateScale(q.Get("k"))

// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	_ = k
// 	img := createImageFromTime(t)
// 	w.Header().Set("Content-Type", "image/png")
// 	png.Encode(w, img)

// }

// func drawNumber(img draw.Image, numRepresentation string, bias int) draw.Image {
// 	matrix := strings.Split(numRepresentation, "\n")
// 	rect := image.Rect(0, 0, len(matrix[0]), len(matrix))
// 	numImg := image.NewRGBA(rect)

// 	for i, lines := range matrix {
// 		for j, s := range lines {
// 			if s == '1' {
// 				numImg.Set(j, i, Cyan)
// 				continue
// 			}
// 			numImg.Set(j, i, color.White)
// 		}
// 	}

// 	new_png_file := "two_rectangles.png" // output image will live here

// 	myfile, err := os.Create(new_png_file) // ... now lets save output image
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer myfile.Close()
// 	png.Encode(myfile, numImg) // output file /tmp/two_rectangles.png

// 	ofssetRect := rect.Bounds().Add(image.Point{bias, 0})
// 	draw.Draw(img, ofssetRect, numImg, image.Point{0, 0}, draw.Src)
// 	return img
// }

// func createImageFromTime(tstr string) draw.Image {
// 	baseOffset := 8
// 	colonOffset := 4
// 	base := image.NewRGBA(image.Rect(0, 0, 56, 12))
// 	parts := strings.Split(tstr, ":")
// 	bias := 0
// 	hours := parts[0]
// 	if len(hours) == 1 {
// 		parts[0] = "0" + parts[0]
// 	}
// 	for i, val := range parts {
// 		for j, num := range val {
// 			numstr, err := getNumStrRepr(string(num))
// 			if err != nil {
// 				fmt.Println(err)
// 			}
// 			drawNumber(base, numstr, bias)
// 			bias += baseOffset
// 			if j == len(val)-1 {
// 				continue
// 			}
// 		}
// 		if i == len(parts)-1 {
// 			continue
// 		}
// 		numstr, err := getNumStrRepr(":")
// 		if err != nil {
// 			fmt.Println(err)
// 		}
// 		drawNumber(base, numstr, bias)
// 		bias += colonOffset

// 	}
// 	return base
// }

// func resizeImage(img draw.Image, ratio int) draw.Image {
// 	resizeFactor := 1 / float32(ratio)
// 	bounds := img.Bounds()
// 	height := bounds.Dy()
// 	width := bounds.Dx()
// 	width = width * ratio
// 	height = height * ratio
// 	resizedImage := image.NewRGBA(image.Rect(0, 0, width, height))
// 	var imgX, imgY int
// 	var imgColor color.Color
// 	for x := 0; x < width; x++ {
// 		for y := 0; y < height; y++ {
// 			imgX = int(resizeFactor*float32(x) + 0.5)
// 			imgY = int(resizeFactor*float32(y) + 0.5)
// 			imgColor = img.At(imgX, imgY)
// 			resizedImage.Set(x, y, imgColor)
// 		}
// 	}
// 	return resizedImage
// }

// func fdraw(fpath string) {
// 	t, _ := time.Parse(time.TimeOnly, "9:04:43")
// 	tstr := t.Format(time.TimeOnly)
// 	img := createImageFromTime(tstr)
// 	resimg := resizeImage(img, 7)
// 	f, _ := os.Create(fpath)
// 	png.Encode(f, resimg)
// }

// func main() {
// 	fdraw("img.png")
// }

// // // Main main func
// // func main() {
// // 	portStr := flag.String("port", defaultPort, "port on which server will listen")
// // 	flag.Parse()

// // 	http.HandleFunc("/", TimeHandler)
// // 	http.ListenAndServe(":"+*portStr, nil)

// // }

//go:build !solution

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort = "8889"
)

func validateTime(t string) (string, error) {
	if t == "" {
		return time.Now().Format(time.TimeOnly), nil
	}
	parts := strings.Split(t, ":")
	if len(parts) < 2 || len(parts[0]) < 2 {
		return "", fmt.Errorf("invalid time format")
	}

	var err error
	parsedt, err := time.Parse(time.TimeOnly, t)

	if err != nil {
		return "", fmt.Errorf("invalid time format")
	}
	return parsedt.Format(time.TimeOnly), nil
}

func validateScale(kstr string) (int, error) {
	if kstr == "" {
		return 1, nil
	}
	k, err := strconv.Atoi(kstr)
	if err != nil {
		return -1, fmt.Errorf("invalid k param")
	}
	if k < 1 || k > 30 {
		return -1, fmt.Errorf("k must be from 1 to 30")
	}
	return k, nil
}

func drawSymbol(img draw.Image, repr string, xBias int) int {
	lines := strings.Split(repr, "\n")
	h := len(lines)
	w := len(lines[0])

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			switch lines[y][x] {
			case '1':
				img.Set(xBias+x, y, Cyan)
			case '.':
				img.Set(xBias+x, y, color.White)
			}
		}
	}
	return w
}

func createImageFromTime(tstr string) draw.Image {
	parts := strings.Split(tstr, ":")
	if len(parts[0]) == 1 {
		parts[0] = "0" + parts[0]
	}

	h := strings.Count(Zero, "\n") + 1
	w := len(strings.Split(Zero, "\n")[0])
	wColon := len(strings.Split(Colon, "\n")[0])

	totalWidth := 6*w + 2*wColon
	totalHeight := h

	base := image.NewRGBA(image.Rect(0, 0, totalWidth, totalHeight))

	draw.Draw(base, base.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	bias := 0
	for i, val := range parts {
		for _, num := range val {
			repr, _ := getNumStrRepr(string(num))
			shift := drawSymbol(base, repr, bias)
			bias += shift
		}
		if i < len(parts)-1 {
			repr, _ := getNumStrRepr(":")
			shift := drawSymbol(base, repr, bias)
			bias += shift
		}
	}

	return base
}

func resizeImage(img draw.Image, k int) draw.Image {
	if k == 1 {
		return img
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	newW, newH := w*k, h*k
	out := image.NewRGBA(image.Rect(0, 0, newW, newH))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := img.At(x, y)
			for dy := 0; dy < k; dy++ {
				for dx := 0; dx < k; dx++ {
					out.Set(x*k+dx, y*k+dy, c)
				}
			}
		}
	}
	return out
}

func TimeHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	t, err := validateTime(q.Get("time"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	k, err := validateScale(q.Get("k"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	img := createImageFromTime(t)
	img = resizeImage(img, k)

	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, img)
}

func main() {
	portStr := flag.String("port", defaultPort, "port on which server will listen")
	flag.Parse()

	http.HandleFunc("/", TimeHandler)
	http.ListenAndServe(":"+*portStr, nil)
}
