package common

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

//获取指定目录下的所有文件和目录
// 文件遍历
func ListDir(dirPth string) (files []string) {
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil
	}
	PthSep := string(os.PathSeparator)
	// suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			//files1 = append(files1, dirPth+PthSep+fi.Name())
			//ListDir(dirPth + PthSep + fi.Name())
			//fmt.Println(dirPth + PthSep + fi.Name())
		} else {
			//fmt.Println("s")
			fileP := dirPth + PthSep + fi.Name()
			if filepath.Ext(fileP) == ".mp4" {
				files = append(files, fileP)
			}
		}
	}
	return files
}
