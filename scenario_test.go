package goatt

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"testing"
)

func TestEvaluate(t *testing.T) {
	output := bytes.Buffer{}
	vars := map[string]string{
		"src":  "111",
		"type": "222",
		"id":   "333",
	}
	msg := `{
        "id": "b050fff1-ff9d-4dbf-8889-1ad5bd9c6df5",
        "errors": [],
        "src": "$(src)",
        "type": "$(type)",
        "body": {
          "id": "$(id)",
          "first_name": "Abi Abu",
          "last_name": "Abe",
          "image": "avatar.jpg",
          "service": "deluxe"
        }
      }`

	data := msg
	for {
		prefixIdx := strings.Index(data, "$(")
		suffixIdx := strings.Index(data, ")")
		if prefixIdx == -1 || suffixIdx == -1 || prefixIdx > suffixIdx {
			fmt.Fprintf(&output, data)
			break
		}
		prefix := data[:prefixIdx]
		varName := data[prefixIdx+2 : suffixIdx]
		suffix := data[suffixIdx+1:]

		value, _ := vars[varName]
		fmt.Fprintf(&output, prefix)
		fmt.Fprintf(&output, value)
		//fmt.Fprintf(&output, data)
		// fmt.Fprintf(os.Stderr, "prefix [%s]\n", prefix)
		// fmt.Fprintf(os.Stderr, "var %s=%s (%s)\n", varName, value, ok)
		// fmt.Fprintf(os.Stderr, "suffix [%s]\n", suffix)
		data = suffix
	}
	fmt.Fprintf(os.Stderr, "%s\n", output.String())

	/*
		subs := strings.SplitAfterN(msg, "$(", -1)
		for i, s := range subs {
			fmt.Fprintf(os.Stderr, "[%03d] [%s]\n", i, s)
			if suffix := strings.Index(s, ")"); suffix != -1 {
				v := s[:suffix]
				fmt.Fprintf(os.Stderr, "var %s=%s (%s)\n", v, value, ok)
				fmt.Fprintf(&output, "%s")
			} else {
				fmt.Fprintf(&output, s)
			}
		}
	*/
}
