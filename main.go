package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Version 1.2
const Version = "1.3"

// ExtensionsPath : Liman's Extension Folder
const ExtensionsPath = "/liman/extensions/"

// LimanUser : Just in case if liman user changed to something else.
const LimanUser = "liman"

// DefaultShell : Default sh shell
const DefaultShell = "/bin/bash"

// ResolvPath : Dns server' configuration path.
const ResolvPath = "/etc/resolv.conf"

// DNSOptions : Options to have multiple dns servers
const DNSOptions = "options rotate timeout:1 retries:1"

// AuthKeyPath : Key Path for liman to auth with this system
const AuthKeyPath = "/liman/keys/service.key"

// ExtensionKeysPath : Extension Key' Path to fix permissions
const ExtensionKeysPath = "/liman/keys/"

var currentToken = ""

func main() {
	r := mux.NewRouter()
	log.Println("Starting Liman System Helper & Extension Renderer")
	log.Println("Version : " + Version)
	storeRandomKey()
	r.HandleFunc("/dns", dnsHandler)
	r.HandleFunc("/userAdd", userAddHandler)
	r.HandleFunc("/userRemove", userRemoveHandler)
	r.HandleFunc("/fixPermissions", permissionFixHandler)
	r.HandleFunc("/certificateAdd", certificateAddHandler)
	r.HandleFunc("/installPackages", packageInstallHandler)
	r.HandleFunc("/certificateRemove", certificateRemoveHandler)
	r.HandleFunc("/fixExtensionKeysPermission", fixExtensionKeyHandler)
	r.HandleFunc("/extensionRun", runExtensionHandler)
	r.HandleFunc("/test", testHandler)
	r.Use(loggingMiddleware)
	r.Use(verifyTokenMiddleware)
	log.Println("Service started at 127.0.0.1:3008")
	log.Fatal(http.ListenAndServe("127.0.0.1:3008", r))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func verifyTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limanToken, _ := r.URL.Query()["liman_token"]
		if len(limanToken) == 0 || limanToken[0] != currentToken {
			log.Println("Invalid Auth Token Received")
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("It works!\n"))
}

func dnsHandler(w http.ResponseWriter, r *http.Request) {
	server1, _ := r.URL.Query()["server1"]
	server2, _ := r.URL.Query()["server2"]
	server3, _ := r.URL.Query()["server3"]
	result := setDNSServers(server1[0], server2[0], server3[0])
	if result == true {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("DNS updated!\n"))
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
		_, _ = w.Write([]byte("DNS update failed!\n"))
	}
}

func packageInstallHandler(w http.ResponseWriter, r *http.Request) {
	packages := r.URL.Query()["packages"][0]

	command := "if [ -z '$(find /var/cache/apt/pkgcache.bin -mmin -60)' ]; then sudo apt-get update; fi;"
	go func() {
		executeCommand(command)
		cmd := exec.Command("sudo", "apt-get", "install", "-o", "Dpkg::Use-Pty=0", "-o", "Dpkg::Options::='--force-confdef'", "-o", "Dpkg::Options::='--force-confold'", packages, "-qqy", "--force-yes")
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "DEBIAN_FRONTEND=noninteractive")
		cmd.Run()
	}()

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Installation requested!\n"))
}

func fixExtensionKeyHandler(w http.ResponseWriter, r *http.Request) {
	extensionID, _ := r.URL.Query()["extension_id"]
	result := fixExtensionKeys(extensionID[0])
	if result == true {
		_, _ = w.Write([]byte("Key permissions updated!\n"))
		w.WriteHeader(http.StatusOK)
	} else {
		_, _ = w.Write([]byte("Key permission update failed!\n"))
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

func userAddHandler(w http.ResponseWriter, r *http.Request) {
	extensionID, _ := r.URL.Query()["extension_id"]
	result := addUser(extensionID[0])
	if result == true {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("New User Added!\n"))
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
		_, _ = w.Write([]byte("User add failed!\n"))
	}
}

func userRemoveHandler(w http.ResponseWriter, r *http.Request) {
	extensionID, _ := r.URL.Query()["extension_id"]
	result := removeUser(extensionID[0])
	if result == true {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("User Removed!\n"))
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
		_, _ = w.Write([]byte("User remove failed!\n"))
	}
}

func cleanDash(text string) (string){
	return strings.ReplaceAll(text, "-", "")
}

func permissionFixHandler(w http.ResponseWriter, r *http.Request) {
	extensionID, _ := r.URL.Query()["extension_id"]
	extensionName, _ := r.URL.Query()["extension_name"]

	result := fixExtensionPermissions(extensionID[0], extensionName[0])
	result = fixExtensionKeys(extensionID[0])
	if result == true {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Permissions fixed!\n"))
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
		_, _ = w.Write([]byte("Permission fix failed!\n"))
	}
}

func certificateAddHandler(w http.ResponseWriter, r *http.Request) {
	tmpPath, _ := r.URL.Query()["tmpPath"]
	targetName, _ := r.URL.Query()["targetName"]
	result := addSystemCertificate(tmpPath[0], targetName[0])
	if result == true {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("New Certificate Added!\n"))
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
		_, _ = w.Write([]byte("Certificate add failed!\n"))
	}
}

func certificateRemoveHandler(w http.ResponseWriter, r *http.Request) {
	targetName, _ := r.URL.Query()["targetName"]
	result := removeSystemCertificate(targetName[0])
	if result == true {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Certificate removed!\n"))
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
		_, _ = w.Write([]byte("Certificate remove failed!\n"))
	}
}

