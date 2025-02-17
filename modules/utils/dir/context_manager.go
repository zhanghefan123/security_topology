package dir

import (
	"fmt"
	"os"
)

type ContextManager struct {
	OldDirectory string
	NewDirectory string
}

func (cm *ContextManager) Enter(directory string) error {
	fmt.Printf("enter directory: %s\n", directory)
	if err := os.Chdir(directory); err != nil {
		return fmt.Errorf("change to new dir failed: %v", err)
	} else {
		cm.NewDirectory = directory
		cm.OldDirectory, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("get current dir failed: %v", err)
		}
		return nil
	}
}

func (cm *ContextManager) Exit() error {
	fmt.Printf("exit dir: %s\n", cm.NewDirectory)
	if err := os.Chdir(cm.OldDirectory); err != nil {
		return fmt.Errorf("change back to old dir failed: %v", err)
	} else {
		return nil
	}
}

// WithContextManager 进入指定目录, 执行一段指令之后回到之前的目录
func WithContextManager(directory string, fn func() error) (err error) {
	cm := &ContextManager{}
	err = cm.Enter(directory)
	if err != nil {
		return fmt.Errorf("context manager enter directory failed: %v", err)
	}
	defer func() {
		leaveErr := cm.Exit()
		if err == nil {
			err = leaveErr
		}
	}()
	err = fn()
	return err
}
