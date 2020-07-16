package main

import (
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
	storeRandomKey()
	r.HandleFunc("/dns", dnsHandler)
	r.HandleFunc("/userAdd", userAddHandler)
	r.HandleFunc("/userRemove", userRemoveHandler)
	r.HandleFunc("/fixPermissions", permissionFixHandler)
	r.HandleFunc("/certificateAdd", certificateAddHandler)
	r.HandleFunc("/certificateRemove", certificateRemoveHandler)
	r.HandleFunc("/fixExtensionKeysPermission", fixExtensionKeyHandler)
	r.HandleFunc("/extensionRun", runExtensionHandler)
	r.HandleFunc("/test", testHandler)
	r.Use(loggingMiddleware)
	r.Use(verifyTokenMiddleware)
	_ = http.ListenAndServe("127.0.0.1:3008", r)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func verifyTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limanToken, _ := r.URL.Query()["liman_token"]
		if len(limanToken) == 0 || limanToken[0] != currentToken {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("It works!\n"))
	w.WriteHeader(http.StatusOK)
}

func dnsHandler(w http.ResponseWriter, r *http.Request) {
	server1, _ := r.URL.Query()["server1"]
	server2, _ := r.URL.Query()["server2"]
	server3, _ := r.URL.Query()["server3"]
	result := setDNSServers(server1[0], server2[0], server3[0])
	if result == true {
		_, _ = w.Write([]byte("DNS updated!\n"))
		w.WriteHeader(http.StatusOK)
	} else {
		_, _ = w.Write([]byte("DNS update failed!\n"))
		w.WriteHeader(http.StatusNotAcceptable)
	}
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
		_, _ = w.Write([]byte("New User Added!\n"))
		w.WriteHeader(http.StatusOK)
	} else {
		_, _ = w.Write([]byte("User add failed!\n"))
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

func userRemoveHandler(w http.ResponseWriter, r *http.Request) {
	extensionID, _ := r.URL.Query()["extension_id"]
	result := removeUser(extensionID[0])
	if result == true {
		_, _ = w.Write([]byte("User Removed!\n"))
		w.WriteHeader(http.StatusOK)
	} else {
		_, _ = w.Write([]byte("User remove failed!\n"))
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

func permissionFixHandler(w http.ResponseWriter, r *http.Request) {
	extensionID, _ := r.URL.Query()["extension_id"]
	extensionName, _ := r.URL.Query()["extensionName"]
	result := fixExtensionPermissions(extensionID[0], extensionName[0])
	if result == true {
		_, _ = w.Write([]byte("Permissions fixed!\n"))
		w.WriteHeader(http.StatusOK)
	} else {
		_, _ = w.Write([]byte("Permission fix failed!\n"))
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

func certificateAddHandler(w http.ResponseWriter, r *http.Request) {
	tmpPath, _ := r.URL.Query()["tmpPath"]
	targetName, _ := r.URL.Query()["targetName"]
	result := addSystemCertificate(tmpPath[0], targetName[0])
	if result == true {
		_, _ = w.Write([]byte("New Certificate Added!\n"))
		w.WriteHeader(http.StatusOK)
	} else {
		_, _ = w.Write([]byte("Certificate add failed!\n"))
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

func certificateRemoveHandler(w http.ResponseWriter, r *http.Request) {
	targetName, _ := r.URL.Query()["targetName"]
	result := removeSystemCertificate(targetName[0])
	if result == true {
		_, _ = w.Write([]byte("Certificate removed!\n"))
		w.WriteHeader(http.StatusOK)
	} else {
		_, _ = w.Write([]byte("Certificate remove failed!\n"))
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

func runExtensionHandler(w http.ResponseWriter, r *http.Request) {
	command, _ := r.URL.Query()["command"]
	output := runExtensionCommand(command[0])
	_, _ = w.Write([]byte(output))
	w.WriteHeader(http.StatusOK)
}

func runExtensionCommand(command string) string {
	out, err := executeCommand(command)
	if err == nil {
		return out
	}
	return err.Error()
}

func fixExtensionKeys(extensionID string) bool {
	_, err := executeCommand("chmod -R 700 " + ExtensionKeysPath + extensionID)
	if err != nil {
		return false
	}

	_, err = executeCommand("chown -R " + extensionID + ":" + LimanUser + " " + ExtensionKeysPath + extensionID)
	if err == nil {
		return true
	}
	return false
}

func storeRandomKey() {
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
	_, err2 := executeCommand("chown liman:liman " + AuthKeyPath)
	if err != nil || err2 != nil {
		panic("Key can't be stored")
	}
}

func addUser(extensionID string) bool {
	_, err := executeCommand("useradd -r -s " + DefaultShell + " " + extensionID)
	if err == nil {
		return true
	}
	return false
}

func removeUser(extensionID string) bool {
	_, err := executeCommand("userdel " + extensionID)
	if err == nil {
		return true
	}
	return false
}

func fixExtensionPermissions(extensionID string, extensionName string) bool {
	_, err := executeCommand("chmod -R 770 " + ExtensionsPath + extensionName)
	if err != nil {
		return false
	}

	_, err = executeCommand("chown -R " + extensionID + ":" + LimanUser + " " + ExtensionsPath + extensionID)
	if err == nil {
		return true
	}
	return false
}

func addSystemCertificate(tmpPath string, targetName string) bool {
	certPath, certUpdateCommand := getCertificateStrings()

	_, err := executeCommand("mv " + tmpPath + " " + certPath + "/" + targetName + ".crt")
	if err != nil {
		return false
	}

	_, err = executeCommand(certUpdateCommand)
	if err == nil {
		return true
	}
	return false
}

func removeSystemCertificate(targetName string) bool {
	certPath, certUpdateCommand := getCertificateStrings()
	_, err := executeCommand("rm " + certPath + "/" + targetName + ".crt")
	if err != nil {
		return false
	}

	_, err = executeCommand(certUpdateCommand)
	if err == nil {
		return true
	}
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
	_, err := executeCommand("chattr -i " + ResolvPath)
	if err != nil {
		return false
	}
	newData := []byte(DNSOptions + "\n" + server1 + "\n" + server2 + "\n" + server3 + "\n")

	err = ioutil.WriteFile(ResolvPath, newData, 0644)

	if err != nil {
		return false
	}

	_, err = executeCommand("chattr +i " + ResolvPath)
	if err != nil {
		return false
	}
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
