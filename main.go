package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/go-ini/ini"
	"github.com/golang/freetype"
	"golang.org/x/image/font"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	dpi      = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile = flag.String("fontfile", "C:\\Windows\\Fonts\\simkai.ttf", "filename of the ttf font")
	hinting  = flag.String("hinting", "none", "none | full")
	size     = flag.Float64("size", 24, "font size in points")
	spacing  = flag.Float64("spacing", 1.5, "line spacing (e.g. 2 means double spaced)")
	wonb     = flag.Bool("whiteonblack", false, "white text on a black background")
)

const (
	dataFile            = "my.ini"
	textSection         = "text"
	textWords           = "words"
	textSeparator       = ","
	iniPackageSeparator = "."
)

type Text struct {
	Content string
	X, Y    int
}

func main() {
	data, err := ini.Load(dataFile)
	if os.IsNotExist(err) {
		fmt.Printf("装载信息的文件%q不存在\r\n", dataFile)
		os.Exit(1)
	} else if err != nil {
		fmt.Printf("程序读取装载信息的文件出错:%v\r\n", err)
		os.Exit(1)
	}
	// 打开需要画的图片
	backgroundImageFile, err := os.Open("1.jpg")
	if err != nil {
		fmt.Println("打开模板文件失败:%v", err)
		os.Exit(1)
	}
	backgroundImage, err := jpeg.Decode(backgroundImageFile)
	if err != nil {
		fmt.Println("解码模板文件失败:%v", err)
		os.Exit(1)
	}
	backgroundImageBounds := backgroundImage.Bounds()
	newBackgroundImage := image.NewRGBA(backgroundImageBounds)
	// 把模板文件画到新的图片文件上
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
	draw.Draw(newBackgroundImage, backgroundImageBounds, backgroundImage, image.ZP, draw.Src)
	textSectionData := data.Section(textSection)
	c := freetype.NewContext()
	c.SetDPI(*dpi)
	c.SetFont(f)
	c.SetFontSize(*size)
	c.SetClip(backgroundImageBounds)
	c.SetDst(newBackgroundImage)
	c.SetSrc(image.Black)
	c.SetHinting(font.HintingNone)
	// 画文本
	for _, key := range textSectionData.KeyStrings() {
		if key == textWords {
			textSectionValue := textSectionData.Key(key).String()
			textWordsData := strings.Split(strings.Replace(textSectionValue, "，", textSeparator, -1), textSeparator)
			for _, word := range textWordsData {
				textData := data.Section(textSection + iniPackageSeparator + word)
				fmt.Printf("Section:%v\r\n", textData.Name())
				keyStrings := textData.KeyStrings()
				imageText := Text{}
				for _, key := range keyStrings {
					fmt.Printf("key:%v, value:%v\r\n", key, textData.Key(key).String())
					// 画出文字
					switch key {
					case word:
						imageText.Content = textData.Key(key).String()
					case "位置":
						position := strings.Split(textData.Key(key).String(), ",")
						imageText.X, err = strconv.Atoi(position[0])
						if err != nil {
							fmt.Println("输入的位置不是数字!:%v", err)
							os.Exit(1)
						}
						imageText.Y, err = strconv.Atoi(position[1])
						if err != nil {
							fmt.Println("输入的位置不是数字!:%v", err)
							os.Exit(1)
						}
					}
				}
				if imageText.Content == "" && imageText.X == 0 && imageText.Y == 0 {
					continue
				}
				pt := freetype.Pt(imageText.X, imageText.Y)
				p, err := c.DrawString(imageText.Content, pt)
				if err != nil {
					fmt.Println("画文本失败!:%v", err)
					os.Exit(1)
				}
				fmt.Printf("%#v\r\n", p)
			}
		}
	}
	// 画圆
	for i := 0; i < 10; i++ {
		if i < 5 {
			drawCircle(newBackgroundImage, 580, 610, 90+i, color.RGBA{R: 189, G: 59, B: 25, A: 0xff}, c)
		} else {
			//drawCircle(newBackgroundImage, 580, 610, 90+i-5, color.RGBA{R:170, G:76, B:51})
		}
	}
	// 校长的章
	//for i:=750; i<800; i++ {
	//	newBackgroundImage.Set(i+1, 540, colornames.Red)
	//	newBackgroundImage.Set(i+1, 540+1, colornames.Red)
	//}
	//for i:=540; i<650; i++ {
	//	newBackgroundImage.Set(750, i, colornames.Red)
	//	newBackgroundImage.Set(750+1, i, colornames.Red)
	//}
	//for i:=750; i<800; i++ {
	//	newBackgroundImage.Set(i+1, 649, colornames.Red)
	//	newBackgroundImage.Set(i+1, 650, colornames.Red)
	//}
	//for i:=540; i<650; i++ {
	//	newBackgroundImage.Set(800, i, colornames.Red)
	//	newBackgroundImage.Set(800+1, i, colornames.Red)
	//}
	drawRect(newBackgroundImage, image.Point{X: 760, Y: 570}, image.Point{X: 830, Y: 640}, color.RGBA{R: 157, G: 85, B: 50, A: 0xff}, c)
	// 保存
	saveImage, err := os.Create("a.jpg")
	if err != nil {
		fmt.Println("创建文件失败:%v", err)
		os.Exit(1)
	}
	defer saveImage.Close()
	saveImageWriter := bufio.NewWriter(saveImage)
	err = jpeg.Encode(saveImageWriter, newBackgroundImage, &jpeg.Options{Quality: 75})
	if err != nil {
		fmt.Println("保存文件失败:%v", err)
		os.Exit(1)
	}
}

