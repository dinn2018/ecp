package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	tty "github.com/mattn/go-tty"
)

func makeDefaultDir() {
	if home := homeDir(); home != "" {
		var accountPath string
		if runtime.GOOS == "darwin" {
			accountPath = filepath.Join(home, "Library", "Application Support", "Account")
		} else if runtime.GOOS == "windows" {
			accountPath = filepath.Join(home, "AppData", "Roaming", "Account")
		} else {
			accountPath = filepath.Join(home, "Account")
		}
		// if err := os.MkdirAll(accountPath, 0700); err != nil {
		// 	fatal(fmt.Sprintf("create account data dir [%v]: %v", accountPath, err))
		// }
		privPath := filepath.Join(accountPath, "PrivateKey")
		if err := os.MkdirAll(privPath, 0700); err != nil {
			fatal(fmt.Sprintf("create private key dir [%v]: %v", privPath, err))
		}
		keystorePath := filepath.Join(accountPath, "KeyStore")
		if err := os.MkdirAll(keystorePath, 0700); err != nil {
			fatal(fmt.Sprintf("create keystore dir [%v]: %v", keystorePath, err))
		}
	}
}

func makeKeyStoreDirWithExistPath(path string) string {
	keystorePath := filepath.Join(path, "KeyStore")
	if err := os.MkdirAll(keystorePath, 0700); err != nil {
		fatal(fmt.Sprintf("create KeyStore dir [%v]: %v", keystorePath, err))
	}
	return keystorePath
}

func makePrivDirWithExistPath(path string) string {
	pkp := filepath.Join(path, "PrivateKey")
	if err := os.MkdirAll(pkp, 0700); err != nil {
		fatal(fmt.Sprintf("create PrivateKey dir [%v]: %v", pkp, err))
	}
	return pkp
}

func defaultDataDir() string {
	// Try to place the data folder in the user's home dir
	if home := homeDir(); home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "Application Support", "Account")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "Account")
		} else {
			return filepath.Join(home, "Account")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func fatal(args ...interface{}) {
	var w io.Writer
	if runtime.GOOS == "windows" {
		// The SameFile check below doesn't work on Windows.
		// stdout is unlikely to get redirected though, so just print there.
		w = os.Stdout
	} else {
		outf, _ := os.Stdout.Stat()
		errf, _ := os.Stderr.Stat()
		if outf != nil && errf != nil && os.SameFile(outf, errf) {
			w = os.Stderr
		} else {
			w = io.MultiWriter(os.Stdout, os.Stderr)
		}
	}
	fmt.Fprint(w, "Fatal: ")
	fmt.Fprintln(w, args...)
	os.Exit(1)
}

func readPasswordFromNewTTY(prompt string) (string, error) {
	t, err := tty.Open()
	if err != nil {
		return "", err
	}
	defer t.Close()
	fmt.Fprint(t.Output(), prompt)
	pass, err := t.ReadPasswordNoEcho()
	if err != nil {
		return "", err
	}
	return pass, err
}
