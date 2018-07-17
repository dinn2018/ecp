package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mattn/go-isatty"
	"github.com/pborman/uuid"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	version   string
	gitCommit string
	gitTag    string
)

func fullVersion() string {
	versionMeta := "release"
	if gitTag == "" {
		versionMeta = "dev"
	}
	return fmt.Sprintf("%s-%s-%s", version, gitCommit, versionMeta)
}
func main() {
	app := cli.App{
		Version: fullVersion(),
		Name:    "EGP",
		Usage:   "create ecdsa private key or parse it",
		Action:  defaultAction,
		Commands: []cli.Command{
			{
				Name:      "create",
				ShortName: "c",
				Usage:     "create a random account",
				Action:    createAccount,
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "priv",
						Usage: "create a private key",
					},
					cli.BoolFlag{
						Name:  "keystore",
						Usage: "create keystore",
					},
					cli.BoolFlag{
						Name:  "export",
						Usage: "export private key or kestore at default directory, default: false",
					},
					cli.StringFlag{
						Name:  "dir",
						Usage: "export private key or kestore into dir, default:" + defaultDataDir(),
					},
				},
			},
			{
				Name:      "parse",
				ShortName: "p",
				Usage:     "parse a private key or keystore,return an address of ecdsa public key",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "priv",
						Usage: "path of a private key",
					},
					cli.StringFlag{
						Name:  "keystore",
						Usage: "path of keystore",
					},
				},
				Action: parseAccount,
			},
		},
	}

	makeDefaultDir()

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func defaultAction(ctx *cli.Context) {
	cli.ShowAppHelp(ctx)
	fmt.Println("please specify one cmd at least")
}

func createAccount(ctx *cli.Context) error {
	if !ctx.Bool("priv") && !ctx.Bool("keystore") {
		fmt.Println("should specify `priv` or `keystore`")
		return nil
	}

	privKey, err := crypto.GenerateKey()
	if err != nil {
		fmt.Println("generate account failed:", err)
		return err
	}

	if ctx.Bool("keystore") {
		password, err := readPasswordFromNewTTY("Enter passphrase: ")
		if err != nil {
			return err
		}
		if password == "" {
			return errors.New("non-empty passphrase required")
		}
		confirm, err := readPasswordFromNewTTY("Confirm passphrase: ")
		if err != nil {
			return err
		}
		if password != confirm {
			return errors.New("passphrase confirmation mismatch")
		}
		keyjson, err := keystore.EncryptKey(&keystore.Key{
			PrivateKey: privKey,
			Address:    crypto.PubkeyToAddress(privKey.PublicKey),
			Id:         uuid.NewRandom()},
			password, keystore.StandardScryptN, keystore.StandardScryptP)
		if err != nil {
			return err
		}
		if isatty.IsTerminal(os.Stdout.Fd()) {
			fmt.Println("========= JSON keystore =========")
			fmt.Println(string(keyjson))
			fmt.Println("=================================")
		}
		if ctx.Bool("export") {
			kp := filepath.Join(defaultDataDir(), "KeyStore", time.Now().String())
			if ctx.String("dir") != "" {
				kp = filepath.Join(makeKeyStoreDirWithExistPath(ctx.String("dir")), time.Now().String())
			}
			if err := ioutil.WriteFile(kp, keyjson, 0600); err != nil {
				return err
			}
			fmt.Println("KeyStore has been saved at:", kp)
			fmt.Println("=================================")
		}
	}

	if ctx.Bool("priv") {
		fmt.Println("==== your private key of hex ====")
		fmt.Println(hexutil.Encode(crypto.FromECDSA(privKey)))
		fmt.Println("========= your address ==========")
		fmt.Println(hexutil.Encode(crypto.PubkeyToAddress(privKey.PublicKey).Bytes()))
		fmt.Println("=================================")
		if ctx.Bool("export") {
			privPath := filepath.Join(defaultDataDir(), "PrivateKey", time.Now().String())
			if ctx.String("dir") != "" {
				privPath = filepath.Join(makePrivDirWithExistPath(ctx.String("dir")), "PrivateKey", time.Now().String())
			}
			if err := crypto.SaveECDSA(privPath, privKey); err != nil {
				fmt.Println("save private key failed:", err)
				return err
			}
			fmt.Println("private key has been saved at:", privPath)
			fmt.Println("=================================")
		}
	}

	return nil
}

func parseAccount(ctx *cli.Context) error {
	if ctx.String("priv") == "" && ctx.String("keystore") == "" {
		fmt.Println("should specify `priv` or `keystore`")
		return nil
	}

	if ctx.String("priv") != "" {
		pk, err := crypto.LoadECDSA(ctx.String("priv"))
		if err != nil {
			return err
		}
		fmt.Println("=====  your address ====")
		fmt.Println(hexutil.Encode(crypto.PubkeyToAddress(pk.PublicKey).Bytes()))
		fmt.Println("========================")
		return nil
	}

	if ctx.String("keystore") != "" {
		b, err := ioutil.ReadFile(ctx.String("keystore"))
		if err != nil {
			return err
		}
		password, err := readPasswordFromNewTTY("Enter passphrase: ")
		if err != nil {
			return err
		}
		key, err := keystore.DecryptKey(b, password)
		if err != nil {
			return err
		}
		fmt.Println("=====  your private key ====")
		fmt.Println(hexutil.Encode(crypto.FromECDSA(key.PrivateKey)))
		fmt.Println("======   your address  =====")
		fmt.Println(hexutil.Encode(key.Address.Bytes()))
		fmt.Println("============================")
	}
	return nil
}
