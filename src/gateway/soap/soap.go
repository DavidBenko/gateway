package soap

import (
	"fmt"
	"gateway/config"
	"log"
	"os/exec"
	"path"
	"regexp"
	"strconv"
)

const bin = "bin"
const wsimport = "wsimport"
const java = "java"
const minSupportedJdkVersion = 8 // as in Java 1.8

var jdkHome string
var fullJavaCommandPath = java
var fullWsimportCommandPath = wsimport

var javaAvailable = false
var wsimportAvailable = false

var javaVersionRegex = regexp.MustCompile("^java version \"1\\.(\\d+)\\..+\"")

// Available indicates whether or not the dependencies are met so that SOAP
// remote endpoints may be available
func Available() bool {
	return javaAvailable && wsimportAvailable
}

// Configure initializes the soap package
func Configure(soap config.Soap) error {
	jdkHome = soap.JdkPath

	if jdkHome != "" {
		fullJavaCommandPath = path.Join(path.Clean(jdkHome), bin, java)
		fullWsimportCommandPath = path.Join(path.Clean(jdkHome), bin, wsimport)
	}

	cmd := exec.Command(fullJavaCommandPath, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Received error checking for existence of java command: %s", err)
	}

	javaVersion, _ := strconv.Atoi(javaVersionRegex.FindStringSubmatch(string(output))[1])
	if javaVersion < minSupportedJdkVersion {
		return fmt.Errorf("Invalid Java version: Java must be version 1.8 or higher")
	}

	javaAvailable = true

	cmd = exec.Command(fullWsimportCommandPath, "-version")
	output, err = cmd.CombinedOutput()
	if err != nil {
		wsimportAvailable = false
		return fmt.Errorf("Received error checking for existince of wsimport command: %s", err)
	}

	wsimportAvailable = true

	// TODO launch the JVM

	return nil
}

// Wsimport runs the java utility 'wsimport' if it is available on the local system
func Wsimport(wsdlFile string, jarOutputFile string) error {
	if !Available() {
		return fmt.Errorf("Wsimport is not configured on this system")
	}

	cmd := exec.Command(fullWsimportCommandPath, "-extension", "-clientjar", jarOutputFile, wsdlFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error invoking wsimport: %v", err)
	}

	log.Println(string(output))

	return nil
}
