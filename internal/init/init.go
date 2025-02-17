package init

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// EnvConfig holds the required configuration values.
type EnvConfig struct {
	PrivKeyWallet        string
	HeliusHTTPSURI       string
	HeliusWSSURI         string
	HeliusHTTPSURITx     string
	JupHTTPSQuoteURI     string
	JupHTTPSSwapURI      string
	JupHTTPSPriceURI     string
	DexHTTPSLatestTokens string
}

func init() {
	// Load .env file if available
	_ = godotenv.Load()

	requiredEnvVars := []string{
		"PRIV_KEY_WALLET",
		"HELIUS_HTTPS_URI",
		"HELIUS_WSS_URI",
		"HELIUS_HTTPS_URI_TX",
		"JUP_HTTPS_QUOTE_URI",
		"JUP_HTTPS_SWAP_URI",
		"JUP_HTTPS_PRICE_URI",
		"DEX_HTTPS_LATEST_TOKENS",
	}

	// Check for missing variables (allow PRIV_KEY_WALLET to be empty)
	var missingVars []string
	for _, envVar := range requiredEnvVars {
		if envVar == "PRIV_KEY_WALLET" {
			continue
		}
		if os.Getenv(envVar) == "" {
			missingVars = append(missingVars, envVar)
		}
	}
	if len(missingVars) > 0 {
		log.Printf("🚫 Missing required environment variables: %s", strings.Join(missingVars, ", "))
		os.Exit(1)
	}

	// Validate PRIV_KEY_WALLET length if provided
	privKeyWallet := os.Getenv("PRIV_KEY_WALLET")
	if privKeyWallet != "" {
		length := len(privKeyWallet)
		if length != 87 && length != 88 {
			log.Printf("🚫 PRIV_KEY_WALLET must be 87 or 88 characters long (got %d)", length)
			os.Exit(1)
		}
	}

	// Helper to validate URL environment variables
	validateURL := func(envVar, expectedProtocol string, checkApiKey bool) {
		value := os.Getenv(envVar)
		if value == "" {
			log.Printf("🚫 %s is missing or empty", envVar)
			os.Exit(1)
		}

		parsed, err := url.Parse(value)
		if err != nil {
			log.Printf("🚫 Failed to parse %s: %v", envVar, err)
			os.Exit(1)
		}

		expectedScheme := strings.TrimSuffix(expectedProtocol, ":")
		if parsed.Scheme != expectedScheme {
			log.Printf("🚫 %s must start with %s", envVar, expectedProtocol)
			os.Exit(1)
		}

		if checkApiKey {
			apiKey := parsed.Query().Get("api-key")
			if strings.TrimSpace(apiKey) == "" {
				log.Printf("🚫 The 'api-key' parameter is missing or empty in the URL: %s", value)
				os.Exit(1)
			}
		}
	}

	// Validate the URL variables with appropriate protocols and API key checks
	validateURL("HELIUS_HTTPS_URI", "https:", true)
	validateURL("HELIUS_WSS_URI", "wss:", true)
	validateURL("HELIUS_HTTPS_URI_TX", "https:", true)
	validateURL("JUP_HTTPS_QUOTE_URI", "https:", false)
	validateURL("JUP_HTTPS_SWAP_URI", "https:", false)
	validateURL("JUP_HTTPS_PRICE_URI", "https:", false)
	validateURL("DEX_HTTPS_LATEST_TOKENS", "https:", false)

	// Check for "{function}" in HELIUS_HTTPS_URI_TX
	heliusHTTPSURITx := os.Getenv("HELIUS_HTTPS_URI_TX")
	if strings.Contains(heliusHTTPSURITx, "{function}") {
		log.Printf("🚫 HELIUS_HTTPS_URI_TX contains {function}. Check your configuration.")
		os.Exit(1)
	}

	// Build the configuration struct (if needed, you can assign it to a package-level variable)
	_ = &EnvConfig{
		PrivKeyWallet:        privKeyWallet,
		HeliusHTTPSURI:       os.Getenv("HELIUS_HTTPS_URI"),
		HeliusWSSURI:         os.Getenv("HELIUS_WSS_URI"),
		HeliusHTTPSURITx:     heliusHTTPSURITx,
		JupHTTPSQuoteURI:     os.Getenv("JUP_HTTPS_QUOTE_URI"),
		JupHTTPSSwapURI:      os.Getenv("JUP_HTTPS_SWAP_URI"),
		JupHTTPSPriceURI:     os.Getenv("JUP_HTTPS_PRICE_URI"),
		DexHTTPSLatestTokens: os.Getenv("DEX_HTTPS_LATEST_TOKENS"),
	}
}
