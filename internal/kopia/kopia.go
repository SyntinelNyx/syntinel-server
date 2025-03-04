package kopia

import (
	"context"
	"fmt"
	"os"

	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/repo/blob/filesystem"
    "github.com/kopia/kopia/repo/content/format"
)

func initialize_kopia_repo(){
    // Define the repository configuration
    repoPath := "/data/kopia-repo"
    password := "your-password"

    // Create a filesystem-based storage
    storage, err := filesystem.New(context.Background(), &filesystem.Options{
        Path: repoPath,
    }, true)
    if err != nil {
        fmt.Printf("Failed to create storage: %v\n", err)
        os.Exit(1)
    }

    // Initialize the repository
    err = repo.Initialize(context.Background(), storage, &repo.NewRepositoryOptions{
        BlockFormat: format.FormattingOptions{
            Hash:       "HMAC-SHA256",
            Encryption: "AES256-GCM-HMAC-SHA256",
        },
    }, password)
    if err != nil {
        fmt.Printf("Failed to initialize repository: %v\n", err)
        os.Exit(1)
    }

    fmt.Println("Repository initialized successfully")
}