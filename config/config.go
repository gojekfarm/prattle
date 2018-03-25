package config

import (
	"github.com/spf13/viper"
	"os"
	"log"
	"strconv"
)

type Config struct {
	rpcPort     int
	siblingAddr string
	members     string
}

var appConfig *Config

func Load() {
	viper.SetConfigName("application")
	viper.AddConfigPath("./")
	viper.SetConfigType("yaml")
	viper.ReadInConfig()

	viper.AutomaticEnv()
	appConfig = &Config{
		rpcPort:     getIntOrPanic("RPC_PORT"),
		siblingAddr: fatalGetString("SIBLING_ADDR"),
		members:     getString("MEMBERS"),
	}
}

func RpcPort() int {
	return appConfig.rpcPort
}

func SiblingAddr() string {
	return appConfig.siblingAddr
}

func members() string {
	return appConfig.members
}

func fatalGetString(key string) string {
	checkKey(key)
	value := os.Getenv(key)
	if value == "" {
		value = viper.GetString(key)
	}
	return value
}

func panicIfErrorForKey(err error, key string) {
	if err != nil {
		log.Fatalf("Could not parse key: %s, Error: %s", key, err)
	}
}

func getIntOrPanic(key string) int {
	checkKey(key)
	v, err := strconv.Atoi(fatalGetString(key))
	if err != nil {
		v, err = strconv.Atoi(os.Getenv(key))
	}
	panicIfErrorForKey(err, key)
	return v
}

func checkKey(key string) {
	if !viper.IsSet(key) && os.Getenv(key) == "" {
		log.Fatalf("%s key is not set", key)
	}
}

func getString(key string) string {
	value := os.Getenv(key)
	if value == "" {
		value = viper.GetString(key)
	}
	return value
}
