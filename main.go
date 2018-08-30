package main

import (
	"bufio"
	"fmt"
	"github.com/go-ini/ini"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
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
	globalFontPath string
	globalFontSize float64
)

const (
	dpi                 = 72
	dataFile            = "my.ini"
	textSection         = "text"
	textWords           = "words"
	textSeparator       = ","
	iniPackageSeparator = "."
	chapterOfPrincipal  = "校长的章"
	imageSection        = "image"
	fontFilePath        = "C:\\Windows\\Fonts\\"
)

type Text struct {
	Content, FontName string
	X, Y              int
	FontSize          float64
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
		fmt.Printf("打开模板文件失败:%v\r\n", err)
		os.Exit(1)
	}
	backgroundImage, err := jpeg.Decode(backgroundImageFile)
	if err != nil {
		fmt.Printf("解码模板文件失败:%v\r\n", err)
		os.Exit(1)
	}
	backgroundImageBounds := backgroundImage.Bounds()
	newBackgroundImage := image.NewRGBA(backgroundImageBounds)
	// 把模板文件画到新的图片文件上
	draw.Draw(newBackgroundImage, backgroundImageBounds, backgroundImage, image.ZP, draw.Src)
	textSectionData := data.Section(textSection)
	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetClip(backgroundImageBounds)
	c.SetDst(newBackgroundImage)
	c.SetSrc(image.Black)
	c.SetHinting(font.HintingNone)
	// 先获取全局的字体和字体大小
	globalFontFile := textSectionData.Key("字体").String()
	globalFontPath = fontFilePath + globalFontFile
	if err != nil {
		panic("全局字体不存在")
	}
	globaleFontSizeStr := textSectionData.Key("字体大小").String()
	globalFontSize, err = strconv.ParseFloat(globaleFontSizeStr, 64)
	if err != nil {
		panic("全局字体大小请输入数字!")
	}
	// 画文本
	for _, key := range textSectionData.KeyStrings() {
		if key == textWords {
			textSectionValue := textSectionData.Key(key).String()
			textWordsData := strings.Split(strings.Replace(textSectionValue, "，", textSeparator, -1), textSeparator)
			for _, word := range textWordsData {
				textData := data.Section(textSection + iniPackageSeparator + word)
				keyStrings := textData.KeyStrings()
				imageText := Text{}
				for _, key := range keyStrings {
					// 画出文字
					switch key {
					case word:
						imageText.Content = textData.Key(key).String()
					case "位置":
						position := strings.Split(textData.Key(key).String(), ",")
						imageText.X, err = strconv.Atoi(position[0])
						if err != nil {
							fmt.Printf("输入的位置不是数字!:%v\r\n", err)
							os.Exit(1)
						}
						imageText.Y, err = strconv.Atoi(position[1])
						if err != nil {
							fmt.Printf("输入的位置不是数字!:%v\r\n", err)
							os.Exit(1)
						}
					case "字体":
						imageText.FontName = textData.Key(key).String()
					case "字体大小":
						fontSizeStr := textData.Key(key).String()
						fontSize, err := strconv.ParseFloat(fontSizeStr, 64)
						if err != nil {
							panic("字体大小请输入数字!")
						}
						imageText.FontSize = fontSize
					}
				}
				if imageText.Content == "" && imageText.X == 0 && imageText.Y == 0 {
					continue
				}
				fontSize := imageText.FontSize
				if fontSize == 0.00 {
					fontSize = globalFontSize
				}
				fontFile := imageText.FontName
				if fontFile == "" {
					fontFile = globalFontPath
				} else {
					fontFile = fontFilePath + imageText.FontName
				}
				imageFont, err := getFont(fontFile)
				// 设置字体
				if err != nil {
					panic(fmt.Sprintf("读取字体%q文件出错:%v", fontFilePath+imageText.FontName, err))
				}
				c.SetFont(imageFont)
				c.SetFontSize(fontSize)
				pt := freetype.Pt(imageText.X, imageText.Y)
				_, err = c.DrawString(imageText.Content, pt)
				if err != nil {
					fmt.Printf("画文本失败!:%v\r\n", err)
					os.Exit(1)
				}
			}
		}
	}
	// 画圆
	//for i := 0; i < 4; i++ {
	//	drawCircle(newBackgroundImage, 580, 610, 90+i, color.RGBA{R: 189, G: 59, B: 25, A: 0xff}, c)
	//}
	chapter := data.Section(imageSection + "." + chapterOfPrincipal)
	// 校长的章
	drawRect(newBackgroundImage, image.Point{X: 760, Y: 570}, image.Point{X: 830, Y: 640}, color.RGBA{R: 157, G: 85, B: 50, A: 0xff}, c, chapter)
	// 保存
	saveImage, err := os.Create("a.jpg")
	if err != nil {
		fmt.Printf("创建文件失败:%v\r\n", err)
		os.Exit(1)
	}
	defer saveImage.Close()
	saveImageWriter := bufio.NewWriter(saveImage)
	err = jpeg.Encode(saveImageWriter, newBackgroundImage, &jpeg.Options{Quality: 75})
	if err != nil {
		fmt.Printf("保存文件失败:%v\r\n", err)
		os.Exit(1)
	}
	fmt.Println("success!")
}
func getFont(fontFilePath string) (*truetype.Font, error) {
	fontBytes, err := ioutil.ReadFile(fontFilePath)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return freetype.ParseFont(fontBytes)
}

// 画校长的章
func drawRect(img draw.Image, p1 image.Point, p2 image.Point, col color.Color, context *freetype.Context, chapterSection *ini.Section) {
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
	context.SetSrc(image.NewUniform(color.RGBA{R: 187, G: 80, B: 60, A: 0xff}))
	context.SetFontSize(35)
	names := strings.Split(chapterSection.Key("校长的章").String(), "")
	context.DrawString(names[0], freetype.Pt(newPoint1.X+39, newPoint1.Y+40))
	context.DrawString(names[1], freetype.Pt(newPoint1.X+39, newPoint1.Y+70))
	context.DrawString(names[2], freetype.Pt(newPoint1.X+7, newPoint1.Y+40))
	context.DrawString(names[3], freetype.Pt(newPoint1.X+7, newPoint1.Y+70))
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
