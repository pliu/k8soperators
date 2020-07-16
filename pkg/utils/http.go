package utils

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
)

func SendPost(url string, jsonPayload string) error {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(jsonPayload)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if 300 > resp.StatusCode && resp.StatusCode >= 200 {
		return nil
	}
	return errors.New(resp.Status)
}

func Get(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if 300 > resp.StatusCode && resp.StatusCode >= 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", errors.New(resp.Status)
}
