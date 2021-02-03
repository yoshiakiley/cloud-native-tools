package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	utils "github.com/yametech/cloud-native-tools/pkg/utils"
)

func main() {
	var url string
	var codeType string
	var projectPath string
	var unitTest bool

	flag.StringVar(&url, "url", "./Dockerfile", "-url ./")
	flag.StringVar(&codeType, "codetype", "java-maven", "-codetype java-maven")
	flag.StringVar(&projectPath, "path", "", "-path subdirectory")
	flag.BoolVar(&unitTest, "unittest", false, "-unittest True")
	flag.Parse()

	fmt.Printf("url=%v  codeType=%v\n", url, codeType)
	err := CheckDockerFile(url, codeType, projectPath, unitTest)
	if err != nil {
		panic(err)
	}
}

func CheckDockerFile(url string, codeType string, projectPath string, unitTest bool) error {
	count := strings.Index(url, "Dockerfile")
	if count == -1 {
		url = path.Join(url, "Dockerfile")
	}
	switch codeType {
	case "django":
		err := djangoDocker(url)
		if err != nil {
			return err
		}
	case "java-maven":
		err := javaDocker(url, projectPath, unitTest)
		if err != nil {
			return err
		}
	case "easyswoole":
		err := easyswooleDocker(url)
		if err != nil {
			return err
		}
	case "web":
		err := webDocker(url)
		if err != nil {
			return err
		}
	}

	return nil
}

func webDocker(url string) error {
	if _, err := os.Stat(url); os.IsNotExist(err) {
		err = utils.GenerateFile(url, webContent)
		if err != nil {
			return err
		}
	}

	ngFile := strings.Replace(url, "Dockerfile", "nginx.conf", -1)
	if _, err := os.Stat(ngFile); os.IsNotExist(err) {
		err = utils.GenerateFile(ngFile, ngContent)
		if err != nil {
			return err
		}
	}
	return nil
}

func djangoDocker(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err = utils.GenerateFile(filename, djangoDockerFileContent)
		if err != nil {
			return err
		}
	}
	return nil
}

func javaDocker(filename string, projectPath string, unitTest bool) error {
	type Param struct {
		ProjectPath string
		SkipTest   	string
	}

	param := &Param{ProjectPath: "*",SkipTest:""}
	if len(strings.Trim(projectPath, "")) > 0 {
		param.ProjectPath = projectPath
	}
	if unitTest{
		param.SkipTest = "-Dmaven.test.skip=true"
	}

	content, err := Render(param, javaMavenDockerfileContentTpl)
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err = utils.GenerateFile(filename, content)
		if err != nil {
			return err
		}
	}
	// if settings.xml not exists, create it too
	xmlFile := strings.Replace(filename, "Dockerfile", "settings.xml", -1)
	if _, err := os.Stat(xmlFile); os.IsNotExist(err) {
		err = utils.GenerateFile(xmlFile, javaXmlContent)
		if err != nil {
			return err
		}
	}

	if unitTest {
		unitList := strings.Split(filename, "/")
		unitList = unitList[:len(unitList)-1]
		unitDockerfile := strings.Join(unitList, "/")
		unitDockerfile = fmt.Sprintf("%s/%s", unitDockerfile, "Dockerfile-unittest")

		err = utils.GenerateFile(unitDockerfile, javaMavenUnitContent)
		if err != nil {
			return err
		}
	}

	return nil
}

func easyswooleDocker(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		err = utils.GenerateFile(filename, easySwooleContent)
		if err != nil {
			return err
		}
	}
	return nil
}
