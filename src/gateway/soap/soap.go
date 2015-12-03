package soap

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	"gateway/config"
	aperrors "gateway/errors"
	"gateway/logger"
)

const (
	bin                    = "bin"
	wsimport               = "wsimport"
	java                   = "java"
	minSupportedJdkVersion = 8 // as in Java 1.8
	classpathOption        = "-cp"
	soapMainClass          = "com.anypresence.wsclient.Wsclient"
)

var (
	splitter = regexp.MustCompile(`\s+`)

	jdkHome                 string
	fullJavaCommandPath     = java
	fullWsimportCommandPath = wsimport

	javaAvailable     = false
	wsimportAvailable = false
	soapAvailable     = false

	javaVersionRegex = regexp.MustCompile("^java version \"1\\.(\\d+)\\..+\"")

	jvmCmd *exec.Cmd
)

// Available indicates whether or not the dependencies are met so that SOAP
// remote endpoints may be available
func Available() bool {
	return javaAvailable && wsimportAvailable && soapAvailable
}

// Configure initializes the soap package
func Configure(soap config.Soap, devMode bool) error {
	jdkHome = soap.JdkPath

	var err error
	if jdkHome != "" {
		fullJavaCommandPath = path.Join(path.Clean(jdkHome), bin, java)
		fullWsimportCommandPath = path.Join(path.Clean(jdkHome), bin, wsimport)
	}

	// ensure that we have valid full paths to each executable
	fullJavaCommandPath, err = exec.LookPath(fullJavaCommandPath)
	if err != nil {
		return fmt.Errorf("Received error attempting to execute LookPath for java")
	}
	fullWsimportCommandPath, err = exec.LookPath(fullWsimportCommandPath)
	if err != nil {
		return fmt.Errorf("Received error attempting to execute LookPath for wsimport")
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

	jarFile, err := inflateSoapClient()
	if err != nil {
		return err
	}

	soapAvailable = true

	return launchJvm(soap, jarFile, devMode)
}

// Wsimport runs the java utility 'wsimport' if it is available on the local system
func Wsimport(wsdlFile string, jarOutputFile string) error {
	if !Available() {
		return fmt.Errorf("Wsimport is not configured on this system")
	}

	bytes, err := ioutil.ReadFile(wsdlFile)
	if err != nil {
		return aperrors.NewWrapped("[soap.go] Unable to read wsdl file", err)
	}

	// Use a generated package name that consists of the WSDL's md5 sum, prefixed with 'v' (for version)
	// This will prevent class version collisions when an update is made to a WSDL
	checksum := md5.Sum(bytes)
	hexsum := hex.EncodeToString(checksum[:])
	packageName := fmt.Sprintf("v%s", hexsum)

	jarOutputFile, err = filepath.Abs(jarOutputFile)
	if err != nil {
		return aperrors.NewWrapped("[soap.go] Unable to get absolute path for jar file", err)
	}

	tmpdir, err := ioutil.TempDir("", "wsimport")
	if err != nil {
		return aperrors.NewWrapped("[soap.go] Unable to create tmp directory prior to invoking wsimport", err)
	}

	defer os.Remove(tmpdir)

	cmd := exec.Command(fullWsimportCommandPath, "-d", tmpdir, "-p", packageName, "-extension", "-clientjar", jarOutputFile, wsdlFile)
	output, err := cmd.CombinedOutput()
	logger.Println(string(output))

	if err != nil {
		return fmt.Errorf("Error invoking wsimport: %v", err)
	}

	return nil
}

// EnsureJarPath makes certain that the directory in which jar files are stored exists
func EnsureJarPath() (string, error) {
	dirPerm := os.FileMode(os.ModeDir | 0700)

	dir := path.Clean(path.Join(".", "tmp", "jaxws"))
	err := os.MkdirAll(dir, dirPerm)
	return dir, err
}

// JarURLForSoapRemoteEndpointID takes a remote soap endpoint ID and returns the path where the
// generated JAR will reside
func JarURLForSoapRemoteEndpointID(remoteEndpointID int64) (string, error) {
	jarPath, err := EnsureJarPath()
	if err != nil {
		return "", err
	}

	filePath := path.Join(jarPath, fmt.Sprintf("%d.jar", remoteEndpointID))
	fullFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("file://%s", fullFilePath), nil
}

func inflateSoapClient() (string, error) {
	jarBytes, err := Asset("soapclient-all.jar")

	if err != nil {
		logger.Printf("%s Could not find embedded soapclient", config.System)
		return "", err
	}

	jarPath, err := EnsureJarPath()
	if err != nil {
		return "", aperrors.NewWrapped("[soap.go] Unable to ensure jar path!", err)
	}

	// Write the soapclient jar out to the filesystem
	jarDestFilename := path.Join(jarPath, "soapclient.jar")
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
