package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type CaseInfo struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Source   string `json:"source"`
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run main.go <csvFilePath> <jsonlFilePath>")
		return
	}

	csvFilePath := os.Args[1]
	jsonlFilePath := os.Args[2]

	// CSV 파일 열기
	file, err := os.Open(csvFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// CSV 리더 생성
	reader := csv.NewReader(file)

	// JSONL 파일 생성
	jsonlFile, err := os.Create(jsonlFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer jsonlFile.Close()

	// CSV 데이터 읽기
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		// record[0]: 사건명, record[1]: 판결요지, record[2]: 사건번호
		caseInfo := CaseInfo{
			Question: record[1],
			Answer:   record[9],
			Source:   record[2],
		}

		// JSONL 형식으로 변환하여 파일에 쓰기
		jsonData, err := json.Marshal(caseInfo)
		if err != nil {
			log.Println("Error marshaling JSON:", err)
			continue
		}
		fmt.Fprintln(jsonlFile, string(jsonData))
	}

	fmt.Println("JSONL 파일 생성 완료!")
}
