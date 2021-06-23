package douyin

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"gitee/zzf/chromedp-demo/common"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/unknwon/com"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (s *DYSturct) WriteInfo(str string) {
	file, err := os.OpenFile(s.CsvFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, fs.ModePerm)
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()
	writer := csv.NewWriter(file)
	// https://bj.bcebos.com/v1/ai-studio-online/7cd09ae961a54678ad4b784642192b40ca657f0368724552934384571&ab3234? responseContentDisposition=attachment%3B%20filename%3D文件名.mp4
	err = writer.Write([]string{filepath.Base(s.File), str})
	if err != nil {
		fmt.Println("error write csv douyin ...", common.Enc.ConvertString(filepath.Base(s.File)), common.Enc.ConvertString(str))
		return
	}
	writer.Flush()
}
func (s *DYSturct) Task() {
	// 加载cookie
	if err := chromedp.Run(s.Ctx, s.LoadCookies()); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	// 打开网页
	fmt.Println("info open url ...", common.Enc.ConvertString(s.Url))
	if err := chromedp.Run(s.Ctx, chromedp.Navigate(s.Url)); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	//fmt.Println("开始监听...")
	downloadComplete := make(chan bool, 1)
	var requestId network.RequestID
	//requestId := make([]network.RequestID, 0)
	chromedp.ListenTarget(s.Ctx, func(v interface{}) {
		switch ev := v.(type) {
		//https://creator.douyin.com/web/api/media/video/transend/?video_id=
		case *network.EventRequestWillBeSent: // 开始发送
			//marshal, _ := json.Marshal(v)
			if strings.Contains(ev.Request.URL, "https://creator.douyin.com/web/api/media/video/transend/?video_id=") {
				if ev.HasUserGesture {
					//fmt.Println("监听开始：", ev.HasUserGesture, string(marshal))
					requestId = ev.RequestID
					//requestId = append(requestId, ev.RequestID)
				}
			}
		case *network.EventResponseReceived: // 返回的资源
			//marshal, _ := json.Marshal(v)
			//if strings.Contains(ev.Response.URL, "https://creator.douyin.com/web/api/media/video/transend/?video_id=") {
			//	if _, ok := ev.Response.Headers["content-encoding"]; ok {
			//fmt.Println("监听中：", string(marshal))
			//		requestId = append(requestId, ev.RequestID)
			//	}
			//}
			//resp := ev.Response
			//if len(resp.Headers) != 0 {
			//	log.Printf("received headers: %s", resp.Headers)
			//}
		case *network.EventLoadingFinished: // 资源加载完毕
			if ev.RequestID == requestId {
				//marshal, _ := json.Marshal(v)
				//fmt.Println("结束发送...文件...", string(marshal))
				downloadComplete <- true
				//close(downloadComplete)
			}
		}
	})
	// 上传视频
	fmt.Println("info video upload:", common.Enc.ConvertString(s.File))
	//if err := chromedp.Run(s.Ctx, chromedp.Click(`document.querySelector("#root > div > section > section > main > div > div > div.content-body--1XCPO > div.preview--3koIV > div > div > div > div > div > div > div > div > div")`, chromedp.ByJSPath)); err != nil {
	//	fmt.Println("error",s.File, err)
	//	return
	//}
	now := time.Now()
	if err := chromedp.Run(s.Ctx, chromedp.SetUploadFiles(`document.querySelector("#root > div > section > section > main > div > div > div.content-body--1XCPO > div.preview--3koIV > div > div > div > div > div > div > div > div > input")`,
		[]string{s.File}, chromedp.ByJSPath)); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	// 等待视频传输完成
	fmt.Println("info waiting for video upload ...")

	if err := chromedp.Run(s.Ctx, chromedp.WaitVisible(`#root > div > section > section > main > div > div > div.content-body--1XCPO > div.preview--3koIV > div > div > div > div > div.long-card--Y_Q6H > div:nth-child(2) > label > svg`,
		chromedp.ByID)); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	// 写视频描述
	fmt.Println("info fill in the video description ...")
	if err := chromedp.Run(s.Ctx, chromedp.SendKeys(`#root > div > section > section > main > div > div > div.content-body--1XCPO > div.form--2xPFu > div:nth-child(2) > div.editor--3w64r.editor--xKx-9 > div.DraftEditor-root > div.DraftEditor-editorContainer > div`,
		time.Now().Format(time.RFC3339), chromedp.ByID)); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	// 等待文件上传完成
	//time.Sleep(time.Second*5)
	if <-downloadComplete {
		fmt.Println("info video sent successfully:", common.Enc.ConvertString(s.File))
	} else {
		fmt.Println("error video sent fail:", common.Enc.ConvertString(s.File))
		return
	}
	//fmt.Println("文件上传结果:", !<-downloadComplete)
	fmt.Println("info video upload time:", time.Now().Sub(now).Seconds(), "s")
	var downloadBytes []byte
	//downloadBytes := make([][]byte, 0)
	if err := chromedp.Run(s.Ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		//for _, v := range requestId {
		//	d, _ := network.GetResponseBody(v).Do(ctx)
		//	fmt.Println(v.String(), string(d))
		//	downloadBytes = append(downloadBytes, d)
		//}
		downloadBytes, err = network.GetResponseBody(requestId).Do(ctx)
		//fmt.Println("--dow--", string(downloadBytes))
		return err
	})); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	var t DouyinDataStruct
	err := json.Unmarshal(downloadBytes, &t)
	if err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	splitN := strings.SplitN(t.PlayURL, "?", 2)
	fmt.Println("info results obtained:", common.Enc.ConvertString(splitN[0]))
	// 发布
	fmt.Println("info start publishing ...")
	if err := chromedp.Run(s.Ctx, chromedp.Click(`document.querySelector("#root > div > section > section > main > div > div > div.content-body--1XCPO > div.form--2xPFu > div.content-confirm-container--2AI6I > button.button--1SZwR.primary--1AMXd.fixed--3rEwh")`,
		chromedp.ByJSPath)); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	// 等待页面出来西瓜视频是否绑定的通知
	//fmt.Println("等待页面出来西瓜视频是否绑定的通知...")
	fmt.Println("info deal with post release issues ...")
	if err := chromedp.Run(s.Ctx, chromedp.WaitVisible(`#dialog-0 > div > div.semi-modal-body > div > div.modal-header--3DK-t > span.title--302MY`,
		chromedp.ByID)); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	// 不绑定西瓜帐号
	//fmt.Println("不绑定西瓜帐号...")
	if err := chromedp.Run(s.Ctx, chromedp.Click(`#dialog-0 > div > div.semi-modal-footer > div > button.semi-button.semi-button-tertiary.semi-button-light > span`,
		chromedp.ByID)); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return
	}
	fmt.Println("info file uploaded successfully ...", common.Enc.ConvertString(s.File))
	s.WriteInfo(splitN[0])
}

