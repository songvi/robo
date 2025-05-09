package generator

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/songvi/robo/generator/workspace"
)

// GenerateWorkspace creates a new workspace with a randomly selected list of user UUIDs based on the WorkspaceStrategy
func GenerateWorkspace(wsStrategy workspace.WorkspaceStrategy, availableUserUUIDs []string) (workspace.Workspace, error) {
	rand.Seed(time.Now().UnixNano())

	// Validate WorkspaceStrategy
	if len(wsStrategy.NumberOfUsers) == 0 || len(wsStrategy.NumberOfUsersProbability) == 0 {
		return workspace.Workspace{}, fmt.Errorf("invalid WorkspaceStrategy: number_of_users or number_of_users_probability is empty")
	}
	if len(wsStrategy.NumberOfUsers) != len(wsStrategy.NumberOfUsersProbability) {
		return workspace.Workspace{}, fmt.Errorf("invalid WorkspaceStrategy: number_of_users and number_of_users_probability lengths do not match")
	}

	// Validate availableUserUUIDs
	if len(availableUserUUIDs) == 0 {
		return workspace.Workspace{}, fmt.Errorf("no available user UUIDs provided")
	}

	// Select number of users based on probability
	numUsersIndex := selectWorkspaceIndexByProbability(wsStrategy.NumberOfUsersProbability)
	numUsers := wsStrategy.NumberOfUsers[numUsersIndex]

	// Ensure we don't select more users than available
	if numUsers > len(availableUserUUIDs) {
		numUsers = len(availableUserUUIDs)
	}

	// Shuffle available UUIDs to select random users
	uuids := make([]string, len(availableUserUUIDs))
	copy(uuids, availableUserUUIDs)
	rand.Shuffle(len(uuids), func(i, j int) {
		uuids[i], uuids[j] = uuids[j], uuids[i]
	})

	// Select UUIDs for the workspace
	selectedUUIDs := uuids[:numUsers]

	// Generate workspace name
	workspaceName := generateWspRandomName(8, 16)

	return workspace.Workspace{
		Name:  workspaceName,
		Users: selectedUUIDs,
	}, nil
}

// selectWorkspaceIndexByProbability selects an index based on a probability distribution
func selectWorkspaceIndexByProbability(probabilities []float64) int {
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

// generateWspRandomName generates a random string of specified length range
func generateWspRandomName(minLen, maxLen int) string {
	length := minLen + rand.Intn(maxLen-minLen+1)
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
