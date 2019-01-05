package roi

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	tempDir, err := ioutil.TempDir("", "cockroach-test-")
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command(
		"cockroach", "start", "--insecure",
		"--http-addr=localhost:5454", "--listen-addr=localhost:54545",
		fmt.Sprintf("--store=path=%s", tempDir),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second)

	code := m.Run()

	cmd.Process.Kill()
	os.RemoveAll(tempDir)
	os.Exit(code)
}
