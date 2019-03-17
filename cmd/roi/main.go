package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gorilla/securecookie"

	"github.com/studio2l/roi"
)

// dev는 현재 개발모드인지를 나타낸다.
var dev bool

func main() {
	dev = true

	var (
		init  bool
		https string
		cert  string
		key   string
	)
	flag.BoolVar(&init, "init", false, "setup roi.")
	flag.StringVar(&https, "https", ":443", "address to open https port. it doesn't offer http for security reason.")
	flag.StringVar(&cert, "cert", "cert/cert.pem", "https cert file. default one for testing will created by -init.")
	flag.StringVar(&key, "key", "cert/key.pem", "https key file. default one for testing will created by -init.")
	flag.Parse()

	hashFile := "cert/cookie.hash"
	blockFile := "cert/cookie.block"

	if init {
		// 기본 Self Signed Certificate는 항상 정해진 위치에 생성되어야 한다.
		cert := "cert/cert.pem"
		key := "cert/key.pem"
		// 해당 위치에 이미 파일이 생성되어 있다면 건너 뛴다.
		// 사용자가 직접 추가한 인증서 파일을 덮어쓰는 위험을 없애기 위함이다.
		exist, err := anyFileExist(cert, key)
		if err != nil {
			log.Fatalf("error checking a certificate file %s: %s", cert, err)
		}
		if exist {
			log.Print("already have certificate file. will not create.")
		} else {
			// cert와 key가 없다. 인증서 생성.
			c := exec.Command("sh", "generate-self-signed-cert.sh")
			c.Dir = "cert"
			_, err := c.CombinedOutput()
			if err != nil {
				log.Fatal("error generating certificate files: ", err)
			}
		}

		exist, err = anyFileExist(hashFile, blockFile)
		if err != nil {
			log.Fatalf("could not check cookie key file: %s", err)
		}
		if exist {
			log.Print("already have cookie file. will not create.")
		} else {
			ioutil.WriteFile(hashFile, securecookie.GenerateRandomKey(64), 0600)
			ioutil.WriteFile(blockFile, securecookie.GenerateRandomKey(32), 0600)
		}
		return
	}

	err := roi.InitDB()
	if err != nil {
		log.Fatalf("could not initialize database: %v", err)
	}

	parseTemplate()

	hashKey, err := ioutil.ReadFile(hashFile)
	if err != nil {
		log.Fatalf("could not read cookie hash key from file '%s'", hashFile)
	}
	blockKey, err := ioutil.ReadFile(blockFile)
	if err != nil {
		log.Fatalf("could not read cookie block key from file '%s'", blockFile)
	}
	cookieHandler = securecookie.New(
		hashKey,
		blockKey,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			// 정의되지 않은 페이지로의 이동을 차단
			http.Error(w, "page not found", http.StatusNotFound)
			return
		}
		rootHandler(w, r)
	})
	mux.HandleFunc("/login/", loginHandler)
	mux.HandleFunc("/logout/", logoutHandler)
	mux.HandleFunc("/settings/profile", profileHandler)
	mux.HandleFunc("/update-password", updatePasswordHandler)
	mux.HandleFunc("/signup", signupHandler)
	mux.HandleFunc("/projects", projectsHandler)
	mux.HandleFunc("/add-project", addProjectHandler)
	mux.HandleFunc("/update-project", updateProjectHandler)
	mux.HandleFunc("/search/", searchHandler)
	mux.HandleFunc("/shot/", shotHandler)
	mux.HandleFunc("/add-shot/", addShotHandler)
	mux.HandleFunc("/update-shot", updateShotHandler)
	mux.HandleFunc("/update-task", updateTaskHandler)
	mux.HandleFunc("/version/", versionHandler)
	mux.HandleFunc("/add-version", addVersionHandler)
	mux.HandleFunc("/update-version", updateVersionHandler)
	mux.HandleFunc("/api/v1/shot/add", addShotApiHandler)
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	thumbfs := http.FileServer(http.Dir("roi-userdata/thumbnail"))
	mux.Handle("/thumbnail/", http.StripPrefix("/thumbnail/", thumbfs))

	// Show https binding information
	addrToShow := "https://"
	addrs := strings.Split(https, ":")
	if len(addrs) == 2 {
		if addrs[0] == "" {
			addrToShow += "localhost"
		} else {
			addrToShow += addrs[0]
		}
		if addrs[1] != "443" {
			addrToShow += ":" + addrs[1]
		}
	}
	fmt.Println()
	log.Printf("roi is start to running. see %s", addrToShow)
	fmt.Println()

	// Bind
	log.Fatal(http.ListenAndServeTLS(https, cert, key, mux))
}
