package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

var llm llms.LLM

func initLLM(modelName string) {
	var err error
	llm, err = ollama.New(ollama.WithModel(modelName))
	if err != nil {
		log.Fatal(err)
	}
}

func readJSONLFile(filename string) ([]map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data []map[string]string
	decoder := json.NewDecoder(file)
	for decoder.More() {
		var entry map[string]string
		if err := decoder.Decode(&entry); err != nil {
			return nil, err
		}
		data = append(data, entry)
	}

	return data, nil
}

func storeInRedis(data []map[string]string) error {
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Error connecting to Redis:", err)
		return err
	}
	defer conn.Close()
	for _, entry := range data {
		question := entry["question"]
		answer := entry["answer"]
		source := entry["source"]
		answerWithSource := answer + "\n[출처]:" + source

		_, err := conn.Do("SET", question, answerWithSource)
		if err != nil {
			fmt.Println("Error storing data in Redis:", err)
			return err
		}
	}
	return nil
}

// LevenshteinDistance calculates the Levenshtein distance between two strings.
func LevenshteinDistance(s1, s2 string) int {
	m, n := len(s1), len(s2)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
		dp[i][0] = i
	}
	for j := 1; j <= n; j++ {
		dp[0][j] = j
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}
			dp[i][j] = min(dp[i-1][j]+1, dp[i][j-1]+1, dp[i-1][j-1]+cost)
		}
	}
	return dp[m][n]
}

func min(a, b, c int) int {
	if a < b {
		return a
	}
	if b < c {
		return b
	}
	return c
}

func getSimilarValue(pattern string) (string, string) {
	// Connect to the Redis server
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatal("Error connecting to Redis:", err)
		return "NA", "Error"
	}
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", "*"+pattern+"*"))
	if err != nil {
		log.Fatal("Error fetching keys:", err)
		return "NA", "Error"
	}

	// Initialize variables to store the most similar key and its value
	mostSimilarKey := "NA"
	var mostSimilarValue string
	lowdistance := -1 // Set initial value to -1 indicating no distance has been calculated yet

	// Calculate Levenshtein similarity
	for _, key := range keys {
		fmt.Println("KEY:", strings.ReplaceAll(key, " ", ""), ", PATTERN:", strings.ReplaceAll(pattern, " ", ""))
		distance := LevenshteinDistance(strings.ReplaceAll(key, " ", ""), strings.ReplaceAll(pattern, " ", ""))
		if lowdistance == -1 || distance < lowdistance { // Update when a lower distance is found or on the first calculation
			lowdistance = distance
			mostSimilarKey = key
			// Get the value for the most similar key
			value, err := redis.String(conn.Do("GET", key))
			if err != nil {
				log.Printf("Error fetching value for key %s: %v", key, err)
				return "NA", "Error"
			}
			mostSimilarValue = value
		}
	}

	return mostSimilarKey, mostSimilarValue
}

func generateResponse(prompt string) (string, float64, error) {
	ctx := context.Background()
	key, value := getSimilarValue(prompt)

	fmt.Println("===================================================")
	fmt.Println("PROMPT:", prompt, "\nKEY:", key, "\nValue:", value)
	startTime := time.Now()
	var completion string
	var err error
	if key == "NA" {
		query := "당신은 질문에 답하는 AI 비서입니다.\n"
		query += prompt
		fmt.Println("LLM에 요청하는 내용:", query)
		completion, err = llms.GenerateFromSinglePrompt(ctx, llm, query)
	} else {
		completion = key + "\n" + value
	}
	executionTime := time.Since(startTime).Seconds()

	// Replace newline characters with <br> tags
	completion = strings.Replace(completion, "\n[출처]:", "<br>[출처] : ", -1)
	fmt.Println("Response:", completion)

	return completion, executionTime, err
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		prompt := r.FormValue("prompt")
		if prompt == "" {
			http.Error(w, "prompt is required", http.StatusBadRequest)
			return
		}

		completion, executionTime, err := generateResponse(prompt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := struct {
			Prompt        string  `json:"prompt"`
			Completion    string  `json:"completion"`
			ExecutionTime float64 `json:"execution_time"`
		}{
			Prompt:        prompt,
			Completion:    completion,
			ExecutionTime: executionTime,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
		return
	}

	http.ServeFile(w, r, "index.html")
}

func main() {
	modelName := flag.String("model", "", "ollama model name")
	jsonlFile := flag.String("jsonl", "", "path to JSONL file")
	flag.Parse()

	if *modelName == "" {
		log.Fatal("model name is required")
	}

	initLLM(*modelName)

	if *jsonlFile != "" {
		data, err := readJSONLFile(*jsonlFile)
		if err != nil {
			log.Fatal(err)
		}

		if err := storeInRedis(data); err != nil {
			log.Fatal(err)
		}
	}

	http.HandleFunc("/", handler)
	fmt.Println("서버가 http://localhost:8510 에서 시작되었습니다.")
	log.Fatal(http.ListenAndServe(":8510", nil))
}

// go run mix.go --model gemma:2b
// go run mix.go --model gemma:2b --jsonl dbpedia_sample.jsonl
