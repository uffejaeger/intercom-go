package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type metadata struct {
	Source             string `json:"source"`
	UpstreamRepository string `json:"upstreamRepository"`
	UpstreamPath       string `json:"upstreamPath"`
	UpstreamCommit     string `json:"upstreamCommit"`
	APIVersion         string `json:"apiVersion"`
}

type commitResponse struct {
	SHA string `json:"sha"`
}

func main() {
	var metadataPath string
	var specPath string
	var ref string

	flag.StringVar(&metadataPath, "metadata", "spec/metadata.json", "path to spec metadata JSON")
	flag.StringVar(&specPath, "spec", "spec/intercom.openapi.yaml", "path to write pinned OpenAPI spec")
	flag.StringVar(&ref, "ref", "main", "upstream git ref to inspect")
	flag.Parse()

	if err := run(metadataPath, specPath, ref); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "update-spec: %v\n", err)
		os.Exit(1)
	}
}

func run(metadataPath, specPath, ref string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	meta, err := readMetadata(metadataPath)
	if err != nil {
		return err
	}
	if meta.UpstreamRepository == "" {
		return fmt.Errorf("metadata upstreamRepository is required")
	}
	if meta.UpstreamPath == "" {
		return fmt.Errorf("metadata upstreamPath is required")
	}

	commit, err := latestCommit(ctx, meta.UpstreamRepository, meta.UpstreamPath, ref)
	if err != nil {
		return err
	}

	spec, err := downloadSpec(ctx, meta.UpstreamRepository, meta.UpstreamPath, commit)
	if err != nil {
		return err
	}
	if err := os.WriteFile(specPath, spec, 0o644); err != nil {
		return fmt.Errorf("write spec: %w", err)
	}

	meta.UpstreamCommit = commit
	meta.Source = fmt.Sprintf("https://github.com/%s/blob/%s/%s", meta.UpstreamRepository, commit, meta.UpstreamPath)
	if err := writeMetadata(metadataPath, meta); err != nil {
		return err
	}

	fmt.Printf("Pinned %s/%s at %s\n", meta.UpstreamRepository, meta.UpstreamPath, commit)
	return nil
}

func readMetadata(path string) (metadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return metadata{}, fmt.Errorf("read metadata: %w", err)
	}

	var meta metadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return metadata{}, fmt.Errorf("parse metadata: %w", err)
	}
	return meta, nil
}

func writeMetadata(path string, meta metadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}
	return nil
}

func latestCommit(ctx context.Context, repo, path, ref string) (string, error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/commits/%s?path=%s", repo, url.PathEscape(ref), url.QueryEscape(path))
	var response commitResponse
	if err := getJSON(ctx, endpoint, &response); err != nil {
		return "", fmt.Errorf("fetch latest upstream commit: %w", err)
	}
	if response.SHA == "" {
		return "", fmt.Errorf("latest upstream commit response did not include sha")
	}
	return response.SHA, nil
}

func downloadSpec(ctx context.Context, repo, path, commit string) ([]byte, error) {
	endpoint := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", repo, url.PathEscape(commit), path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create spec request: %w", err)
	}
	req.Header.Set("Accept", "application/yaml,text/yaml,text/plain")
	addAuth(req)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download spec: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return nil, fmt.Errorf("download spec returned %s: %s", res.Status, strings.TrimSpace(string(body)))
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read spec response: %w", err)
	}
	return data, nil
}

func getJSON(ctx context.Context, endpoint string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	addAuth(req)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return fmt.Errorf("%s: %s", res.Status, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(res.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func addAuth(req *http.Request) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}