func runExtensionHandler(w http.ResponseWriter, r *http.Request) {
	command, _ := r.URL.Query()["command"]
	output := runExtensionCommand(command[0])
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(output))
}

func runExtensionCommand(command string) string {
	out, _ := exec.Command(DefaultShell, "-c", command).Output()
	return string(out)
}

func runExtensionBackgroundCommand(command string, userID string, handler string) {
	output := runExtensionCommand(command)
	requestBody, _ := json.Marshal(map[string]string{
		"output":       output,
		"user_id":      userID,
		"system_token": currentToken,
		"handler":      handler,
	})
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	time.Sleep(1 * time.Second)
	http.Post("https://127.0.0.1/lmn/private/extensionRenderedHandler", "application/json", bytes.NewBuffer(requestBody))
	log.Println("Background Render result sent to Liman 🤓")
}

func fixExtensionKeys(extensionID string) bool {
	_, err := exec.Command("chmod", "-R", "700", ExtensionKeysPath+extensionID).Output()
	if err != nil {
		return false
	}

	_, err = exec.Command("chown", "-R", cleanDash(extensionID)+":"+LimanUser, ExtensionKeysPath+extensionID).Output()
	if err == nil {
		return true
	}
	return false
}

func storeRandomKey() {
	if _, err := os.Stat(AuthKeyPath); err == nil {
		log.Println("Reading existing key.")
		dat, _ := ioutil.ReadFile(AuthKeyPath)
		currentToken = string(dat)
		return
	}

	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ" +
		"abcdefghijklmnopqrstuvwxyzåäö" +
		"0123456789")
	length := 32
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	key := b.String()
	d1 := []byte(key)
	currentToken = string(d1)
	err := ioutil.WriteFile(AuthKeyPath, d1, 0700)
	_, err2 := exec.Command("chown", "liman:liman", AuthKeyPath).Output()
	if err != nil || err2 != nil {
		panic("Key can't be stored")
	}
}

func addUser(extensionID string) bool {
	log.Println("Adding System User : " + extensionID)
	_, err := exec.Command("useradd", "-r", "-s", DefaultShell, extensionID).Output()
	if err == nil {
		log.Println("System User Added : " + extensionID)
		return true
	}
	log.Println(err)
	return false
}

func removeUser(extensionID string) bool {
	log.Println("Removing System User : " + extensionID)
	_, err := exec.Command("userdel", extensionID).Output()
	if err == nil {
		log.Println("System User Removed : " + extensionID)
		return true
	}
	log.Println(err)
	return false
}

func fixExtensionPermissions(extensionID string, extensionName string) bool {
	_, err := exec.Command("chmod", "-R", "770", ExtensionsPath+extensionName).Output()
	log.Println("Fixing Extension Permissions")
	if err != nil {
		log.Println(err)
		return false
	}

	_, err = exec.Command("chown", "-R", cleanDash(extensionID)+":"+LimanUser, ExtensionsPath+extensionName).Output()
	if err == nil {
		log.Println("Extension Permissions Fixed")
		return true
	}
	log.Println(err)
	return false
}

func addSystemCertificate(tmpPath string, targetName string) bool {
	certPath, certUpdateCommand := getCertificateStrings()
	log.Println("Adding System Certificate")
	_, err := exec.Command("mv", tmpPath, certPath+"/"+targetName+".crt").Output()
	if err != nil {
		log.Println(err)
		return false
	}

	_, err = executeCommand(certUpdateCommand + " 2>&1")
	if err == nil {
		log.Println("System Certificate Added")
		return true
	}
	log.Println(err)
	return false
}

func removeSystemCertificate(targetName string) bool {
	log.Println("Removing System Certificate")
	certPath, certUpdateCommand := getCertificateStrings()
	_, err := exec.Command("rm ", certPath+"/"+targetName+".crt").Output()
	if err != nil {
		log.Println(err)
		return false
	}

	_, err = executeCommand(certUpdateCommand)
	if err == nil {
		log.Println("System Certificate Removed")
		return true
	}
	log.Println(err)
	return false
}

func getCertificateStrings() (string, string) {
	certPath := "/usr/local/share/ca-certificates/"
	certUpdateCommand := "update-ca-certificates"
	if isCentOs() == true {
		certPath = "/etc/pki/ca-trust/source/anchors/"
		certUpdateCommand = "sudo update-ca-trust"
	}
	return certPath, certUpdateCommand
}

func setDNSServers(server1 string, server2 string, server3 string) bool {
	_, err := exec.Command("chattr", "-i", ResolvPath).Output()
	log.Println("Updating DNS Servers")
	if err != nil {
		log.Println(err)
		return false
	}
	newData := DNSOptions + "\n"
	if server1 != "" {
		newData += "nameserver " + server1 + "\n"
	}

	if server2 != "" {
		newData += "nameserver " + server2 + "\n"
	}

	if server3 != "" {
		newData += "nameserver " + server3 + "\n"
	}

	err = ioutil.WriteFile(ResolvPath, []byte(newData), 0644)

	if err != nil {
		log.Println(err)
		return false
	}

	_, err = exec.Command("chattr", "+i", ResolvPath).Output()
	if err != nil {
		log.Println(err)
		return false
	}
	log.Println("DNS Servers Updated")
	return true
}

func isCentOs() bool {
	_, err := os.Stat("/etc/redhat-release")
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func executeCommand(input string) (string, error) {
	cmd := exec.Command(DefaultShell, "-c", input)
	stdout, stderr := cmd.Output()
	return string(stdout), stderr
}
