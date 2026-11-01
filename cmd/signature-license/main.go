// Command signature-license issues the partnership licences of
// hyz-is/arandu-signature.
//
//	signature-license -genkey -key ~/.config/hyzis/signature-license.key
//	signature-license -key ~/.config/hyzis/signature-license.key \
//	    -project demo-douglas -domain demo-douglas.hyz.is -years 10
//
// -genkey writes a new Ed25519 private key, readable by its owner only, and
// prints the public key to embed in the module. Without it, the command signs
// a licence with the key and prints it. The private key never leaves the file:
// it is not printed, logged or echoed.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hyz-is/arandu-signature/internal/licence"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "signature-license:", err)
		os.Exit(1)
	}
}

func run() error {
	genkey := flag.Bool("genkey", false, "write a new private key to -key and print its public key")
	keyPath := flag.String("key", "", "the Ed25519 private key file")
	project := flag.String("project", "", "the project the licence is for")
	domains := flag.String("domain", "", "the hosts it is valid on, comma separated")
	years := flag.Int("years", 10, "how many years it is valid for")
	flag.Parse()

	if *keyPath == "" {
		return errors.New("-key is required")
	}
	if *genkey {
		if _, err := os.Stat(*keyPath); err == nil {
			return fmt.Errorf("%s already exists; refusing to replace a key that may have signed licences", *keyPath)
		}
		public, private, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}
		if err := os.WriteFile(*keyPath, []byte(base64.StdEncoding.EncodeToString(private.Seed())+"\n"), 0o600); err != nil {
			return err
		}
		fmt.Println(base64.StdEncoding.EncodeToString(public))
		return nil
	}

	if *project == "" || *domains == "" || *years <= 0 {
		return errors.New("-project, -domain and a positive -years are required")
	}
	raw, err := os.ReadFile(*keyPath)
	if err != nil {
		return err
	}
	seed, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil || len(seed) != ed25519.SeedSize {
		return fmt.Errorf("%s is not a key this command wrote", *keyPath)
	}
	var hosts []string
	for _, d := range strings.Split(*domains, ",") {
		if d = strings.ToLower(strings.TrimSpace(d)); d != "" {
			hosts = append(hosts, d)
		}
	}
	now := time.Now().UTC()
	token, err := licence.Sign(licence.Claims{
		Project: *project, Domains: hosts,
		Issued: now.Unix(), Expires: now.AddDate(*years, 0, 0).Unix(),
	}, ed25519.NewKeyFromSeed(seed))
	if err != nil {
		return err
	}
	fmt.Println(token)
	return nil
}
