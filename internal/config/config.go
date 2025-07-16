package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defAuthEmailsPath         string        = "./data_source/authorized_emails.json"
	defVerificCodeLifetimeSec time.Duration = time.Second * 60 * 15
	defSessionLifetimeSec     time.Duration = time.Second * 60 * 60 * 3
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
	MyIp                 string
	BaseUrl              string
	AppEnv               string
	GinPort              string
	LoginCodeLifetimeSec time.Duration // in seconds
	SessionLifetimeSec   time.Duration // in seconds
}

type PGVectorConfig struct {
	PostgresUser     string
	PostgresDB       string
	PostgresPassword string
	PostgresPort     string
}

type Config struct {
	Common   CommonConfig
	Milvus   MilvusConfig
	Embed    EmbedConfig
	Email    EmailConfig
	Data     DataConfig
	PGVector PGVectorConfig
}

func NewConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	return Config{
		Common:   ReadCommonConfig(),
		Milvus:   ReadMilvusConfig(),
		Embed:    ReadEmbedConfig(),
		Email:    ReadEmailConfig(),
		Data:     ReadDataConfig(),
		PGVector: ReadPGVectorConfig(),
	}
}

func ReadCommonConfig() CommonConfig {
	verificCodeLifetimeSec := defVerificCodeLifetimeSec

	codeLifetimeEnv := os.Getenv("VERIFIC_CODE_LIFETIME_SEC")
	if codeLifetimeEnv != "" {
		verificCodeLifetimeInt, err := strconv.Atoi(codeLifetimeEnv)
		if err != nil {
			log.Fatal(fmt.Sprintf("failed to convert VERIFIC_CODE_LIFETIME_SEC to int: %v", err))
		}

		verificCodeLifetimeSec = time.Duration(verificCodeLifetimeInt) * time.Second
	}

	sessionLifetimeSec := defSessionLifetimeSec

	sessionLifetimeEnv := os.Getenv("SESSION_LIFETIME_SEC")
	if sessionLifetimeEnv != "" {
		sessionLifetimeInt, err := strconv.Atoi(sessionLifetimeEnv)
		if err != nil {
			log.Fatal(fmt.Sprintf("failed to convert SESSION_LIFETIME_SEC to int: %v", err))
		}

		sessionLifetimeSec = time.Duration(sessionLifetimeInt) * time.Second
	}

	return CommonConfig{
		MyIp:                 os.Getenv("MY_IP"),
		BaseUrl:              os.Getenv("BASE_URL"),
		AppEnv:               os.Getenv("APP_ENV"),
		GinPort:              os.Getenv("GIN_PORT"),
		LoginCodeLifetimeSec: verificCodeLifetimeSec,
		SessionLifetimeSec:   sessionLifetimeSec,
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
		log.Fatal(fmt.Errorf("failed to convert port: %w", err))
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
	authEmailsPath := os.Getenv("AUTH_EMAILS_PATH")
	if authEmailsPath == "" {
		authEmailsPath = defAuthEmailsPath
	}

	if _, err := os.Stat(authEmailsPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatal(fmt.Sprintf("auth emails file do not exists: %v", err))
		}

		log.Fatal(fmt.Sprintf("failed to test if auth emails file exist: %v", err))
	}

	return DataConfig{
		AuthEmailsPath: authEmailsPath,
	}
}

func ReadPGVectorConfig() PGVectorConfig {
	return PGVectorConfig{
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresDB:       os.Getenv("POSTGRES_DB"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresPort:     os.Getenv("POSTGRES_PORT"),
	}
}

func (c *Config) IsProduction() bool {
	return c.Common.AppEnv == "production"
}
