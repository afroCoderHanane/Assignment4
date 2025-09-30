package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "sort"
    "strings"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type ReducerRequest struct {
    ResultURLs []string `json:"result_urls"`
}

type ReducerResponse struct {
    FinalURL     string            `json:"final_url"`
    TotalWords   int               `json:"total_words"`
    UniqueWords  int               `json:"unique_words"`
    TopWords     []WordCount       `json:"top_10_words"`
    Timing       float64           `json:"processing_time_ms"`
}

type WordCount struct {
    Word  string `json:"word"`
    Count int    `json:"count"`
}

func main() {
    http.HandleFunc("/reduce", handleReduce)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Reducer service starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleReduce(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    ctx := context.Background()
    
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req ReducerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Create S3 client
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
        return
    }
    client := s3.NewFromConfig(cfg)

    // Aggregate word counts
    finalCounts := make(map[string]int)
    
    for _, resultURL := range req.ResultURLs {
        // Parse S3 URL
        parts := strings.SplitN(strings.TrimPrefix(resultURL, "s3://"), "/", 2)
        if len(parts) != 2 {
            http.Error(w, "Invalid S3 URL format", http.StatusBadRequest)
            return
        }
        bucket, key := parts[0], parts[1]

        // Download result from S3
        result, err := client.GetObject(ctx, &s3.GetObjectInput{
            Bucket: aws.String(bucket),
            Key:    aws.String(key),
        })
        if err != nil {
            http.Error(w, fmt.Sprintf("Failed to get object: %v", err), http.StatusInternalServerError)
            return
        }
        defer result.Body.Close()

        // Parse JSON
        var counts map[string]int
        if err := json.NewDecoder(result.Body).Decode(&counts); err != nil {
            http.Error(w, fmt.Sprintf("Failed to decode JSON: %v", err), http.StatusInternalServerError)
            return
        }

        // Merge counts
        for word, count := range counts {
            finalCounts[word] += count
        }
    }

    // Find top 10 words
    var wordList []WordCount
    totalWords := 0
    for word, count := range finalCounts {
        wordList = append(wordList, WordCount{Word: word, Count: count})
        totalWords += count
    }
    
    sort.Slice(wordList, func(i, j int) bool {
        return wordList[i].Count > wordList[j].Count
    })

    topWords := wordList
    if len(topWords) > 10 {
        topWords = topWords[:10]
    }

    // Prepare final result
    finalResult := map[string]interface{}{
        "word_counts":  finalCounts,
        "total_words":  totalWords,
        "unique_words": len(finalCounts),
        "top_10_words": topWords,
    }

    // Convert to JSON
    jsonData, err := json.MarshalIndent(finalResult, "", "  ")
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
        return
    }

    // Upload final result to S3
    timestamp := time.Now().Unix()
    // Extract bucket from first URL
    parts := strings.SplitN(strings.TrimPrefix(req.ResultURLs[0], "s3://"), "/", 2)
    bucket := parts[0]
    
    finalKey := fmt.Sprintf("results/final-wordcount-%d.json", timestamp)
    _, err = client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(bucket),
        Key:         aws.String(finalKey),
        Body:        bytes.NewReader(jsonData),
        ContentType: aws.String("application/json"),
    })
    if err != nil {
        http.Error(w, fmt.Sprintf("Failed to upload final result: %v", err), http.StatusInternalServerError)
        return
    }

    elapsed := time.Since(start).Milliseconds()
    
    resp := ReducerResponse{
        FinalURL:    fmt.Sprintf("s3://%s/%s", bucket, finalKey),
        TotalWords:  totalWords,
        UniqueWords: len(finalCounts),
        TopWords:    topWords,
        Timing:      float64(elapsed),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}