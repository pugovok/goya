package main_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/pugovok/goya/app"
)

func TestAppRun(t *testing.T) {
	var server app.Server
	err := server.LoadConfig("../../config")
	if err != nil {
		panic(err)
	}

	err = server.InitLogger()
	if err != nil {
		fmt.Println(err)
	}

	ctx := context.Background()
	go server.Run(ctx)

	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		t.Error(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Error(fmt.Errorf("wrong http status: %s", resp.Status))
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	bodyString := string(bodyBytes)
	if bodyString != "Hello world" {
		t.Error(fmt.Errorf("wrong http message: %s", bodyString))
	}

	err = server.Stop(ctx)
	if err != nil {
		t.Error(err)
	}
}
