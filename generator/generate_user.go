package generator

import (
	"fmt"
	"math/rand"

	"github.com/songvi/robo/generator/user"
)

// GenerateUser creates a new user based on the UserStrategy configuration
func GenerateUser(strategy user.UserStrategy) (user.User, error) {
	if len(strategy.UserLang) == 0 || len(strategy.LangProbability) == 0 {
		return user.User{}, fmt.Errorf("invalid UserStrategy: user_lang or lang_probability is empty")
	}
	if len(strategy.UserLang) != len(strategy.LangProbability) {
		return user.User{}, fmt.Errorf("invalid UserStrategy: user_lang and lang_probability lengths do not match")
	}

	// Select language based on probability distribution
	langIndex := selectIndexByProbability(strategy.LangProbability)
	language := strategy.UserLang[langIndex]

	// Generate random display name and username
	displayName := user.GenerateDisplayName(strategy)
	username := generateRandomUserName(6, 12)

	return user.User{
		DisplayName: displayName,
		UserName:    username,
		Language:    language,
	}, nil
}

// selectIndexByProbability selects an index based on a probability distribution
func selectIndexByProbability(probabilities []float64) int {
	r := rand.Float64()
	sum := 0.0
	for i, p := range probabilities {
		sum += p
		if r <= sum {
			return i
		}
	}
	return len(probabilities) - 1
}

// generateRandomUserName generates a random string of specified length range
func generateRandomUserName(minLen, maxLen int) string {
	length := minLen + rand.Intn(maxLen-minLen+1)
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
