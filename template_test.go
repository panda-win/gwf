package gwf

import (
	"os"
	"testing"
)

func TestListAllFileInDir(t *testing.T) {
	dir, _ := os.Getwd()
	t.Logf("cur dir is: %s", dir)
	filepaths, err := listAllFileInDir(dir, ".go")
	if err != nil {
		t.Errorf("err: %v", err)
	}
	for _, p := range filepaths {
		t.Logf("filepath : %s", p)
	}
}
