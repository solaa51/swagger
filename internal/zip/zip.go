package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ZipDir 将文件夹压缩为dst文件
func ZipDir(dir, dst string) error {
	//创建目标文件
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	zipWriter := zip.NewWriter(dstFile)
	defer zipWriter.Close()

	//srcDir := dir + string(filepath.Separator) + "Payload"
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if path == dir {
			return nil
		}

		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".DS_Store") { //mac的临时文件 跳过
			return nil
		}

		//通过文件信息，创建压缩文件的 文件信息
		zInfo, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		zInfo.Name = path[len(dir):]
		/*if runtime.GOOS == "windows" {
			zInfo.Name = strings.ReplaceAll(zInfo.Name, "\\", "/")
		}*/
		if info.IsDir() {
			zInfo.Name += "/" //zip文件规定 文件内必须使用 /
		} else {
			zInfo.Method = zip.Deflate
		}

		//将文件头信息写入到压缩文件中，可以理解为 创建文件了
		zIW, err := zipWriter.CreateHeader(zInfo)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			pFData, err := os.Open(path)
			if err != nil {
				fmt.Println(err)
			}
			defer pFData.Close()

			_, err = io.Copy(zIW, pFData)
			if err != nil {
				return err
			}
		}

		//fmt.Println("成功压缩：", zInfo.Name)

		return nil
	})

	return err
}

// UnZip 解压缩zip文件 zipFile 待解压的文件 dir保存目录
func UnZip(zipFile, dir string) error {
	zr, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zr.Close()

	if dir != "" {
		//判断并创建多层文件夹
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
	}

	for _, file := range zr.File {
		path := filepath.Join(dir, file.Name)
		//如果是目录 则 直接创建 并跳过
		if file.FileInfo().IsDir() {
			if err = os.MkdirAll(path, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		tfPath := filepath.Dir(path)
		if err = os.MkdirAll(tfPath, os.ModePerm); err != nil {
			return err
		}

		//获得待解压的 内部文件的句柄
		fs, err := file.Open()
		if err != nil {
			return err
		}

		//创建并获得目的文件句柄
		fd, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
		if err != nil {
			fs.Close()
			return err
		}

		//将内容写入
		_, err = io.Copy(fd, fs)
		if err != nil {
			fd.Close()
			fs.Close()
			return err
		}

		fd.Close()
		fs.Close()
	}

	return nil
}
