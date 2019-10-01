package main

import (
	"fmt"
	"github.com/tinyhole/pblib/format"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type IDLFile struct {
	Files []string `yaml:"protos"`
	Rel string `yaml:"rel"`
}

func main() {
	cfgFile, err := ioutil.ReadFile("idl.yaml")
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
	importPath := os.Getenv("THIDL")
	//fmt.Println(importPath)
	importPath = strings.TrimRight(importPath, "/")
	importPath, err = filepath.Abs(importPath)
	if err != nil || len(importPath) == 0{
		fmt.Println("THIDL is not abs")
		os.Exit(1)
	}

	//fmt.Println("importPath: ", importPath)
	pwd, _ := os.Getwd()
	pwd, _ = filepath.Abs(pwd)

	idlFile := IDLFile{}

	err = yaml.Unmarshal(cfgFile, &idlFile)
	if err != nil {
		fmt.Println("yaml file err")
		os.Exit(1)
	}

	outputDir := filepath.Join(pwd, "idl")
	err = os.RemoveAll(outputDir)
	if err != nil {
		fmt.Println("Remove dir failed: ", outputDir)
		os.Exit(1)
	}
	err = os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		fmt.Println("Mkdir failed: ", outputDir)
		os.Exit(1)
	}

	var descSource format.DescriptorSource

	importPaths := []string{importPath}
	descSource, err = format.DescriptorSourceFromProtoFiles(importPaths, idlFile.Files...)
	if err != nil {
		fmt.Println(err, "Failed to process proto source files.")
		os.Exit(1)
	}

	for name, fileDesc := range descSource.GetFiles() {
		path := filepath.Join(importPath, name)
		//fmt.Println("path:", path)

		fs := fileDesc.GetDependencies()
		var ms []string
		for _, fd := range fs {

			opt := fd.GetFileOptions().GoPackage
			if opt == nil {
				str := fd.GetPackage()
				opt = &str
			}

			m := fmt.Sprintf("M%s=%s/idl/%s", fd.GetName(), idlFile.Rel, *opt)
			ms = append(ms, m)
		}

		M := strings.Join(ms, ",")

		var args []string
		if len(fileDesc.GetServices()) == 0 {
			args = []string{
				"-I" + importPath,
				//"-I" + srcPath,
				"--go_out=" + M + ":" + outputDir,
				path,
			}
		} else {
			args = []string{
				"-I" + importPath,
				//"-I" + srcPath,
				"--go_out=" + M + ":" + outputDir,
				"--micro_out=" + M + ":" + outputDir,
				path,
			}
		}

		cmd := exec.Command("protoc", args...)
		//fmt.Println("cmd: ", cmd.Args)
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(cmd.Args)
			fmt.Println("Error:", err)
			fmt.Println(string(out))
		}

	}

}

