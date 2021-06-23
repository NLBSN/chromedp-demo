package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/axgle/mahonia"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

var Enc = mahonia.NewEncoder("GBK")

type AutoTranI interface {
	Loginx() error
	LoadCookies() chromedp.ActionFunc
	LoadCookiesExistRemove() error
	SaveCookie() chromedp.ActionFunc
	UploadFiles(files string)
	Task()
	WriteInfo(str string)
}

func (s *AutoTranS) PrintErr(err string) {
	if err != "" {
		fmt.Println("error:", err)
	}
	fmt.Println("error file uploaded fail:", Enc.ConvertString(s.File))
}

func (s *AutoTranS) WaitTimes(timeout int) {
	s.Ctx, s.Cancel = context.WithTimeout(s.Ctx, time.Second*time.Duration(timeout))
	go func() {
		select {
		//case <-time.After(time.Minute):
		//case <-time.After(time.Minute):
		//	fmt.Println("error timeout!!!", enc.ConvertString(s.File))
		//	cancel()
		//	os.Exit(0)
		case <-s.Ctx.Done():
			fmt.Println("error timeout!!!", Enc.ConvertString(s.File))
			s.Cancel()
		}
	}()
}

func (s *AutoTranS) Task() {
	return
}

func (s *AutoTranS) WriteInfo(str string) {
	return
}

func (s *AutoTranS) UploadFiles(filestr string) {
	files := strings.Split(filestr, ",")
	for _, file := range files {
		s.File = file
		fmt.Println("准备开始上传的文件:", Enc.ConvertString(s.File))
		s.Task()
	}
}

func (s *AutoTranS) LoadCookies() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		// 如果cookies临时文件不存在则直接跳过
		if _, _err := os.Stat(s.Cookie); os.IsNotExist(_err) {
			fmt.Println("error cookie file not here", Enc.ConvertString(s.File))
			return errors.New("cookie not exist" + s.Cookie)
		}

		// 如果存在则读取cookies的数据
		cookiesData, err := ioutil.ReadFile(s.Cookie)
		if err != nil {
			fmt.Println("error cookie read file")
			return err
		}

		// 反序列化
		cookiesParams := network.SetCookiesParams{}
		if err = cookiesParams.UnmarshalJSON(cookiesData); err != nil {
			fmt.Println("error cookie json unmarshl")
			return err
		}

		// 设置cookies
		return network.SetCookies(cookiesParams.Cookies).Do(ctx)
	}
}

func (s *AutoTranS) LoadCookiesExistRemove() error {
	if _, _err := os.Stat(s.Cookie); !os.IsNotExist(_err) {
		err := os.Remove(s.Cookie)
		if err != nil {
			fmt.Println("error deleting cookie")
			return errors.New("err:error deleting cookie")
		}
		fmt.Println("info delete expired cookies...", Enc.ConvertString(s.Cookie))
	} else {
		fmt.Println("info the cookie does not exist...")
	}
	return nil
}

func (s *AutoTranS) LoginContext(boo bool) (context.Context, context.CancelFunc) {
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", boo), // 是否打开浏览器调试
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.103 Safari/537.36"), // 设置User-Agent
	}
	options = append(chromedp.DefaultExecAllocatorOptions[:], options...)
	// chromdp依赖context上限传递参数
	ctx, cancel := chromedp.NewExecAllocator(
		context.Background(),
		// 以默认配置的数组为基础，覆写headless参数
		// 当然也可以根据自己的需要进行修改，这个flag是浏览器的设置
		options...,
	)
	//defer cancel()
	// 创建新的chromedp上下文对象，超时时间的设置不分先后
	//ctx, cancel = context.WithTimeout(ctx, time.Second*5)
	//ctx, _ = context.WithTimeout(ctx, 30*time.Hour)
	ctx, cancel = chromedp.NewContext(
		ctx,
		// 设置日志方法
		chromedp.WithLogf(log.Printf),
	)
	// todo 这儿的超时退出机制需要完善
	return ctx, cancel
}

func (s *AutoTranS) SaveCookie() chromedp.ActionFunc {
	fmt.Println("====auto save cookie")
	return nil
}

func (s *AutoTranS) Loginx() error {
	// 加载cookie
	fmt.Println("--------------auto login")
	return nil
}

func NewAutoTrans(boo bool) AutoTranI {
	return &AutoTranS{}
}

type AutoTranS struct {
	Url     string
	Cookie  string
	File    string
	CsvFile string
	Ctx     context.Context
	Cancel  context.CancelFunc
}
