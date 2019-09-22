package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/securecookie"

	"github.com/studio2l/roi"
)

// dev는 현재 개발모드인지를 나타낸다.
var dev bool

func redirectToHttps(w http.ResponseWriter, r *http.Request) {
	to := "https://" + strings.Split(r.Host, ":")[0] + r.URL.Path
	if r.URL.RawQuery != "" {
		to += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, to, http.StatusTemporaryRedirect)
}

func main() {
	dev = true

	var (
		addr  string
		https bool
		cert  string
		key   string
	)
	flag.StringVar(&addr, "addr", "localhost:80:443", `site address and it's http/https port.
when two ports are specified, first port is for http and second is for https.
with -https flag turned on, it will automatically forward http port to https port.
when one port is specified, the port will be used for current protocol.
when no port is specified, default port for current protocol will be used.
`)
	flag.BoolVar(&https, "https", false, "use https port.")
	flag.StringVar(&cert, "cert", "cert/cert.pem", "https cert file. valid only if -https flag is turned on.")
	flag.StringVar(&key, "key", "cert/key.pem", "https key file. valid only if -https flag is turned on.")
	flag.Parse()

	hashFile := "cert/cookie.hash"
	blockFile := "cert/cookie.block"
	blockFileExist, err := anyFileExist(hashFile, blockFile)
	if err != nil {
		log.Fatalf("could not check cookie key file: %s", err)
	}
	if !blockFileExist {
		err = os.MkdirAll(filepath.Dir(hashFile), 0755)
		if err != nil && !os.IsExist(err) {
			log.Fatalf("could not create directory for cookie hash file: %s", err)
		}
		err = os.MkdirAll(filepath.Dir(blockFile), 0755)
		if err != nil && !os.IsExist(err) {
			log.Fatalf("could not create directory for cookie block file: %s", err)
		}
		ioutil.WriteFile(hashFile, securecookie.GenerateRandomKey(64), 0600)
		ioutil.WriteFile(blockFile, securecookie.GenerateRandomKey(32), 0600)
	}

	db, err := roi.InitDB()
	if err != nil {
		log.Fatalf("could not initialize database: %v", err)
	}
	exist, err := roi.UserExist(db, "admin")
	if err != nil {
		log.Fatalf("could not check admin user exist: %v", err)
	}
	if !exist {
		err := roi.AddUser(db, "admin", "password1!")
		if err != nil {
			log.Fatalf("could not create admin user: %v", err)
		}
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
	mux.HandleFunc("/add-shot/", addShotHandler)
	mux.HandleFunc("/update-shot", updateShotHandler)
	mux.HandleFunc("/update-task", updateTaskHandler)
	mux.HandleFunc("/version/", versionHandler)
	mux.HandleFunc("/add-version", addVersionHandler)
	mux.HandleFunc("/update-version", updateVersionHandler)
	mux.HandleFunc("/api/v1/project/add", addProjectApiHandler)
	mux.HandleFunc("/api/v1/shot/add", addShotApiHandler)
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	thumbfs := http.FileServer(http.Dir("roi-userdata/thumbnail"))
	mux.Handle("/thumbnail/", http.StripPrefix("/thumbnail/", thumbfs))

	// Show https binding information
	addrs := strings.Split(addr, ":")

	var protocol string
	site := ""
	httpPort := "80"
	httpsPort := "443"
	portForwarding := false
	if https {
		protocol = "https://"
		site = addrs[0]
		if len(addrs) == 3 {
			portForwarding = true
			httpPort = addrs[1]
			httpsPort = addrs[2]
		} else if len(addrs) == 2 {
			httpsPort = addrs[1]
		}
	} else {
		protocol = "http://"
		site = addrs[0]
		if len(addrs) == 2 {
			httpPort = addrs[1]
		}
	}
	if site == "" {
		site = "localhost"
	}

	addrToShow := protocol + site + ":" + httpPort
	if https {
		addrToShow = protocol + site + ":" + httpsPort
	}
	fmt.Println()
	log.Printf("roi is start to running. see %s", addrToShow)
	fmt.Println()

	// Bind
	if https {
		if portForwarding {
			go func() {
				log.Fatal(http.ListenAndServe(site+":"+httpPort, http.HandlerFunc(redirectToHttps)))
			}()
		}
		log.Fatal(http.ListenAndServeTLS(site+":"+httpsPort, cert, key, mux))
	} else {
		log.Fatal(http.ListenAndServe(site+":"+httpPort, mux))
	}
}