func (s *DYSturct) UploadFiles(filestr string) {
	files := strings.Split(filestr, ",")
	for _, file := range files {
		if file == "" {
			fmt.Println("error file name, next file ...local filename:", common.Enc.ConvertString(file))
			continue
		}
		if !com.IsExist(file) { // 文件不存在
			fmt.Println("error file not exist, next file ... local filename:", common.Enc.ConvertString(file))
			continue
		}
		s.File = file
		fmt.Println("info files ready to start uploading:", common.Enc.ConvertString(s.File))
		s.Task()
	}
}

func (s *DYSturct) SaveCookie() chromedp.ActionFunc {
	return func(ctx context.Context) (err error) {
		// 等待二维码登陆
		fmt.Println("info please log in to douyin ...")
		time.Sleep(time.Second * 2)
		fmt.Println("info please move the mouse to the user's position in the upper right corner of the page ...")
		if err = chromedp.WaitVisible(`#root > div > div.content--3ncCf.creator-content > div.sider--137Dm > div > div > div > div.semi-navigation-header > button > span`, chromedp.ByID).Do(ctx); err != nil {
			fmt.Println("error", common.Enc.ConvertString(s.File), err)
			return err
		} // #main > header > div > div.studio-logged-wrapper > div > div > span
		fmt.Println("info login successful, save cookie ...")
		// cookies的获取对应是在devTools的network面板中
		// 1. 获取cookies
		cookies, err := network.GetAllCookies().Do(ctx)
		if err != nil {
			return err
		}

		// 2. 序列化
		cookiesData, err := network.GetAllCookiesReturns{Cookies: cookies}.MarshalJSON()
		if err != nil {
			return err
		}

		// 3. 存储到临时文件
		if err = ioutil.WriteFile(s.Cookie, cookiesData, 0755); err != nil {
			return err
		}
		fmt.Println("info cookie saved successfully ...")
		return
	}
}

