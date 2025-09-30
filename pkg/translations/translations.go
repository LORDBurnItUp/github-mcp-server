package translations

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type TranslationHelperFunc func(key string, defaultValue string) string

func NullTranslationHelper(_ string, defaultValue string) string {
	return defaultValue
}

func TranslationHelper() (TranslationHelperFunc, func()) {
	var translationKeyMap = map[string]string{}
    v := viper.New()

	// Load from JSON file
	v.SetConfigName("github-mcp-server-config")
	v.SetConfigType("json")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		// ignore error if file not found as it is not required
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Could not read JSON config: %v", err)
		}
	}

    // Ensure environment variable support with the expected prefix and key format
    v.AutomaticEnv()
    v.SetEnvPrefix("GITHUB_MCP")
    v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

    // create a function that takes both a key, and a default value and returns either the default value or an override value
	return func(key string, defaultValue string) string {
            // Normalize the key to upper snake-case for env lookups and viper
            normalizedKey := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
            if value, exists := translationKeyMap[normalizedKey]; exists {
                return value
            }

            // Set default and then read via viper which respects env overrides
            v.SetDefault(normalizedKey, defaultValue)
            resolved := v.GetString(normalizedKey)
            translationKeyMap[normalizedKey] = resolved
            return resolved
		}, func() {
			// dump the translationKeyMap to a json file
			if err := DumpTranslationKeyMap(translationKeyMap); err != nil {
				log.Fatalf("Could not dump translation key map: %v", err)
			}
		}
}

// DumpTranslationKeyMap writes the translation map to a json file called github-mcp-server-config.json
func DumpTranslationKeyMap(translationKeyMap map[string]string) error {
	file, err := os.Create("github-mcp-server-config.json")
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer func() { _ = file.Close() }()

	// marshal the map to json
	jsonData, err := json.MarshalIndent(translationKeyMap, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling map to JSON: %v", err)
	}

	// write the json data to the file
	if _, err := file.Write(jsonData); err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}
