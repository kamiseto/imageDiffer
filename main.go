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


	if len(args) < 2 {
		fmt.Println("画像ファイルを2つ指定してください")
		os.Exit(1)
	}
	//argsを2つづつ処理
	for i := 0; i < len(args); i += 2 {
		if len(args) == i+1 {
			fmt.Println("画像ファイルを2つ指定してください")
			os.Exit(1)
		}
		main2(args[i], args[i+1])
	}
}

//処理をまとめた関数
func main2(arg1 string, arg2 string) {

	img1, ext1 := readImage(arg1)
	img2, ext2 := readImage(arg2)

	//画像の情報を表示
	ImageInfo(img1, ext1, arg1)
	ImageInfo(img2, ext2, arg2)
	//画像のサイズを取得
	bounds1 := img1.Bounds()
	width1 := bounds1.Max.X
	height1 := bounds1.Max.Y

	bounds2 := img2.Bounds()
	width2 := bounds2.Max.X
	height2 := bounds2.Max.Y

	//画像のサイズが違う場合はエラー
	if width1 != width2 || height1 != height2 {
		fmt.Println("画像のサイズが違います")
		os.Exit(1)
	}

	//img1のカラーがYCbCrの場合はRGBAに変換
	if img1.ColorModel() == color.YCbCrModel {
		img3 := image.NewRGBA(image.Rect(0, 0, width1, height1))
		draw.Draw(img3, img3.Bounds(), img1, bounds1.Min, draw.Src)
		img1 = img3
	}
	//img2のカラーがYCbCrの場合はRGBAに変換
	if img2.ColorModel() == color.YCbCrModel {
		img4 := image.NewRGBA(image.Rect(0, 0, width2, height2))
		draw.Draw(img4, img4.Bounds(), img2, bounds2.Min, draw.Src)
		img2 = img4
	}

	//画像の色が違う場合はエラー
	if img1.ColorModel() != img2.ColorModel() {
		fmt.Println("画像の色が違います")
		os.Exit(1)
	}
	switch img1.ColorModel() {
	case color.RGBAModel, color.RGBA64Model, color.NRGBAModel, color.NRGBA64Model:
	SaveImage(Case_RGBA(img1, img2, ext1, ext2), ext1, arg1)
	case color.CMYKModel:
	SaveImage(Case_CMYK(img1, img2, ext1, ext2), ext1, arg1)
	case color.GrayModel, color.Gray16Model:
	SaveImage(Case_Gray(img1, img2, ext1, ext2), ext1, arg1)
	default:
		fmt.Println("対応していない画像形式です")
		os.Exit(1)
	}
}

// 画像の保存
func SaveImage(img image.Image, ext1 string, imgname string) {
	out, err := os.Create(imgname + "_diff" + ext1)
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
func Case_RGBA(img1 image.Image, img2 image.Image, ext1 string, ext2 string) image.Image {
	//画像のサイズを取得
	bounds1 := img1.Bounds()
	width1 := bounds1.Max.X
	height1 := bounds1.Max.Y

	//差分画像のサイズを指定
	diffimg := image.NewGray(image.Rect(0, 0, width1, height1))

	//差分画像を作成
	ParallelForEachPixel(image.Point{width1, height1}, func(x, y int) {
		//画像の色を取得
		r1, g1, b1, a1 := img1.At(x, y).RGBA()
		r2, g2, b2, a2 := img2.At(x, y).RGBA()
		//画像の色が違う場合は赤色にする
		if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
			diffimg.Set(x, y, color.Gray{0})
		} else {
			diffimg.Set(x, y, color.Gray{255})
		}
	})
	return diffimg
}

//CMYK画像の場合
func Case_CMYK(img1 image.Image, img2 image.Image, ext1 string, ext2 string) image.Image {
	fmt.Println("Case_CMYK")
	//画像のサイズを取得
	bounds1 := img1.Bounds()
	width1 := bounds1.Max.X
	height1 := bounds1.Max.Y

	//差分画像のサイズを指定
	diffimg := image.NewGray(image.Rect(0, 0, width1, height1))

	//差分画像を作成
	ParallelForEachPixel(image.Point{width1, height1}, func(x, y int) {
		//画像の色を取得
		c1, m1, y1, k1 := img1.At(x, y).RGBA()
		c2, m2, y2, k2 := img2.At(x, y).RGBA()
		//画像の色が違う場合は赤色にする
		if c1 != c2 || m1 != m2 || y1 != y2 || k1 != k2 {
			diffimg.Set(x, y, color.Gray{0})
		} else {
			diffimg.Set(x, y, color.Gray{255})
		}
	})
	return diffimg
}

// Gray画像の場合
func Case_Gray(img1 image.Image, img2 image.Image, ext1 string, ext2 string) image.Image {
	//画像のサイズを取得
	bounds1 := img1.Bounds()
	width1 := bounds1.Max.X
	height1 := bounds1.Max.Y

	//差分画像のサイズを指定
	diffimg := image.NewGray(image.Rect(0, 0, width1, height1))

	//差分画像を作成
	ParallelForEachPixel(image.Point{width1, height1}, func(x, y int) {
		//画像の色を取得
		r1, _, _, _ := img1.At(x, y).RGBA()
		r2, _, _, _ := img2.At(x, y).RGBA()
		//画像の色が違う場合は赤色にする
		if r1 != r2 {
			diffimg.Set(x, y, color.Gray{0})
		} else {
			diffimg.Set(x, y, color.Gray{255})
		}
	})
	return diffimg
}

//wasm sample.js
//go build -o sample.wasm -tags=js
//javascript
//const go = new Go();
//WebAssembly.instantiateStreaming(fetch("sample.wasm"), go.importObject).then((result) => {
//  go.run(result.instance);
//});
