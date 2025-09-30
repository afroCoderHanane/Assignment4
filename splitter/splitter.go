package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type SplitterRequest struct {
    S3URL  string `json:"s3_url"`
    Chunks int    `json:"chunks,omitempty"`
}

type SplitterResponse struct {
    ChunkURLs []string `json:"chunk_urls"`
    Timing    float64  `json:"processing_time_ms"`
}

func main() {
    http.HandleFunc("/split", handleSplit)  // or /map or /reduce
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Service starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleSplit(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    ctx := context.Background()
    
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req SplitterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if req.Chunks == 0 {
        req.Chunks = 3 // Default to 3 chunks
    }

    // Parse S3 URL (s3://bucket/key)
    parts := strings.SplitN(strings.TrimPrefix(req.S3URL, "s3://"), "/", 2)
    if len(parts) != 2 {
        http.Error(w, "Invalid S3 URL format", http.StatusBadRequest)
        return
    }
    bucket, key := parts[0], parts[1]

    // Create S3 client
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
        return
    }
    client := s3.NewFromConfig(cfg)

    // Download file from S3
    result, err := client.GetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(key),
    })
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to get object: %v", err), http.StatusInternalServerError)
        return
    }
    defer result.Body.Close()

    content, err := io.ReadAll(result.Body)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to read content: %v", err), http.StatusInternalServerError)
        return
    }

    // Split content into chunks
    words := strings.Fields(string(content))
    wordsPerChunk := len(words) / req.Chunks
    if wordsPerChunk == 0 {
        wordsPerChunk = 1
    }

    var chunkURLs []string
    for i := 0; i < req.Chunks; i++ {
        start := i * wordsPerChunk
        end := start + wordsPerChunk
        if i == req.Chunks-1 {
            end = len(words)
        }
        if start >= len(words) {
            break
        }

        chunkWords := words[start:end]
        chunkContent := strings.Join(chunkWords, " ")

        // Upload chunk to S3
        chunkKey := fmt.Sprintf("chunks/%s-chunk-%d.txt", strings.TrimSuffix(key, ".txt"), i)
        _, err := client.PutObject(ctx, &s3.PutObjectInput{
            Bucket: aws.String(bucket),
            Key:    aws.String(chunkKey),
            Body:   bytes.NewReader([]byte(chunkContent)),
        })
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to upload chunk: %v", err), http.StatusInternalServerError)
            return
        }

        chunkURLs = append(chunkURLs, fmt.Sprintf("s3://%s/%s", bucket, chunkKey))
    }

    elapsed := time.Since(start).Milliseconds()
    
    resp := SplitterResponse{
        ChunkURLs: chunkURLs,
        Timing:    float64(elapsed),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}