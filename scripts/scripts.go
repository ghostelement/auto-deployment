package scripts

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed scripts/*
var scripts embed.FS

// 拷贝scripts目录内脚本文件
func CopyShellScriptToWorkingDir(destPath string) error {
	files, err := scripts.ReadDir("scripts")
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := file.Name()
		fmt.Println("Write scripts: ", filename)
		data, err := scripts.ReadFile("scripts/" + filename)
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(destPath, filename), data, 0755)
		if err != nil {
			return err
		}

	}

	return err
}