// 画校长的章
func drawRect(img draw.Image, p1 image.Point, p2 image.Point, col color.Color, context *freetype.Context) {
	// 先画横的
	x := p1.X
	x1 := p2.X
	for x <= x1 {
		img.Set(x, p1.Y, col)
		img.Set(x, p2.Y, col)
		img.Set(x, p1.Y+1, col)
		img.Set(x, p2.Y+1, col)
		img.Set(x, p1.Y+2, col)
		img.Set(x, p2.Y+2, col)
		img.Set(x, p1.Y+3, col)
		img.Set(x, p2.Y+3, col)
		x++
	}
	// 画竖的
	y := p1.Y
	y1 := p2.Y
	for y <= y1 {
		img.Set(p1.X, y, col)
		img.Set(p2.X, y, col)
		img.Set(p1.X+1, y, col)
		img.Set(p2.X-1, y, col)
		img.Set(p1.X+2, y, col)
		img.Set(p2.X-2, y, col)
		img.Set(p1.X+3, y, col)
		img.Set(p2.X-3, y, col)
		y++
	}
	// 画字
	newPoint1 := image.Point{X: p1.X - 5, Y: p1.Y - 5}
	//newPoint2 := image.Point{X: p2.X-5, Y: p2.Y-5}
	//context.SetSrc(image.NewUniform(color.RGBA{R:230, G:190, B:150}))
	context.SetSrc(image.NewUniform(color.RGBA{R: 187, G: 80, B: 60, A: 0xff}))
	context.SetFontSize(35)

	context.DrawString("陈", freetype.Pt(newPoint1.X+39, newPoint1.Y+40))
	context.DrawString("中", freetype.Pt(newPoint1.X+39, newPoint1.Y+70))
	context.DrawString("江", freetype.Pt(newPoint1.X+7, newPoint1.Y+40))
	context.DrawString("印", freetype.Pt(newPoint1.X+7, newPoint1.Y+70))
}

func drawCircle(img draw.Image, x0, y0, r int, c color.Color, context *freetype.Context) {
	x, y, dx, dy := r-1, 0, 1, 1
	err := dx - (r * 2)

	for x > y {
		img.Set(x0+x, y0+y, c)
		img.Set(x0+y, y0+x, c)
		img.Set(x0-y, y0+x, c)
		img.Set(x0-x, y0+y, c)
		img.Set(x0-x, y0-y, c)
		img.Set(x0-y, y0-x, c)
		img.Set(x0+y, y0-x, c)
		img.Set(x0+x, y0-y, c)

		if err <= 0 {
			y++
			err += dy
			dy += 2
		}
		if err > 0 {
			x--
			dx += 2
			err += dx - (r * 2)
		}
	}
	// 写字
	//context.
}
