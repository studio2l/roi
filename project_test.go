package roi

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
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

func TestProject(t *testing.T) {
	want := &Project{
		ID:            "TEST",
		Name:          "테스트 프로젝트",
		Status:        "waiting",
		Client:        "레이지 픽처스",
		Director:      "윤지은",
		Producer:      "김한웅",
		VFXSupervisor: "김성환",
		VFXManager:    "조경식",
		CGSupervisor:  "김용빈",
		CrankIn:       time.Date(2018, 12, 31, 7, 30, 0, 0, time.Local).UTC(),
		CrankUp:       time.Date(2019, 8, 31, 19, 0, 0, 0, time.Local).UTC(),
		StartDate:     time.Date(2018, 12, 29, 0, 0, 0, 0, time.Local).UTC(),
		ReleaseDate:   time.Date(2018, 10, 1, 0, 0, 0, 0, time.Local).UTC(),
		VFXDueDate:    time.Date(2018, 9, 31, 0, 0, 0, 0, time.Local).UTC(),
		OutputSize:    "1920x1080",
		ViewLUT:       "some/place/aces.lut",
	}

	db, err := sql.Open("postgres", "postgresql://root@localhost:54545/roi?sslmode=disable")
	if err != nil {
		t.Fatalf("error connecting to the database: %s", err)
	}
	if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS roi"); err != nil {
		log.Fatal("error creating db 'roi': ", err)
	}
	err = CreateTableIfNotExists(db, "projects", ProjectTableFields)
	if err != nil {
		t.Fatalf("could not create projects table: %s", err)
	}
	err = AddProject(db, want)
	if err != nil {
		t.Fatalf("could not add project to projects table: %s", err)
	}
	got, err := GetProject(db, want.ID)
	if err != nil {
		t.Fatalf("could not get project to projects table: %s", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %v, want: %v", got, want)
	}
}
