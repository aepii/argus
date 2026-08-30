package embed

import (
	"context"
	"errors"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/azure"
)

type Client struct {
	client openai.Client
	model  string
	dim    uint16
}

func NewClient(endpoint string, apiKey string, apiVersion string, model string, dim uint16) (*Client, error) {
	if len(endpoint) == 0 {
		return nil, errors.New("endpoint is required")
	}
	if len(apiKey) == 0 {
		return nil, errors.New("api key is required")
	}
	if len(apiVersion) == 0 {
		return nil, errors.New("api version is required")
	}
	if len(model) == 0 {
		return nil, errors.New("api model is required")
	}
	if dim == 0 {
		return nil, errors.New("dimensions must be greater than zero")
	}

	client := openai.NewClient(
		azure.WithEndpoint(endpoint, apiVersion),
		azure.WithAPIKey(apiKey),
	)

	return &Client{client: client, model: model, dim: dim}, nil
}

func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	if len(text) == 0 {
		return nil, errors.New("raw text is required")
	}

	response, err := c.client.Embeddings.New(
		ctx,
		openai.EmbeddingNewParams{
			Model: c.model,
			Input: openai.EmbeddingNewParamsInputUnion{
				OfString: openai.String(text),
			},
			Dimensions: openai.Int(int64(c.dim)),
		},
	)
	if err != nil {
		return nil, err
	}

	emb := response.Data[0].Embedding
	emb32 := make([]float32, len(emb))
	for i, v := range emb {
		emb32[i] = float32(v)
	}

	return emb32, nil
}

func (c *Client) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, errors.New("array of raw text is required")
	}

	response, err := c.client.Embeddings.New(
		ctx,
		openai.EmbeddingNewParams{
			Model: c.model,
			Input: openai.EmbeddingNewParamsInputUnion{
				OfArrayOfStrings: texts,
			},
			Dimensions: openai.Int(int64(c.dim)),
		},
	)
	if err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(texts))
	for id, data := range response.Data {
		emb := data.Embedding
		emb32 := make([]float32, len(emb))
		for i, v := range emb {
			emb32[i] = float32(v)
		}
		embeddings[id] = emb32
	}

	return embeddings, nil
}
