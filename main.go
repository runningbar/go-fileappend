package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

var startPoint = `./test`

//var startPoint = "C:/mine/GoWorkspace/src/github.com/runningbar/go-fileappend"
var except = []string{"handled", "out"}
var flags = []string{"rmvb", "avi", "mp4", "wmv"}
var source = except[0]
var dest = except[1]

/*读取sp路径下的所有文件*/
func listFiles(sp string, except []string, fileNames []string) []string {
	files, _ := ioutil.ReadDir(sp)
	for _, file := range files {
		var skip = false
		if sp == startPoint {
			for _, e := range except {
				if file.IsDir() && file.Name() == e {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}
		if !file.IsDir() {
			fileNames = append(fileNames, sp+"/"+file.Name())
		} else {
			fileNames = listFiles(sp+"/"+file.Name(), except, fileNames)
		}
	}
	return fileNames
}

/*过滤出指定的文件*/
func filterFiles(fileNames []string, flags []string) []string {
	var r []string
	for _, fileName := range fileNames {
		for _, flag := range flags {
			if strings.HasSuffix(fileName, flag) {
				r = append(r, fileName)
				break
			}
		}
	}
	return r
}

func copyFiles(fileNames []string, sp string, dest string) {
	os.MkdirAll(sp+"/"+dest, 1777)
	var length = len(fileNames)
	for i, fileName := range fileNames {
		fmt.Println("file copying: ", i, " / ", length)
		var fn = fileName[strings.LastIndex(fileName, "/")+1:]
		fs, _ := os.Open(fileName)
		fd, _ := os.Create(sp + "/" + dest + "/" + fn)
		io.Copy(fd, fs)
		fs.Close()
		fd.Close()
		fileNames[i] = sp + "/" + dest + "/" + fn
		os.Remove(fileName)
	}
	fmt.Println("file copy complete")
}

/*提炼出文件名*/
func abstractNames(fileNames []string) []string {
	var re = regexp.MustCompile(`[a-zA-Z]+-\d+`)
	var names []string
	for _, fileName := range fileNames {
		fn := fileName[strings.LastIndex(fileName, "/")+1:]
		ff := fn[strings.LastIndex(fn, `.`)+1:]
		an := re.FindString(fn)
		if an == "" {
			an = fn[:strings.LastIndex(fn, `.`)]
		}
		names = append(names, an+`.`+ff)
	}
	return names
}

func createUniqueNames(names []string) []string {
	var newNames []string
	for _, name := range names {
		dotPos := strings.LastIndex(name, `.`)
		n := name[:dotPos]
		f := name[dotPos+1:]
		s1 := fmt.Sprintf("%x", sha256.Sum256([]byte(n)))
		s2 := fmt.Sprintf("%x", sha256.Sum256([]byte(f)))
		s3 := fmt.Sprintf("%x", sha256.Sum256([]byte(name)))
		newName := fmt.Sprintf("%x", sha256.Sum256([]byte(s1+s2+s3))) + `.` + f
		newNames = append(newNames, newName)
	}
	return newNames
}

func appendFile(fileNames []string, newNames []string, sp string, dest string) {
	os.MkdirAll(sp+"/"+dest, 1777)
	var length = len(fileNames)
	for i, fileName := range fileNames {
		fmt.Println("file appending: ", i, " / ", length)
		fs, _ := os.Open(fileName)
		fd, _ := os.Create(sp + "/" + dest + "/" + newNames[i])
		io.Copy(fd, fs)
		fs.Close()
		fd.Close()
		fd, _ = os.OpenFile(sp+"/"+dest+"/"+newNames[i], os.O_WRONLY, 0666)
		fd.Seek(0, os.SEEK_END)
		fd.WriteString(newNames[i][:strings.LastIndex(newNames[i], `.`)])
		fd.Close()

		n := fileName[:strings.LastIndex(fileName, `.`)]
		f := fileName[strings.LastIndex(fileName, `.`)+1:]
		successName := n + "_OK." + f
		os.Rename(fileName, successName)
	}
	fmt.Println("file append complete")
}

func getFileSearchKey(fileName string) string {
	re := regexp.MustCompile(`[a-zA-Z]+-\d+`)
	fileFormat := fileName[strings.LastIndex(fileName, `.`)+1:]
	abstractName := re.FindString(fileName)
	if abstractName == "" {
		abstractName = fileName[:strings.LastIndex(fileName, `.`)]
	}
	name := abstractName + `.` + fileFormat

	s1 := fmt.Sprintf("%x", sha256.Sum256([]byte(abstractName)))
	s2 := fmt.Sprintf("%x", sha256.Sum256([]byte(fileFormat)))
	s3 := fmt.Sprintf("%x", sha256.Sum256([]byte(name)))
	searchKey := fmt.Sprintf("%x", sha256.Sum256([]byte(s1+s2+s3))) + `.` + fileFormat
	fmt.Println(searchKey)
	return searchKey
}

func main() {
	/*
		fileNames是源文件绝对路径全名
		names是提炼出来的文件名，不含父路径
		newNames是生成的目标文件名，不含父路径
	*/
	var fileNames []string
	fileNames = listFiles(startPoint, except, fileNames)
	//fmt.Println(fileNames)
	fileNames = filterFiles(fileNames, flags)
	//fmt.Println(fileNames)
	copyFiles(fileNames, startPoint, source)
	//fmt.Println(fileNames)
	names := abstractNames(fileNames)
	//fmt.Println(names)
	newNames := createUniqueNames(names)
	//fmt.Println(newNames)
	appendFile(fileNames, newNames, startPoint, dest)

	//getFileSearchKey(`SHKD664_OK.avi`)
}
