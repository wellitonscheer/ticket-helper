package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type DataConfig struct {
	AuthEmailsPath string
}

type EmailConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	From     string
}

type EmbedConfig struct {
	Port          string
	ContainerName string
}

type MilvusConfig struct {
	MilvulPort        string
	AttuPort          string
	AttuContainerName string
}

type CommonConfig struct {
	MyIp    string
	BaseUrl string
	AppEnv  string
	GinPort string
}

type Config struct {
	Common CommonConfig
	Milvus MilvusConfig
	Embed  EmbedConfig
	Email  EmailConfig
	Data   DataConfig
}

func NewConfig() Config {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	return Config{
		Common: ReadCommonConfig(),
		Milvus: ReadMilvusConfig(),
		Embed:  ReadEmbedConfig(),
		Email:  ReadEmailConfig(),
		Data:   ReadDataConfig(),
	}
}

func ReadCommonConfig() CommonConfig {
	return CommonConfig{
		MyIp:    os.Getenv("MY_IP"),
		BaseUrl: os.Getenv("BASE_URL"),
		AppEnv:  os.Getenv("APP_ENV"),
		GinPort: os.Getenv("GIN_PORT"),
	}
}

func ReadMilvusConfig() MilvusConfig {
	return MilvusConfig{
		MilvulPort:        os.Getenv("MILVUS_PORT"),
		AttuPort:          os.Getenv("ATTU_PORT"),
		AttuContainerName: os.Getenv("ATTU_CONTAINER_NAME"),
	}
}

func ReadEmbedConfig() EmbedConfig {
	return EmbedConfig{
		Port:          os.Getenv("EMBED_PORT"),
		ContainerName: os.Getenv("EMBED_CONTAINER_NAME"),
	}
}

func ReadEmailConfig() EmailConfig {
	port, err := strconv.Atoi(os.Getenv("EMAIL_SERVER_PORT"))
	if err != nil {
		panic(fmt.Errorf("failed to convert port: %w", err))
	}

	return EmailConfig{
		User:     os.Getenv("EMAIL_SERVER_USER"),
		Password: os.Getenv("EMAIL_SERVER_PASSWORD"),
		Host:     os.Getenv("EMAIL_SERVER_HOST"),
		Port:     port,
		From:     os.Getenv("EMAIL_FROM"),
	}
}

func ReadDataConfig() DataConfig {
	const defaultAuthEmailsPath = "./data_source/authorized_emails.json"

	authEmailsPath := os.Getenv("AUTH_EMAILS_PATH")
	if authEmailsPath == "" {
		authEmailsPath = defaultAuthEmailsPath
	}

	if _, err := os.Stat(authEmailsPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			panic(fmt.Sprintf("auth emails file do not exists: %v", err))
		}

		panic(fmt.Sprintf("failed to test if auth emails file exist: %v", err))
	}

	return DataConfig{
		AuthEmailsPath: authEmailsPath,
	}
}

func (c *Config) IsProduction() bool {
	return c.Common.AppEnv == "production"
}
