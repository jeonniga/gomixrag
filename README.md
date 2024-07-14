GO언어로 제작된 Ollama, REDIS를 기반으로 하는 RAG 프로그램입니다.
1. dataset 폴더에 *.csv 파일을 준비
2. eda 프로그램을 실행하면 csv -> jsonl
    ./eda/eda 입력할_csv_파일 출력할_jsonl_파일
    ./eda/eda data1.csv data1.jsonl 
3. loader 프로그램을 실행하면 jsonl 파일 -> Redis에 load
    ./loader/loader jsonl이 있는 디렉토리
    ./loader/loader dataset
4. ./mix 프로그램 실행해 inference
    go run mix.go --model gemma:2b
    go run mix.go --model gemma:2b --jsonl dbpedia_sample.jsonl