package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"runtime"

	"github.com/go-fsnotify/fsnotify"
	"github.com/lazywei/go-opencv/opencv"
)

func brightnessAdjust(image *opencv.IplImage, b float64) {
	w := image.Width()
	h := image.Height()
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			color := image.Get2D(i, j)
			v := color.Val()
			for k := 0; k < 3; k++ {
				v[k] *= b
				if v[k] > 255 {
					v[k] = 255
				} else if v[k] < 0 {
					v[k] = 0
				}
			}
			image.Set2D(i, j, opencv.NewScalar(v[0], v[1], v[2], v[3]))
		}
	}
}

var (
	win1     = opencv.NewWindow("ar-lucky-money")
	resuffix = regexp.MustCompile(`\.(?i:png)|(?i:jpg)`)
)

func fuckTheLuckyMoney(filename string) {
	if !resuffix.Match([]byte(filename)) {
		return
	}
	image := opencv.LoadImage(filename)
	if image == nil {
		return
	}
	fmt.Println(filename)
	// 红包在图片中的位置
	if image.Width() != image.Height() {
		x := image.Width() * 366 / 1242
		y := image.Height() * 1164 / 2208
		w := image.Width() * 510 / 1242
		image = opencv.Crop(image, x, y, w, w)
	}
	image = opencv.Resize(image, 510, 510, opencv.CV_INTER_LINEAR)
	inpaint_mask := opencv.LoadImage("mask.png", opencv.CV_LOAD_IMAGE_GRAYSCALE)
	dst := image.Clone()
	opencv.Inpaint(image, inpaint_mask, dst, 3, opencv.CV_INPAINT_TELEA)
	brightnessAdjust(dst, 1.3)
	win1.ShowImage(dst)

	// 识别后删除文件
	os.Remove(filename)

	image.Release()
	inpaint_mask.Release()
	dst.Release()
}

func main() {
	_, currentfile, _, _ := runtime.Caller(0)
	if len(os.Args) == 2 {
		currentfile = os.Args[1]
	}
	dir := path.Dir(currentfile)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
					fuckTheLuckyMoney(event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()
	err = watcher.Add(dir)
	fmt.Println("watching:", dir)
	if err != nil {
		log.Fatal(err)
	}
	// 读取文件
	win1.Move(200, 500)
	fmt.Println(dir, currentfile)
	fuckTheLuckyMoney(currentfile)

	opencv.WaitKey(0)
	os.Exit(0)
}
