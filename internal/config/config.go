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
	defAuthEmailsPath      string        = "./data_source/authorized_emails.json"
	defVerificCodeLifetime time.Duration = time.Second * 60 * 15
	defSessionLifetime     time.Duration = time.Second * 60 * 60 * 3
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

type CommonConfig struct {
	MyIp              string
	BaseUrl           string
	AppEnv            string
	GinPort           string
	LoginCodeLifetime time.Duration // in seconds
	SessionLifetime   time.Duration // in seconds
}

type PGVectorConfig struct {
	PostgresUser     string
	PostgresDB       string
	PostgresPassword string
	PostgresPort     string
}

type LLMConfig struct {
	LLMPort                string
	LLMContextLengthTokens int
	LLMModel               string
	LLMTemperature         float32
	LLMMaxTokens           int
	LLMStream              bool
}

type Config struct {
	Common   CommonConfig
	Embed    EmbedConfig
	Email    EmailConfig
	Data     DataConfig
	PGVector PGVectorConfig
	LLM      LLMConfig
}

func NewConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	return Config{
		Common:   ReadCommonConfig(),
		Embed:    ReadEmbedConfig(),
		Email:    ReadEmailConfig(),
		Data:     ReadDataConfig(),
		PGVector: ReadPGVectorConfig(),
		LLM:      ReadLLMConfig(),
	}
}

func ReadCommonConfig() CommonConfig {
	verificCodeLifetime := defVerificCodeLifetime

	codeLifetimeEnv := os.Getenv("VERIFIC_CODE_LIFETIME")
	if codeLifetimeEnv != "" {
		verificCodeLifetimeInt, err := strconv.Atoi(codeLifetimeEnv)
		if err != nil {
			log.Fatalf("failed to convert VERIFIC_CODE_LIFETIME to int: %v", err)
		}

		verificCodeLifetime = time.Duration(verificCodeLifetimeInt) * time.Second
	}

	sessionLifetime := defSessionLifetime

	sessionLifetimeEnv := os.Getenv("SESSION_LIFETIME")
	if sessionLifetimeEnv != "" {
		sessionLifetimeInt, err := strconv.Atoi(sessionLifetimeEnv)
		if err != nil {
			log.Fatalf("failed to convert SESSION_LIFETIME to int: %v", err)
		}

		sessionLifetime = time.Duration(sessionLifetimeInt) * time.Second
	}

	return CommonConfig{
		MyIp:              os.Getenv("MY_IP"),
		BaseUrl:           os.Getenv("BASE_URL"),
		AppEnv:            os.Getenv("APP_ENV"),
		GinPort:           os.Getenv("GIN_PORT"),
		LoginCodeLifetime: verificCodeLifetime,
		SessionLifetime:   sessionLifetime,
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
			log.Fatalf("auth emails file do not exists: %v", err)
		}

		log.Fatalf("failed to test if auth emails file exist: %v", err)
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

func ReadLLMConfig() LLMConfig {
	contextLengthTokens, err := strconv.Atoi(os.Getenv("LLM_CONTEXT_LENGTH_TOKENS"))
	if err != nil {
		log.Fatal(fmt.Errorf("failed to convert llm context length tokens: %w", err))
	}

	temperature, err := strconv.ParseFloat(os.Getenv("LLM_TEMPERATURE"), 32)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to convert llm temperature: %w", err))
	}

	maxTokens, err := strconv.Atoi(os.Getenv("LLM_MAX_TOKENS"))
	if err != nil {
		log.Fatal(fmt.Errorf("failed to convert llm max tokens: %w", err))
	}

	stream, err := strconv.ParseBool(os.Getenv("LLM_STREAM"))
	if err != nil {
		log.Fatal(fmt.Errorf("failed to convert llm stream: %w", err))
	}

	return LLMConfig{
		LLMPort:                os.Getenv("LLM_PORT"),
		LLMContextLengthTokens: contextLengthTokens,
		LLMModel:               os.Getenv("LLM_MODEL"),
		LLMTemperature:         float32(temperature),
		LLMMaxTokens:           maxTokens,
		LLMStream:              stream,
	}
}

func (c *Config) IsProduction() bool {
	return c.Common.AppEnv == "production"
}
