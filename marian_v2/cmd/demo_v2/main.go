package main

import (
	"fmt"
	"log"

	"github.com/techwithsergiu/marian_tokenizer_go/marian_v2"
)

func main() {

	fmt.Println("\n>> use_marian_v2:")
	// ===

	tok, err := marian_v2.NewTokenizer("./models/opus-mt-ru-en")
	if err != nil {
		log.Fatalf("NewTokenizer: %v", err)
	}
	defer tok.Close()

	// ===

	text := "Привет, как у тебя дела?"
	ids, err := tok.Encode(text, true)
	if err != nil {
		log.Fatalf("Encode: %v", err)
	}
	fmt.Println("text:", text)
	fmt.Println("ids :", ids)

	// ===

	inputIDs, attn, err := tok.EncodeBatch([]string{
		"Привет, как у тебя дела?",
		"Это тестовая строка для проверки.",
	})
	if err != nil {
		log.Fatalf("EncodeBatch: %v", err)
	}
	fmt.Println("input_ids:", inputIDs)
	fmt.Println("attention:", attn)

	// ===

	generated_ids := []int64{62517, 160, 200, 2, 508, 55, 33, 19, 0}
	decoded, err := tok.Decode(generated_ids, true)
	if err != nil {
		log.Fatalf("Decode: %v", err)
	}
	fmt.Println("decoded:", decoded)

	// ===
	fmt.Println("")

}
