package kopia

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/joho/godotenv"
	"github.com/kopia/kopia/repo"
	"github.com/kopia/kopia/repo/blob/s3"
    // "github.com/kopia/kopia/snapshot"
    )


func InitializeKopiaRepo() {
    if err := godotenv.Load(); err != nil {
    log.Fatal("Error loading .env file")
    }
    
    ctx := context.Background()
    password := os.Getenv("KOPIA_REPO_PASSWORD")
    opts := &s3.Options{
    BucketName:      os.Getenv("S3_BUCKET_NAME"),
    Prefix:          "snapshot-", // potentially add hostname to prefix
    Endpoint:        os.Getenv("S3_ENDPOINT"),
    DoNotUseTLS:     true,
    AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
    SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),
    Region:          os.Getenv("S3_REGION"),
    }
    
    blobStorage, err := s3.New(ctx, opts, true)
    if err != nil {
    log.Fatalf("unable to create s3 blob storage: %v", err)
    }
    fmt.Println("Created s3 blob.")
    
    if err := repo.Initialize(ctx, blobStorage, nil, password); err != nil {
    log.Printf("failed to create repository: %v", err)
    } else {
    fmt.Println("Repository created successfully.")
    }
    
    err = repo.Connect(ctx, "./config.json", blobStorage, password, nil)
    if err != nil {
    log.Fatalf("failed to connect and write config file: %v", err)
    }
    
    log.Println("S3 repository connected and config file created successfully.")

    hostname := repo.GetDefaultHostName(ctx)
    fmt.Println("Default hostname: ", hostname)
}


