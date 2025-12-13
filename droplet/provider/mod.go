package provider

import (
	"github.com/codecrafter404/kairos-re-unlock/droplet/config"
	"github.com/mudler/go-pluggable"
)

func GetResponse(config config.Config) pluggable.EventResponse {

	datachan := make(chan pluggable.EventResponse)

	go getAsyncNodePairResponse(config, datachan)
	go getAsyncHttpsResponse(config, datachan)

	return <-datachan
}
