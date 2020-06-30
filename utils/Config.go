package utils

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-redis/redis"
)

type EmailConfig struct {
	SenderName     string   `json:"sender_name"`
	SenderEmail    string   `json:"sender_email"`
	ReceiverEmails []string `json:"receiver_emails"`
	APIKey         string   `json:"api_key"`
}

type PlaybookConfig struct {
	DefaultPath     string   `json:"default_path"`
	SpecialBranches []string `json:"special_branches"`
	SpecialPath     string   `json:"special_path"`
}

type RedisConfig struct {
	Host     string `json:"redis_host"`
	Port     string `json:"redis_port"`
	DB       int    `json:"redis_db"`
	Password string `json:"redis_password"`
}

type ConfigurationMap struct {
	Address   	string         `json:"address"`
	Redis     	RedisConfig    `json:"redis"`
	Playbooks 	PlaybookConfig `json:"playbooks"`
	Email     	EmailConfig    `json:"email"`
	ConsulToken string		   `json:"consul_token"`
	ConsulAddr	string		   `json:"consul_addr"`
}

var Address string
var Playbooks *PlaybookConfig
var RedisClient *redis.Client
var Email *EmailConfig
var ConsulToken string
var ConsulAddr string

// This is the only class allowed to use log.Fatal
func ReadConfig() func() {
	configuration := loadConfigurationFile()
	Address = configuration.Address
	ConsulToken = configuration.ConsulToken
	ConsulAddr = configuration.ConsulAddr
	Playbooks = &configuration.Playbooks
	Email = &configuration.Email
	RedisClient = configuration.connectRedis()
	return closeServices
}

func loadConfigurationFile() *ConfigurationMap {
	var configuration ConfigurationMap
	fileName := getConfigurationFileName()
	fmt.Printf("CONF FILE: %v\n", fileName)

	fileName, _ = filepath.Abs(fileName)
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err.Error())
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Fatal(err.Error())
	}
	return &configuration
}

func closeServices() {
	if RedisClient != nil {
		RedisClient.Close()
	}
}

func (configuration *ConfigurationMap) connectRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     configuration.Redis.Host + ":" + configuration.Redis.Port,
		Password: configuration.Redis.Password,
		DB:       configuration.Redis.DB,
	})
}

func PlaybookPath(branchName string) string {
	var playbook_path string
	if checkSpecialBranch(branchName) {
		playbook_path = Playbooks.SpecialPath
	} else {
		playbook_path = Playbooks.DefaultPath
	}
	return playbook_path
}

func checkSpecialBranch(branchName string) bool {
	if branchName == "" ||
		Playbooks == nil ||
		Playbooks.SpecialBranches == nil {
		return false
	}
	for _, value := range Playbooks.SpecialBranches {
		if value == branchName {
			return true
		}
	}
	return false
}

func SetConfigurationFlag() {
	flag.String("config", "./conf/configuration.json", "config file")
	flag.Parse()
}

func getConfigurationFileName() string {
	if !flag.Parsed() {
		SetConfigurationFlag()
	}
	fileName := flag.Lookup("config").Value.(flag.Getter).Get().(string)
	return fileName
}
