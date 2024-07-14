package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

type CaseInfo struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Source   string `json:"source"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <jsonlDirectory>")
		return
	}

	jsonlDirectory := os.Args[1]

	// Redis 클라이언트 설정
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 서버 주소
		Password: "",               // 비밀번호 (없으면 빈 문자열)
		DB:       0,                // 데이터베이스 번호
	})

	// JSONL 파일 읽기
	files, err := os.ReadDir(jsonlDirectory)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".jsonl") {
			filePath := jsonlDirectory + "/" + file.Name()
			processJSONLFile(ctx, client, filePath)
		}
	}

	fmt.Println("데이터 저장 완료!")
}

func processJSONLFile(ctx context.Context, client *redis.Client, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var caseInfo CaseInfo
		if err := json.Unmarshal([]byte(scanner.Text()), &caseInfo); err != nil {
			log.Println("Error unmarshaling JSON:", err)
			continue
		}

		// Redis에 저장
		key := caseInfo.Question
		value := caseInfo.Answer + "\n[출처]:" + caseInfo.Source
		if err := client.Set(ctx, key, value, 0).Err(); err != nil {
			log.Println("Error saving to Redis:", err)
		}
	}
}
