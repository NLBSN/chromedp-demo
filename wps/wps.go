package wps

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	goQrcode "github.com/skip2/go-qrcode"
	"image"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const loginURL = "https://account.wps.cn/"

// 自定义任务
func MyTasks() chromedp.Tasks {
	return chromedp.Tasks{
		// 0. 加载cookies
		loadCookies(),

		// 1. 打开金山文档的登陆界面
		chromedp.Navigate(loginURL),

		// 判断一下是否已经登陆
		checkLoginStatus(),

		// 2. 点击微信登陆按钮
		// #wechat > span:nth-child(2)
		chromedp.Click(`#wechat > span:nth-child(2)`),

		// 3. 点击确认按钮
		// #dialog > div.dialog-wrapper > div > div.dialog-footer > div.dialog-footer-ok
		chromedp.Click(`#dialog > div.dialog-wrapper > div > div.dialog-footer > div.dialog-footer-ok`),

		// 4. 获取二维码
		// #wximport
		getCode(),

		// 5. 若二维码登录后，浏览器会自动跳转到用户信息页面
		saveCookies(),
	}
}

// 获取二维码的过程
func getCode() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		// 1. 用于存储图片的字节切片
		var code []byte

		// 2. 截图
		// 注意这里需要注明直接使用ID选择器来获取元素（chromedp.ByID）
		if err = chromedp.Screenshot(`#wximport`, &code, chromedp.ByID).Do(ctx); err != nil {
			return
		}

		// 3. 把二维码输出到标准输出流
		if err = printQRCode(code); err != nil {
			return err
		}
		return
	}
}

// 输出二维码
func printQRCode(code []byte) (err error) {
	// 1. 因为我们的字节流是图像，所以我们需要先解码字节流
	img, _, err := image.Decode(bytes.NewReader(code))
	if err != nil {
		return
	}

	// 2. 然后使用gozxing库解码图片获取二进制位图
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return
	}

	// 3. 用二进制位图解码获取gozxing的二维码对象
	res, err := qrcode.NewQRCodeReader().Decode(bmp, nil)
	if err != nil {
		return
	}

	// 4. 用结果来获取go-qrcode对象（注意这里我用了库的别名）
	qr, err := goQrcode.New(res.String(), goQrcode.High)
	if err != nil {
		return
	}

	// 5. 输出到标准输出流
	fmt.Println(qr.ToSmallString(false))

	return
}

// 保存Cookies
func saveCookies() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		// 等待二维码登陆
		if err = chromedp.WaitVisible(`#app`, chromedp.ByID).Do(ctx); err != nil {
			return
		}

		// cookies的获取对应是在devTools的network面板中
		// 1. 获取cookies
		cookies, err := network.GetAllCookies().Do(ctx)
		if err != nil {
			return
		}

		// 2. 序列化
		cookiesData, err := network.GetAllCookiesReturns{Cookies: cookies}.MarshalJSON()
		if err != nil {
			return
		}

		// 3. 存储到临时文件
		if err = ioutil.WriteFile("cookies.tmp", cookiesData, 0755); err != nil {
			return
		}
		return
	}
}

//file, _ := ioutil.ReadFile("aistudio.tmp")
//fmt.Println("cookid:", string(file))
// 加载Cookies
func loadCookies() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		// 如果cookies临时文件不存在则直接跳过
		if _, _err := os.Stat("cookies.tmp"); os.IsNotExist(_err) {
			return
		}

		// 如果存在则读取cookies的数据
		cookiesData, err := ioutil.ReadFile("cookies.tmp")
		if err != nil {
			return
		}

		// 反序列化
		cookiesParams := network.SetCookiesParams{}
		if err = cookiesParams.UnmarshalJSON(cookiesData); err != nil {
			return
		}

		// 设置cookies
		return network.SetCookies(cookiesParams.Cookies).Do(ctx)
	}
}

// 检查是否登陆
func checkLoginStatus() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		var url string
		if err = chromedp.Evaluate(`window.location.href`, &url).Do(ctx); err != nil {
			return
		}
		if strings.Contains(url, "https://account.wps.cn/usercenter/apps") {
			log.Println("已经使用cookies登陆")
			chromedp.Stop()
		}
		return
	}
}
