package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// CopyAndReplaceTemplate 拷贝并且替换指定的东西
// @param src 代表的是源文件
// @param dest 代表的是目标文件
// @param replacements 代表的是替换
func CopyAndReplaceTemplate(src, dest string, replacements map[string]string) error {
	content, err := ioutil.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read template: %v", err)
	}

	// Replace placeholders
	text := string(content)
	for placeholder, value := range replacements {
		text = strings.ReplaceAll(text, placeholder, value)
	}

	// Write to destination
	err = ioutil.WriteFile(dest, []byte(text), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %v", dest, err)
	}
	return nil
}
