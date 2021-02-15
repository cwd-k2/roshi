package fil_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/cwd-k2/roshi/pkg/roka"
)

func TestCreateFiltrations(t *testing.T) {
	s := "contents/aosij00rm_kliajr-iur/main.cpp"
	b := []byte(`{"contents/#2-#1/main.cpp": "#2/#1.cpp"}`)

	var obj map[string]string
	if err := json.Unmarshal(b, &obj); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	filtrations, err := roka.CreateFiltrations(obj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	for _, filtration := range filtrations {
		t.Logf("%+v\n", filtration)
		pat := roka.CreateMatchingRegexp(filtration.OriginPattern)
		str := roka.CreateTemplateString(filtration.DerivePattern, filtration.Numberings)
		replaced := pat.ReplaceAllString(s, str)
		t.Log(replaced)
		if replaced != "aosij00rm_kliajr/iur.cpp" {
			t.Fail()
		}
	}
}
