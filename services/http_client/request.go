package http_client

import (
	"fmt"
	"gopkg.in/resty.v1"
)

func PostJson(url string, data interface{}) error {

	client := resty.New()

	var result map[string]interface{}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		SetResult(&result).
		Post(url)

	if err != nil {
		return fmt.Errorf("post json err: %v", err)
	}

	fmt.Printf("status code: %d\n", resp.StatusCode())
	fmt.Printf("response: %+v\n", result)

	return nil
}
