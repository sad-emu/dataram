package ramformats

// Function to generate UUID
import (
	"github.com/google/uuid"
)

func GenerateUUID() string {
	// Generate a new UUID
	return uuid.New().String()
}
