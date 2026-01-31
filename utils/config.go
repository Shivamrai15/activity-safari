package utils

import (
	"os"
	// "github.com/joho/godotenv"
)

var Config map[string]string

func GetEnvValue(key string, defaultValue string, throwError bool) string {
	value := os.Getenv(key)
	if value == "" {
		if throwError {
			panic(key + " environment variable not set")
		}
		return defaultValue
	}
	return value
}

func InitConfig() {

	// err := godotenv.Load()
	// if err != nil {
	// 	panic("Error loading .env file")
	// }

	Config = make(map[string]string)

	Config["PORT"] = GetEnvValue("PORT", "3000", false)

	Config["JWT_ACCESS_SECRET"] = GetEnvValue("JWT_ACCESS_SECRET", "", true)
	Config["JWT_REFRESH_SECRET"] = GetEnvValue("JWT_REFRESH_SECRET", "", true)
	
	Config["REDIS_URL"] = GetEnvValue("REDIS_URL", "", true)
	Config["REDIS_PASSWORD"] = GetEnvValue("REDIS_PASSWORD", "", true)

}

func GetConfigValue(key string) string {
	return Config[key]
}
