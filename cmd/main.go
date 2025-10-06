package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func main() {
	now := time.Now().UTC()
	j, err := json.Marshal(now)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(j))
}
