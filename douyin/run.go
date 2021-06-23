package douyin

// 抖音登录
func LoginDY() {
	dy := NewDY(false)
	_ = dy.Loginx()
}

// 抖音视频上传
func UploadDY(files string) {
	dy := NewDY(true)
	dy.UploadFiles(files)
}
