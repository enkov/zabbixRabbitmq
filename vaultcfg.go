package main

import (
	log "github.com/Sirupsen/logrus"

	"github.com/hashicorp/vault/api"
)

func readCfgVault(path, addr, token string) (map[string]interface{}, error) {
	config := api.DefaultConfig()
	client, err := api.NewClient(config)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatalf("Can't create connection to vault")
	}

	client.SetAddress(addr)
	client.SetToken(token)

	secret, err := client.Logical().Read(path)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("error reading from vault")
	}
	return secret.Data, err
}
