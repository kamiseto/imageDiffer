package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	//"github.com/chai2010/tiff"
	"github.com/hhrutter/tiff"
)

func getColorModel(img image.Image) string {
	//カラーモデルを文字列に変換
	var colorModel1 string
	switch img.ColorModel() {
	case color.RGBAModel:
		colorModel1 = "RGBA"
	case color.RGBA64Model:
		colorModel1 = "RGBA64"
	case color.NRGBAModel:
		colorModel1 = "NRGBA"
	case color.NRGBA64Model:
		colorModel1 = "NRGBA64"
	case color.AlphaModel:
		colorModel1 = "Alpha"
	case color.Alpha16Model:
		colorModel1 = "Alpha16"
	case color.GrayModel:
		colorModel1 = "Gray"
	case color.Gray16Model:
		colorModel1 = "Gray16"
	case color.CMYKModel:
		colorModel1 = "CMYK"
	case color.YCbCrModel:
		colorModel1 = "YCbCr"
	default:
		colorModel1 = "Unknown"
	}
	return colorModel1
}

// goでtiff画像を読み込むには golang.org/x/image/tiff を使う
// go get golang.org/x/image/tiff
// cmyk画像は読み込めない 読み込むには github.com/chai2010/tiff が必要
func reatTiffImage(arg string) (image.Image, string) {
	file1, err := os.Open(arg)
	if err != nil {
		fmt.Println("ファイルが開けませんでした")
		os.Exit(1)
	}
	defer file1.Close()
	img1, err := tiff.Decode(file1)
	if err != nil {
		fmt.Println("tiff画像ファイルではありません", arg)
		os.Exit(1)
	}
	return img1, ".tif"
}

func readImage(arg string) (image.Image, string) {
	ext1 := strings.ToLower(filepath.Ext(arg))

	img1 := image.Image(nil)
	if ext1 == ".tif" {
		return reatTiffImage(arg)
	} else {
		file1, err := os.Open(arg)
		if err != nil {
			fmt.Println("ファイルが開けませんでした")
			os.Exit(1)
		}
		defer file1.Close()
		img1, _, err = image.Decode(file1)
		if err != nil {
			fmt.Println("jpeg画像ファイルではありません")
			os.Exit(1)
		}
	}
	return img1, ext1
}

// 並列処理
func ParallelForEachPixel(size image.Point, f func(x, y int)) {
	max_cpu := runtime.NumCPU()
	fmt.Println("CPU:", max_cpu)
	procs := runtime.GOMAXPROCS(max_cpu)
	var waitGroup sync.WaitGroup
	for i := 0; i < procs; i++ {
		start, end := calcStartEndVal(i, size.X, procs)
		waitGroup.Add(1)
		go func(sX, eX, sY, eY int) {
			defer waitGroup.Done()
			for x := sX; x < eX; x++ {
				for y := sY; y < eY; y++ {
					f(x, y) // (x,y)番地のブロックのマスキング処理
				}
			}
		}(start, end, 0, size.Y)
	}
	waitGroup.Wait() // 全てのgoroutineの処理が終わるまで待機
}

// 処理する範囲を計算
func calcStartEndVal(i, size, procs int) (start, end int) {
	start = size / procs * i
	end = size / procs * (i + 1)
	if i == procs-1 {
		end = size
	}
	return
}

// 画像の保存
func SaveImage(img image.Image, ext1 string, imgname string) {
	out, err := os.Create(imgname)
	if err != nil {
		fmt.Println("ファイルが作成できませんでした")
		os.Exit(1)
	}
	defer out.Close()

	switch ext1 {
	case ".jpg", ".jpeg":
		jpeg.Encode(out, img, nil)
	case ".png":
		png.Encode(out, img)
	case ".tif":
		tiff.Encode(out, img, nil)
	default:
		fmt.Println("対応していないファイル形式です")
		os.Exit(1)
	}
}

// 画像の情報を表示
func ImageInfo(img image.Image, ext1 string, imgname string) {
	//画像のサイズを取得
	bounds1 := img.Bounds()
	width1 := bounds1.Max.X
	height1 := bounds1.Max.Y
	fmt.Println("file:", imgname)
	fmt.Println("width:", width1)
	fmt.Println("height:", height1)
	fmt.Println("color:", getColorModel(img))
	fmt.Println("ext:", ext1)
}

