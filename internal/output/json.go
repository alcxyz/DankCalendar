package output

import (
	"encoding/json"
	"fmt"
)

func JSON(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		Error(fmt.Sprintf("json marshal: %v", err))
		return
	}
	fmt.Println(string(data))
}

func Error(msg string) {
	data, _ := json.Marshal(map[string]string{"error": msg})
	fmt.Println(string(data))
}
