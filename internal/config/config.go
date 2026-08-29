package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	EmbedDim uint16
	DbPath   string
	Port     string
}

type ClientConfig struct {
	Endpoint   string
	APIKey     string
	APIVersion string
	Model      string
	EmbedDim   uint16
	Address    string
	Port       string
}

type SanityConfig struct {
	Endpoint   string
	APIKey     string
	APIVersion string
	Model      string
	EmbedDim   uint16
	DbPath     string
}

func LoadServer(filenames ...string) (*ServerConfig, error) {
	if err := loadConfig(filenames...); err != nil {
		return nil, err
	}

	embedDim, err := parseEmbedDim(os.Getenv("EMBEDDING_DIM"))
	if err != nil {
		return nil, err
	}

	return &ServerConfig{
		EmbedDim: embedDim,
		DbPath:   os.Getenv("DB_PATH"),
		Port:     os.Getenv("PORT"),
	}, nil
}

func LoadClient(filenames ...string) (*ClientConfig, error) {
	if err := loadConfig(filenames...); err != nil {
		return nil, err
	}

	embedDim, err := parseEmbedDim(os.Getenv("EMBEDDING_DIM"))
	if err != nil {
		return nil, err
	}

	return &ClientConfig{
		Endpoint:   os.Getenv("AZURE_OPENAI_ENDPOINT"),
		APIKey:     os.Getenv("AZURE_OPENAI_API_KEY"),
		APIVersion: os.Getenv("AZURE_OPENAI_API_VERSION"),
		Model:      os.Getenv("AZURE_OPENAI_MODEL"),
		EmbedDim:   embedDim,
		Address:    os.Getenv("ADDRESS"),
		Port:       os.Getenv("PORT"),
	}, nil
}

func LoadSanity(filenames ...string) (*SanityConfig, error) {
	if err := loadConfig(filenames...); err != nil {
		return nil, err
	}

	embedDim, err := parseEmbedDim(os.Getenv("EMBEDDING_DIM"))
	if err != nil {
		return nil, err
	}

	return &SanityConfig{
		Endpoint:   os.Getenv("AZURE_OPENAI_ENDPOINT"),
		APIKey:     os.Getenv("AZURE_OPENAI_API_KEY"),
		APIVersion: os.Getenv("AZURE_OPENAI_API_VERSION"),
		Model:      os.Getenv("AZURE_OPENAI_MODEL"),
		EmbedDim:   embedDim,
		DbPath:     os.Getenv("DB_PATH"),
	}, nil
}

func loadConfig(filenames ...string) error {
	err := godotenv.Load(filenames...)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("loading env file: %q: %w", filenames, err)
	}

	return nil
}

func parseEmbedDim(embedDim string) (uint16, error) {
	dim64, err := strconv.ParseUint(embedDim, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("parsing EMBEDDING_DIM %q: %w", embedDim, err)
	}

	dim16 := uint16(dim64)
	return dim16, nil
}
