package roi

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
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
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	// 테스트가 죽었을 때 cockroach DB 및 임시 디렉토리 삭제
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	go func() {
		<-s
		cmd.Process.Kill()
		os.RemoveAll(tempDir)
		os.Exit(1)
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
