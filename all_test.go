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
	// DB 시작
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
	go func() {
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// DB가 시작되기까지 시간 필요
	time.Sleep(time.Second)

	// 테스트
	code := m.Run()

	// DB 종료, 삭제
	cmd.Process.Kill()
	os.RemoveAll(tempDir)

	os.Exit(code)
}
