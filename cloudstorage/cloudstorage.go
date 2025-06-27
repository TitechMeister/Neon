package cloudstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

// uploadFile uploads an object.
func UploadFile(w io.Writer, bucket, objectName string) (*string, error) {
	// bucket := "bucket-name"
	// object := "object-name"
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	// Open local file.
	f, err := os.Open(objectName)
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w", err)
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	o := client.Bucket(bucket).Object(objectName)

	// Optional: set a generation-match precondition to avoid potential race
	// conditions and data corruptions. The request to upload is aborted if the
	// object's generation number does not match your precondition.
	// For an object that does not yet exist, set the DoesNotExist precondition.
	o = o.If(storage.Conditions{DoesNotExist: true})

	// Upload an object with storage.Writer.
	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return nil, fmt.Errorf("io.Copy: %w", err)
	}
	if err := wc.Close(); err != nil {
		return nil, fmt.Errorf("Writer.Close: %w", err)
	}
	fmt.Println(w, "Blob %v uploaded.\n", objectName)
	url, err := GenerateSignedURL(bucket, objectName, time.Hour*24) // Generate a signed URL for the uploaded object
	if err != nil {
		return nil, fmt.Errorf("OnGenerateSignedURL: %w", err)
	}
	return url, nil
}

// GenerateSignedURL generates a signed URL for downloading an object from Cloud Storage.
// This function requires either:
// 1. GOOGLE_APPLICATION_CREDENTIALS environment variable pointing to a service account key file
// 2. Or running on Google Cloud Platform with default service account
func GenerateSignedURL(bucket, objectName string, expiration time.Duration) (*string, error) {
	// Method 1: Try to get service account info from environment
	serviceAccountPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if serviceAccountPath != "" {
		return GenerateSignedURLWithServiceAccount(bucket, objectName, serviceAccountPath, expiration)
	}

	// Method 2: Use default credentials (works on GCP)
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expiration),
	}

	url, err := storage.SignedURL(bucket, objectName, opts)
	if err != nil {
		return nil, fmt.Errorf("SignedURL: %w", err)
	}

	return &url, nil
}

// GenerateSignedURLWithServiceAccount generates a signed URL using a service account key file.
func GenerateSignedURLWithServiceAccount(bucket, objectName, serviceAccountPath string, expiration time.Duration) (*string, error) {
	// Read the service account key file
	serviceAccountKey, err := os.ReadFile(serviceAccountPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read service account key: %w", err)
	}

	// Parse the service account key to get the email
	var serviceAccount struct {
		ClientEmail string `json:"client_email"`
		PrivateKey  string `json:"private_key"`
	}

	if err := json.Unmarshal(serviceAccountKey, &serviceAccount); err != nil {
		return nil, fmt.Errorf("failed to parse service account key: %w", err)
	}

	opts := &storage.SignedURLOptions{
		GoogleAccessID: serviceAccount.ClientEmail,
		PrivateKey:     []byte(serviceAccount.PrivateKey),
		Method:         "GET",
		Expires:        time.Now().Add(expiration),
		Scheme:         storage.SigningSchemeV4,
	}

	url, err := storage.SignedURL(bucket, objectName, opts)
	if err != nil {
		return nil, fmt.Errorf("SignedURL: %w", err)
	}

	return &url, nil
}

// GetPublicURL returns the public URL for an object (only works if the object is publicly accessible).
func GetPublicURL(bucket, objectName string) string {
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucket, objectName)
}