func (s *DYSturct) Loginx() error {
	defer s.Cancel()
	fmt.Println("info login douyin ...")
	fmt.Println("info cookie remove ...")
	// 加载cookie
	if err := s.LoadCookiesExistRemove(); err != nil {
		return err
	}
	// 打开网页
	fmt.Println("info open chrom ...")
	if err := chromedp.Run(s.Ctx, chromedp.Navigate(s.Url)); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return err
	}
	s.WaitTimes(60)
	if err := chromedp.Run(s.Ctx, s.SaveCookie()); err != nil {
		fmt.Println("error", common.Enc.ConvertString(s.File), err)
		return err
	}
	return nil
}

func NewDY(boo bool) common.AutoTranI {
	s := common.AutoTranS{
		Url:     "https://creator.douyin.com/content/publish",
		Cookie:  "douyin.tmp",
		CsvFile: "douyin.csv",
	}
	s.Ctx, s.Cancel = s.LoginContext(boo)

	return &DYSturct{
		s,
	}
}

type DYSturct struct {
	common.AutoTranS
}

/**
{
    "codec_type":"h264",
    "duration":6.247,
    "encode":1,
    "extra":{
        "logid":"202105271946530102111790695A0D8F8D",
        "now":1622116013000
    },
    "format":"mp4",
    "height":"960",
    "play_url":"http://v83.douyinvod.com/6a2a45bee16dc746b5354fdfdbe4a4cf/60af94c3/video/tos/cn/tos-cn-v-0015c002/8903cf5939db4c9898883443a3c23c07/?a=1128\u0026br=1854\u0026bt=1854\u0026btag=4\u0026cd=0%7C0%7C0\u0026ch=0\u0026cr=0\u0026cs=0\u0026dr=0\u0026ds=2\u0026er=\u0026l=202105271946530102111790695A0D8F8D\u0026lr=\u0026mime_type=video_mp4\u0026net=0\u0026pl=0\u0026qs=13\u0026rc=Mzg8NjU8dGRyNTMzNGkzM0ApaXN4M2Q6OWdxZjMzajM2eWcpcXdyMXJkdTFseGdmY21yLV40M2FrYC0tZC0wc3MtMjYwLzZuXmwvLS0uYy0tOmNxYmtgb3FiYGt2bDo%3D\u0026vl=\u0026vr=",
    "poster_url":"http://p3-tt.bytecdn.cn/obj/tos-cn-p-0015/b4368edeebf2443ba60f8f5612b7e04b",
    "status_code":0,
    "width":"432"
}
*/
type DouyinDataStruct struct {
	CodecType  string           `json:"codec_type"`
	Duration   float64          `json:"duration"`
	Encode     int              `json:"encode"`
	Extra      DouyinDataResult `json:"extra"`
	Format     string           `json:"format"`
	Height     string           `json:"height"`
	PlayURL    string           `json:"play_url"`
	PosterURL  string           `json:"poster_url"`
	StatusCode int              `json:"status_code"`
	Width      string           `json:"width"`
}

type DouyinDataResult struct {
	Logid string `json:"logid"`
	Now   int64  `json:"now"`
}
