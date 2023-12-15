package capabilities

import (
	"fmt"
	"os"
)

func SaveFile(payload []byte, filepath string) {
	f, err := os.Create("./savedfiles/" + filepath)
	defer f.Close()
	if err == nil {
		f.Write(payload)
	} else {
		fmt.Println("save_file err", err)
	}
}