// RGBA画像の場合
func Case_RGBA(img1 image.Image, ext1 string) image.Image {
	fmt.Println("Case_RGBA")
	//画像のサイズを取得
	bounds1 := img1.Bounds()
	width1 := bounds1.Max.X
	height1 := bounds1.Max.Y

	//差分画像のサイズを指定
	diffimg := image.NewRGBA(image.Rect(0, 0, width1, height1))

	//差分画像を作成
	ParallelForEachPixel(image.Point{width1, height1}, func(x, y int) {
		//画像の色を取得
		r1, g1, b1, a1 := img1.At(x, y).RGBA()
		diffimg.Set(x, y, color.RGBA{uint8(g1), uint8(b1), uint8(r1), uint8(a1)})
	})
	return diffimg
}

// CMYK画像の場合
func Case_CMYK(img1 image.Image, ext1 string) image.Image {
	fmt.Println("Case_CMYK")
	//画像のサイズを取得
	bounds1 := img1.Bounds()
	width1 := bounds1.Max.X
	height1 := bounds1.Max.Y

	//差分画像のサイズを指定
	diffimg := image.NewCMYK(image.Rect(0, 0, width1, height1))

	//差分画像を作成
	ParallelForEachPixel(image.Point{width1, height1}, func(x, y int) {
		//画像の色を取得
		c1, m1, y1, k1 := img1.At(x, y).RGBA()
		c1 = 255 - c1
		m1 = 255 - m1
		y1 = 255 - y1
		k1 = 255 - k1
		diffimg.Set(x, y, color.CMYK{uint8(k1), uint8(c1), uint8(m1), uint8(y1)})
	})
	return diffimg
}

// Gray画像の場合
func Case_Gray(img1 image.Image, ext1 string) image.Image {
	fmt.Println("Case_Gray")
	//画像のサイズを取得
	bounds1 := img1.Bounds()
	width1 := bounds1.Max.X
	height1 := bounds1.Max.Y

	//差分画像のサイズを指定
	diffimg := image.NewCMYK(image.Rect(0, 0, width1, height1))

	//差分画像を作成
	ParallelForEachPixel(image.Point{width1, height1}, func(x, y int) {
		//画像の色を取得
		r1, _, _, _ := img1.At(x, y).RGBA()
		//色を反転
		r1 = 255 - r1
		diffimg.Set(x, y, color.CMYK{uint8(r1), 0, 0, 0})
	})
	return diffimg
}

// メイン関数
func main() {
	//画像を読み込む
	flag.Parse()
	args := flag.Args()

	//-hオプションがある場合はヘルプを表示
	if len(args) == 1 && args[0] == "-h" {
		fmt.Println("Usage: imageDiffer [options] file1 file2")
		fmt.Println("Options:")
		fmt.Println("  -h, --help")
		fmt.Println("  -v, --version")
		os.Exit(0)
	}
	//-vオプションがある場合はバージョンを表示
	if len(args) == 1 && args[0] == "-v" {
		fmt.Println("imageDiffer version 1.0.0")
		os.Exit(0)
	}

	//argsを2つづつ処理
	for i := 0; i < len(args); i += 1 {
		main2(args[i])
	}
}

// 処理をまとめた関数
func main2(arg1 string) {

	img1, ext1 := readImage(arg1)

	//画像の情報を表示
	ImageInfo(img1, ext1, arg1)
	//画像のサイズを取得
	bounds1 := img1.Bounds()
	width1 := bounds1.Max.X
	height1 := bounds1.Max.Y

	//img1のカラーがYCbCrの場合はRGBAに変換
	if img1.ColorModel() == color.YCbCrModel {
		img3 := image.NewRGBA(image.Rect(0, 0, width1, height1))
		draw.Draw(img3, img3.Bounds(), img1, bounds1.Min, draw.Src)
		img1 = img3
	}

	switch img1.ColorModel() {
	case color.RGBAModel, color.RGBA64Model, color.NRGBAModel, color.NRGBA64Model:
		SaveImage(Case_RGBA(img1, ext1), ext1, arg1)
	case color.CMYKModel:
		SaveImage(Case_CMYK(img1, ext1), ext1, arg1)
	case color.GrayModel, color.Gray16Model:
		SaveImage(Case_Gray(img1, ext1), ext1, arg1)
	default:
		fmt.Println("対応していない画像形式です")
		os.Exit(1)
	}
}
