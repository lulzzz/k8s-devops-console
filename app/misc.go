package app

import (
	"os"
	"path/filepath"
	"k8s.io/apimachinery/pkg/runtime"
	"io/ioutil"
	"k8s.io/client-go/kubernetes/scheme"
)

func IsDirectory(path string) (bool) {
	fileInfo, err := os.Stat(path)

	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}

func IsRegularFile(path string) (bool) {
	fileInfo, _ := os.Stat(path)
	return fileInfo.Mode().IsRegular()
}

func IsK8sConfigFile(path string) (bool) {
	if !IsRegularFile(path) {
		return false
	}

	switch(filepath.Ext(path)) {
	case ".json":
		return true
	case ".yaml":
		return true
	}

	return false
}

func recursiveFileListByPath(path string) (list []string) {
	filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if IsK8sConfigFile(path) {
			list = append(list, path)
		}
		return nil
	})

	return
}

func KubeParseConfig(path string) (runtime.Object) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, nil)
	if err != nil {
		panic(err)
	}
	return obj
}
