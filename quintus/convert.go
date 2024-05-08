package quintus

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

// Convert 从 template 转换到真实数据
func Convert(data *Data, srcPath string, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}

	dstPath = PathConvert(data, dstPath)
	dstDir := filepath.Dir(dstPath)

	_, err = os.Stat(dstDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dstDir, 0755)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	return ConvertIO(data, srcFile, dstFile)
}

// ConvertIO 从 template 转换到真实数据
func ConvertIO(data *Data, src io.Reader, dst io.Writer) error {
	srcTxt, err := ioutil.ReadAll(src)
	if err != nil {
		return err
	}

	tmpl, err := template.New("quintus").Parse(string(srcTxt))
	if err != nil {
		return err
	}

	err = tmpl.Execute(dst, data)
	if err != nil {
		return err
	}

	return nil
}

// PathConvert 转换文件名
func PathConvert(data *Data, path string) string {
	tmpl, err := template.New("qunitus").Parse(path)
	if err != nil {
		return path
	}

	w := &bytes.Buffer{}
	err = tmpl.Execute(w, data)
	if err != nil {
		return path
	}

	return w.String()
}
