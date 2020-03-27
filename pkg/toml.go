package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
	"strings"
)

// LoadTomlConfig 读取一个TOML文件或者文件夹内所有TOML文件，返回一个Map对象。
func LoadTomlFile(dirOrFile string) (map[string]interface{}, error) {
	if "" == dirOrFile {
		return nil, errors.New("dir or file path is required")
	}

	fi, err := os.Stat(dirOrFile)
	if nil != err {
		return nil, errors.New("failed to get file/dir info")
	}

	var buffer []byte
	if fi.IsDir() {
		if bs, err := LoadTomlDirectoryText(dirOrFile); nil != err {
			return nil, err
		} else {
			buffer = bs
		}
	} else {
		if bs, err := ioutil.ReadFile(dirOrFile); nil != err {
			return nil, errors.New("failed to read .toml pkg file")
		} else {
			buffer = bs
		}
	}

	m := make(map[string]interface{})
	if _, err := toml.Decode(string(buffer), &m); nil != err {
		return nil, err
	} else {
		return m, nil
	}

}

// LoadTomlDirectoryText 加载指定TOML配置文件目录，返回所有配置文件的合并Text文本；
func LoadTomlDirectoryText(dirName string) ([]byte, error) {
	out := new(bytes.Buffer)
	if files, err := ioutil.ReadDir(dirName); nil != err {
		return nil, errors.New("failed to list file in dir: " + dirName)
	} else {
		if 0 == len(files) {
			return nil, errors.New("pkg file NOT FOUND in dir: " + dirName)
		}
		for _, f := range files {
			name := f.Name()
			if !strings.HasSuffix(name, ".toml") {
				continue
			}
			path := fmt.Sprintf("%s%s%s", dirName, "/", f.Name())
			if bs, err := ioutil.ReadFile(path); nil != err {
				return nil, errors.New("Failed to load file: %s" + path)
			} else {
				out.Write(bs)
				out.WriteByte('\n')
			}
		}
	}
	out.WriteByte('\n')
	return out.Bytes(), nil
}
