package soap

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gateway/config"
	aperrors "gateway/errors"
	"gateway/logreport"
)

const (
	bin                    = "bin"
	java                   = "java"
	minSupportedJdkVersion = 8 // as in Java 1.8
	classpathOption        = "-cp"
	soapMainClass          = "com.anypresence.wsclient.Wsclient"
)

var (
	splitter = regexp.MustCompile(`\s+`)

	jdkHome             string
	fullJavaCommandPath = java

	javaAvailable = false
	soapAvailable = false

	javaVersionRegex = regexp.MustCompile("^(java|openjdk) version \"1\\.(\\d+)\\..+\"")

	jvmCmd *exec.Cmd
)

// Available indicates whether or not the dependencies are met so that SOAP
// remote endpoints may be available
func Available() bool {
	return javaAvailable && soapAvailable
}

// Configure initializes the soap package
func Configure(soap config.Soap, devMode bool) error {
	jdkHome = soap.JdkPath

	var err error
	if jdkHome != "" {
		fullJavaCommandPath = path.Join(path.Clean(jdkHome), bin, java)
	}

	// ensure that we have valid full paths to each executable
	fullJavaCommandPath, err = exec.LookPath(fullJavaCommandPath)
	if err != nil {
		return fmt.Errorf("Received error attempting to execute LookPath for java")
	}

	cmd := exec.Command(fullJavaCommandPath, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Received error from java -version command.  Output is %s", output)
		return fmt.Errorf("Received error checking for existence of java command: %s", err)
	}

	lines := strings.Split(string(output), "\n")
	var match *string
	for _, line := range lines {
		matches := javaVersionRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			match = &matches[2]
			break
		}
	}
	if match == nil {
		return fmt.Errorf("Unable to detect Java version!  Output of java -version is:\n\n%s", output)
	}

	javaVersion, _ := strconv.Atoi(*match)
	if javaVersion < minSupportedJdkVersion {
		return fmt.Errorf("Invalid Java version: Java must be version 1.8 or higher")
	}

	javaAvailable = true

	jarFile, err := inflateSoapClient()
	if err != nil {
		return err
	}

	soapAvailable = true

	return launchJvm(soap, jarFile, devMode)
}

// EnsureWsdlPath makes certain that the directory in which wsdl files are stored exists
func EnsureWsdlPath() (string, error) {
	dirPerm := os.FileMode(os.ModeDir | 0700)

	dir := path.Clean(path.Join(".", "tmp", "wsdls"))
	err := os.MkdirAll(dir, dirPerm)
	return dir, err
}

// WsdlURLForSoapRemoteEndpointID takes a remote soap endpoint ID and returns the path where the
// WSDL will reside
func WsdlURLForSoapRemoteEndpointID(remoteEndpointID int64) (string, error) {
	wsdlPath, err := EnsureWsdlPath()
	if err != nil {
		return "", err
	}

	filePath := path.Join(wsdlPath, fmt.Sprintf("%d.wsdl", remoteEndpointID))
	fullFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("file:///%s", strings.Replace(filepath.ToSlash(fullFilePath), " ", "%20", -1)), nil
}

func inflateSoapClient() (string, error) {
	jarBytes, err := Asset("soapclient-all.jar")

	if err != nil {
		logreport.Printf("%s Could not find embedded soapclient", config.System)
		return "", err
	}

	wsdlPath, err := EnsureWsdlPath()
	if err != nil {
		return "", aperrors.NewWrapped("[soap.go] Unable to ensure WSDL path!", err)
	}

	// Write the soapclient jar out to the filesystem
	jarDestFilename := path.Join(wsdlPath, "soapclient.jar")
	file, err := os.Create(jarDestFilename)
	if err != nil {
		return "", aperrors.NewWrapped("[soap.go] Unable to open file to write soapclient.jar", err)
	}
	defer file.Close()

	_, err = file.Write(jarBytes)
	if err != nil {
		return "", aperrors.NewWrapped("[soap.go] Error occurred while writing soapclient.jar", err)
	}

	return jarDestFilename, nil
}

func launchJvm(soap config.Soap, clientJarFile string, devMode bool) error {
	if jvmCmd != nil {
		return fmt.Errorf("Unable to start JVM -- JVM may already be running?")
	}

	javaCommandArgs := []string{}
	if soap.JavaOpts != "" {
		javaCommandArgs = append(javaCommandArgs, splitter.Split(soap.JavaOpts, -1)...)
	}
	if devMode {
		javaCommandArgs = append(javaCommandArgs, "-DDEBUG=true")
	}

	javaCommandArgs = append(javaCommandArgs, classpathOption, clientJarFile, soapMainClass, soap.SoapClientHost, fmt.Sprintf("%d", soap.SoapClientPort), fmt.Sprintf("%d", soap.ThreadPoolSize))

	cmd := exec.Command(fullJavaCommandPath, javaCommandArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()

	if err != nil {
		return aperrors.NewWrapped("[soap.go] Error creating command for running soap client", err)
	}

	jvmCmd = cmd

	return nil
}

// Shutdown gracefully shuts down the soap client
func Shutdown(sig os.Signal) error {
	if jvmCmd == nil {
		return nil
	}

	if err := jvmCmd.Process.Signal(sig); err != nil {
		return err
	}

	return jvmCmd.Wait()
}
