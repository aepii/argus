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

type CoordinatorConfig struct {
	VirtualNodes      uint16
	ReplicationFactor uint16
}

func LoadServer(filenames ...string) (*ServerConfig, error) {
	if err := loadConfig(filenames...); err != nil {
		return nil, err
	}

	embedDim, err := parseUint16(os.Getenv("EMBEDDING_DIM"), "EMBEDDING_DIM")
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

	embedDim, err := parseUint16(os.Getenv("EMBEDDING_DIM"), "EMBEDDING_DIM")
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

	embedDim, err := parseUint16(os.Getenv("EMBEDDING_DIM"), "EMBEDDING_DIM")
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

func LoadCoordinator(filenames ...string) (*CoordinatorConfig, error) {
	if err := loadConfig(filenames...); err != nil {
		return nil, err
	}

	virtualNodes, err := parseUint16(os.Getenv("VIRTUAL_NODES"), "VIRTUAL_NODES")
	if err != nil {
		return nil, err
	}

	replicationFactor, err := parseUint16(os.Getenv("REPLICATION_FACTOR"), "REPLICATION_FACTOR")
	if err != nil {
		return nil, err
	}

	return &CoordinatorConfig{
		VirtualNodes:      virtualNodes,
		ReplicationFactor: replicationFactor,
	}, nil
}

func loadConfig(filenames ...string) error {
	err := godotenv.Load(filenames...)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("loading env file: %q: %w", filenames, err)
	}

	return nil
}

func parseUint16(v string, name string) (uint16, error) {
	v64, err := strconv.ParseUint(v, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("parsing %s=%q: %w", name, v, err)
	}

	return uint16(v64), nil
}
