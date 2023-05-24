package seeds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kneerunjun/botmincock/biz"
)

func Accounts(path string) (error, []biz.UserAccount) {
	f, err := os.Open(path)
	if err != nil || f == nil {
		return fmt.Errorf("failed to open file %s, check if file exists", path), nil
	}
	byt, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read contents of json file %s: %s", path, err), nil
	}
	result := struct {
		Data []biz.UserAccount `json:"data"`
	}{}
	if err := json.Unmarshal(byt, &result); err != nil {
		return fmt.Errorf("failed to unmarshal seeding data from %s: %s", path, err), nil
	}
	return nil, result.Data
}

func Estimates(path string) (error, []biz.Estimate) {
	f, err := os.Open(path)
	if err != nil || f == nil {
		return fmt.Errorf("failed to open file %s, check if file exists", path), nil
	}
	byt, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read contents of json file %s: %s", path, err), nil
	}
	result := struct {
		Data []biz.Estimate `json:"data"`
	}{}
	if err := json.Unmarshal(byt, &result); err != nil {
		return fmt.Errorf("failed to unmarshal seeding data from %s: %s", path, err), nil
	}
	return nil, result.Data
}

func Expenses(path string) (error, []biz.Expense) {
	f, err := os.Open(path)
	if err != nil || f == nil {
		return fmt.Errorf("failed to open file %s, check if file exists", path), nil
	}
	byt, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read contents of json file %s: %s", path, err), nil
	}
	result := struct {
		Data []biz.Expense `json:"data"`
	}{}
	if err := json.Unmarshal(byt, &result); err != nil {
		return fmt.Errorf("failed to unmarshal seeding data from %s: %s", path, err), nil
	}
	return nil, result.Data
}
func Transactions(path string) (error, []biz.Transac) {
	f, err := os.Open(path)
	if err != nil || f == nil {
		return fmt.Errorf("failed to open file %s, check if file exists", path), nil
	}
	byt, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read contents of json file %s: %s", path, err), nil
	}
	result := struct {
		Data []biz.Transac `json:"data"`
	}{}
	if err := json.Unmarshal(byt, &result); err != nil {
		return fmt.Errorf("failed to unmarshal seeding data from %s: %s", path, err), nil
	}
	return nil, result.Data
}
