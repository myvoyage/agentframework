// Agent Framework - Weather Tool with Real API Integration
// Copyright (C) 2025  Agent Framework Contributors
// SPDX-License-Identifier: AGPL-3.0-or-later

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// WeatherAPIConfig holds configuration for weather API clients
type WeatherAPIConfig struct {
	APIKey      string
	Endpoint    string
	CacheExpiry time.Duration
}

// OpenWeatherMapClient implements real weather data retrieval using OpenWeatherMap API
type OpenWeatherMapClient struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
	cache      map[string]*cachedWeatherData
	cacheMutex *sync.RWMutex
}

type cachedWeatherData struct {
	data      string
	timestamp time.Time
}

// WeatherResponse represents the response from OpenWeatherMap API
type WeatherResponse struct {
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike  float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Name    string `json:"name"`
	Cod     int    `json:"cod"`
	Message string `json:"message,omitempty"`
}

// NewOpenWeatherMapClient creates a new OpenWeatherMap API client
func NewOpenWeatherMapClient(apiKey string) (*OpenWeatherMapClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OpenWeatherMap API key cannot be empty")
	}

	return &OpenWeatherMapClient{
		apiKey:     apiKey,
		endpoint:   "https://api.openweathermap.org/data/2.5/weather",
		httpClient: &http.Client{Timeout: 10 * time.Second},
		cache:      make(map[string]*cachedWeatherData),
		cacheMutex: &sync.RWMutex{},
	}, nil
}

// GetWeather retrieves current weather for a location
func (c *OpenWeatherMapClient) GetWeather(ctx context.Context, location string) (*WeatherResponse, error) {
	// Check cache first
	if cached := c.getCachedWeather(location); cached != nil {
		var resp WeatherResponse
		if err := json.Unmarshal([]byte(cached.data), &resp); err == nil {
			return &resp, nil
		}
	}

	// Build request URL
	reqURL := fmt.Sprintf("%s?q=%s&appid=%s&units=metric",
		c.endpoint,
		url.QueryEscape(location),
		c.apiKey)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get weather: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	// Parse response
	var weatherResp WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Cache the response
	respJSON, _ := json.Marshal(weatherResp)
	c.setCachedWeather(location, string(respJSON), 10*time.Minute)

	return &weatherResp, nil
}

func (c *OpenWeatherMapClient) getCachedWeather(location string) *cachedWeatherData {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	if cached, ok := c.cache[location]; ok && time.Since(cached.timestamp) < 10*time.Minute {
		return cached
	}
	return nil
}

func (c *OpenWeatherMapClient) setCachedWeather(location, data string, ttl time.Duration) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.cache[location] = &cachedWeatherData{
		data:      data,
		timestamp: time.Now(),
	}
}

// FormatWeather formats the weather response into a human-readable string
func FormatWeather(resp *WeatherResponse) string {
	var output string

	output += fmt.Sprintf("Current weather in %s:\n", resp.Name)

	if len(resp.Weather) > 0 {
		output += fmt.Sprintf("  Condition: %s (%s)\n", resp.Weather[0].Main, resp.Weather[0].Description)
	}

	output += fmt.Sprintf("  Temperature: %.1f°C (feels like %.1f°C)\n", resp.Main.Temp, resp.Main.FeelsLike)
	output += fmt.Sprintf("  Min/Max: %.1f°C / %.1f°C\n", resp.Main.TempMin, resp.Main.TempMax)
	output += fmt.Sprintf("  Humidity: %d%%\n", resp.Main.Humidity)
	output += fmt.Sprintf("  Wind Speed: %.1f m/s\n", resp.Wind.Speed)
	output += fmt.Sprintf("  Cloudiness: %d%%\n", resp.Clouds.All)

	return output
}

// NewWeatherTool creates a weather tool that uses real API when configured
func NewWeatherTool() (tool.BaseTool, error) {
	apiKey := os.Getenv("OPENWEATHERMAP_API_KEY")
	if apiKey != "" {
		client, err := NewOpenWeatherMapClient(apiKey)
		if err != nil {
			return nil, err
		}
		return &realWeatherTool{client: client}, nil
	}

	// Fall back to mock tool
	return &mockWeatherTool{}, nil
}

// realWeatherTool uses the real OpenWeatherMap API
type realWeatherTool struct {
	client *OpenWeatherMapClient
}

func (t *realWeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_weather",
		Desc: "Get current weather information for a location using OpenWeatherMap API.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"location": {
				Type: "string",
				Desc: "City name or location (e.g., 'London', 'New York', 'Tokyo')",
			},
		}),
	}, nil
}

func (t *realWeatherTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Location string `json:"location"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if args.Location == "" {
		return "", fmt.Errorf("location parameter is required")
	}

	// Get weather from API
	weatherResp, err := t.client.GetWeather(ctx, args.Location)
	if err != nil {
		return "", fmt.Errorf("failed to get weather: %w", err)
	}

	return FormatWeather(weatherResp), nil
}

// mockWeatherTool is a fallback implementation when no API key is configured
type mockWeatherTool struct{}

func (t *mockWeatherTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_weather_mock",
		Desc: "Mock weather tool (no API key configured). Set OPENWEATHERMAP_API_KEY to use real weather data.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"location": {
				Type: "string",
				Desc: "City name or location",
			},
		}),
	}, nil
}

func (t *mockWeatherTool) InvokableRun(ctx context.Context, input string, opts ...tool.Option) (string, error) {
	var args struct {
		Location string `json:"location"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	return fmt.Sprintf("Mock Weather for %s: Sunny, 25°C\n\nNote: Configure OPENWEATHERMAP_API_KEY environment variable for real weather data.", args.Location), nil
}
