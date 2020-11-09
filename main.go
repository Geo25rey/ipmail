package main

import (
	"flag"
	"github.com/kyoh86/xdg"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
	"ipmail/cli"
	"ipmail/ipmail"
	"ipmail/ipmail/crypto"
	"os"
	"path"
	"strings"
)

const configName = "config"

func setupConfig() error {
	viper.SetConfigName(configName)
	viper.SetConfigType("prop")
	viper.AddConfigPath(xdg.ConfigHome() + "/ipmail")
	_ = os.Mkdir(xdg.ConfigHome()+"/ipmail", os.ModeDir|0755)
	_ = os.Mkdir(xdg.DataHome()+"/ipmail", os.ModeDir|0755)
	if _, err := os.Stat(path.Join(xdg.ConfigHome()+"/ipmail", configName+".prop")); os.IsNotExist(err) {
		err := ioutil.WriteFile(
			path.Join(xdg.ConfigHome()+"/ipmail", configName+".prop"),
			[]byte(""),
			0660)
		if err != nil {
			println(err.Error())
		}
		return nil
	}
	return viper.ReadInConfig()
}

func parseCmdLine() error {
	flag.String("config", xdg.ConfigHome()+"/ipmail"+"/"+configName+".prop", "loads specified config file")
	flag.String("identity", xdg.DataHome()+"/ipmail"+"/"+"identity", "")
	flag.String("contacts", xdg.DataHome()+"/ipmail"+"/"+"contacts", "")
	flag.String("messages", xdg.DataHome()+"/ipmail"+"/"+"messages", "")
	flag.String("sent", xdg.DataHome()+"/ipmail"+"/"+"sent", "")
	flag.String("requests", xdg.DataHome()+"/ipmail"+"/"+"requests", "")
	flag.String("ipfs-repo", xdg.DataHome()+"/ipmail"+"/"+"ipfs-repo", "")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return err
	}
	_ = os.Mkdir(viper.GetString("ipfs-repo"), os.ModeDir|0755)

	config := viper.GetString("config")
	if strings.Compare(config, viper.ConfigFileUsed()) != 0 {
		viper.SetConfigFile(config)
		err := viper.ReadInConfig()
		if err != nil {
			println("warning:", err.Error())
		} else {
			_ = viper.BindPFlags(pflag.CommandLine) // ignore err is handled by return
		}
	}

	err = viper.WriteConfig()
	return err
}

func main() {
	err := setupConfig()
	if err != nil {
		panic(err)
	}
	err = parseCmdLine()
	if err != nil {
		panic(err)
	}
	ipfsRepo := viper.GetString("ipfs-repo")
	ipfs, err := ipmail.NewIpfsWithRepo(false, &ipfsRepo)
	if err != nil {
		panic(err)
	}
	sender := ipmail.NewSender(ipfs)
	receiver, err := ipmail.NewReceiver(crypto.MessageTopicName, ipfs)
	if err != nil {
		panic(err)
	}
	var identity crypto.SelfIdentity = nil
	identityFile := viper.GetString("identity")
	identity = crypto.NewSelfIdentityFromFile(identityFile) // nil if file not found
	var contacts crypto.ContactsIdentityList = nil
	contactsFile := viper.GetString("contacts")
	contacts, _ = crypto.NewContactsIdentityListFromFile(contactsFile) // nil if file not found
	messagesFile := viper.GetString("messages")
	messages := ipmail.NewMessageListFromFile(messagesFile, ipfs, identity, contacts) // nil if file not found
	sentFile := viper.GetString("sent")
	sent := ipmail.NewMessageListFromFile(sentFile, ipfs, identity, contacts) // nil if file not found
	requestsFile := viper.GetString("requests")
	requests := ipmail.NewMessageListFromFile(requestsFile, ipfs, identity, contacts) // nil if file not found
	cli.Run(ipfs, sender, receiver, identity, contacts, messages, sent, requests)
	receiver.Close()
}
