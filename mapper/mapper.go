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
    "unicode"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type MapperRequest struct {
    ChunkURL string `json:"chunk_url"`
}

type MapperResponse struct {
    ResultURL string  `json:"result_url"`
    WordCount int     `json:"word_count"`
    Timing    float64 `json:"processing_time_ms"`
}

func main() {
    http.HandleFunc("/map", handleMap)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Mapper service starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleMap(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    ctx := context.Background()
    
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req MapperRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Parse S3 URL
    parts := strings.SplitN(strings.TrimPrefix(req.ChunkURL, "s3://"), "/", 2)
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

    // Download chunk from S3
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

    // Count words
    wordCounts := make(map[string]int)
    words := strings.FieldsFunc(string(content), func(r rune) bool {
        return !unicode.IsLetter(r) && !unicode.IsNumber(r)
    })

    for _, word := range words {
        word = strings.ToLower(word)
        if word != "" {
            wordCounts[word]++
        }
    }

    // Convert to JSON
    jsonData, err := json.MarshalIndent(wordCounts, "", "  ")
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
        return
    }

    // Upload result to S3
    resultKey := strings.Replace(key, "chunks/", "mapped/", 1)
    resultKey = strings.Replace(resultKey, ".txt", ".json", 1)
    
    _, err = client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(bucket),
        Key:         aws.String(resultKey),
        Body:        bytes.NewReader(jsonData),
        ContentType: aws.String("application/json"),
    })
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to upload result: %v", err), http.StatusInternalServerError)
        return
    }

    elapsed := time.Since(start).Milliseconds()
    
    resp := MapperResponse{
        ResultURL: fmt.Sprintf("s3://%s/%s", bucket, resultKey),
        WordCount: len(wordCounts),
        Timing:    float64(elapsed),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}