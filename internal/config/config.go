package config

import (
	"os"

	"github.com/joho/godotenv"
)

type EmailConfig struct {
	User     string
	Password string
	Host     string
	Port     string
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
	}
}

func ReadCommonConfig() CommonConfig {
	return CommonConfig{
		MyIp:    os.Getenv("BASE_URL"),
		BaseUrl: os.Getenv("MY_IP"),
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
	return EmailConfig{
		User:     os.Getenv("EMAIL_SERVER_USER"),
		Password: os.Getenv("EMAIL_SERVER_PASSWORD"),
		Host:     os.Getenv("EMAIL_SERVER_HOST"),
		Port:     os.Getenv("EMAIL_SERVER_PORT"),
		From:     os.Getenv("EMAIL_FROM"),
	}
}

func (c *Config) IsProduction() bool {
	return c.Common.AppEnv == "production"
}
