package main

import (
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

const ExtensionsPath = "/liman/extensions/"
const LimanUser = "liman"
const DefaultShell = "/bin/bash"
const ResolvPath = "/etc/resolv.conf"
const DnsOptions = "options rotate timeout:1 retries:1"

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/dns",dnsHandler)
	r.HandleFunc("/userAdd",userAddHandler)
	r.HandleFunc("/userRemove",userRemoveHandler)
	r.HandleFunc("/fixPermissions",permissionFixHandler)
	r.HandleFunc("/certificateAdd",certificateAddHandler)
	r.HandleFunc("/certificateRemove",certificateRemoveHandler)
	_ = http.ListenAndServe("127.0.0.1:1803", r)
}

func dnsHandler(w http.ResponseWriter, r *http.Request) {
	server1 , _ := r.URL.Query()["server1"]
	server2 , _ := r.URL.Query()["server2"]
	server3 , _ := r.URL.Query()["server3"]
	result := setDnsServers(server1[0],server2[0],server3[0])
	if result == true {
		_, _ = w.Write([]byte("DNS updated!\n"))
		w.WriteHeader(http.StatusOK)
	} else {
		_, _ = w.Write([]byte("DNS update failed!\n"))
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

func userAddHandler(w http.ResponseWriter, r *http.Request) {
	extensionId, _ := r.URL.Query()["extensionId"]
	result := addUser(extensionId[0])
	if result == true {
		_, _ = w.Write([]byte("New User Added!\n"))
		w.WriteHeader(http.StatusOK)
	} else {
		_, _ = w.Write([]byte("User add failed!\n"))
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

func userRemoveHandler(w http.ResponseWriter, r *http.Request) {
	extensionId, _ := r.URL.Query()["extensionId"]
	result := removeUser(extensionId[0])
	if result == true {
		_, _ = w.Write([]byte("User Removed!\n"))
		w.WriteHeader(http.StatusOK)
	} else {
		_, _ = w.Write([]byte("User remove failed!\n"))
		w.WriteHeader(http.StatusNotAcceptable)
	}
}

func permissionFixHandler(w http.ResponseWriter, r *http.Request) {
	extensionId, _ := r.URL.Query()["extensionId"]
	extensionName, _ := r.URL.Query()["extensionName"]
	result := fixExtensionPermissions(extensionId[0], extensionName[0])
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

func addUser(extensionId string) bool {
	_, err := executeCommand("useradd -r -s " + DefaultShell + " " + extensionId)
	if err == nil {
		return true
	} else {
		return false
	}
}

func removeUser(extensionId string) bool {
	_, err := executeCommand("userdel " + extensionId)
	if err == nil {
		return true
	} else {
		return false
	}
}

func fixExtensionPermissions(extensionId string, extensionName string) bool {
	_, err := executeCommand("chmod -R 770 " + ExtensionsPath + extensionName)
	if err != nil {
		return false
	}

	_, err = executeCommand("chown -R " + extensionId + ":" + LimanUser + " " + ExtensionsPath + extensionId)
	if err == nil {
		return true
	} else {
		return false
	}
}

func addSystemCertificate(tmpPath string, targetName string) bool {
	certPath , certUpdateCommand := getCertificateStrings()

	_, err := executeCommand("mv " + tmpPath + " " + certPath + "/" + targetName + ".crt")
	if err != nil {
		return false
	}

	_, err = executeCommand(certUpdateCommand)
	if err == nil {
		return true
	} else {
		return false
	}
}

func removeSystemCertificate(targetName string) bool {
	certPath , certUpdateCommand := getCertificateStrings()
	_, err := executeCommand("rm " + certPath + "/" + targetName + ".crt")
	if err != nil {
		return false
	}

	_, err = executeCommand(certUpdateCommand)
	if err == nil {
		return true
	} else {
		return false
	}
}

func getCertificateStrings () (string,string) {
	certPath := "/usr/local/share/ca-certificates/"
	certUpdateCommand := "update-ca-certificates"
	if isCentOs() == true {
		certPath = "/etc/pki/ca-trust/source/anchors/"
		certUpdateCommand = "sudo update-ca-trust"
	}
	return certPath, certUpdateCommand
}

func setDnsServers(server1 string, server2 string, server3 string) bool {
	_, err := executeCommand("chattr -i " + ResolvPath)
	if err != nil {
		println(err.Error())
		return false
	}
	newData := []byte(DnsOptions + "\n" + server1 + "\n" + server2 + "\n" + server3 + "\n")

	err = ioutil.WriteFile(ResolvPath,newData, 0644)

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
