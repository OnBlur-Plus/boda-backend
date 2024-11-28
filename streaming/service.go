package main

import (
	"context"
	"net/http"
	"sync"
)

type HttpService interface {
	Close() error
	Run(ctx context.Context) error
}

func NewHTTPService() HttpService {
	return &httpService{}
}

type httpService struct {
	servers []*http.Server
}

func (v *httpService) Close() error {
	v.servers = nil

	var wg sync.WaitGroup
	defer wg.Wait()

	return nil
}

func (v *httpService) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	return nil
}