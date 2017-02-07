package config

import (
        "net/http"
        "fmt"
        "io/ioutil"
        "encoding/json"
        "github.com/spf13/viper"
)

func LoadConfiguration(configServerUrl string, appName string, profile string) {
        body, err := fetchConfiguration(fmt.Sprintf("%s/%s-%s/%s", configServerUrl, appName, profile, profile))
        if err != nil {
                panic("Couldn't load configuration, cannot start. Terminating. Error: " + err.Error())
        }
        parseConfiguration(body)
}

func parseConfiguration(body []byte) {
        var cloudConfig springCloudConfig
        json.Unmarshal(body, &cloudConfig)

        for key, value := range cloudConfig.PropertySources[0].Source {
                viper.Set(key, value)
        }
}

func fetchConfiguration(url string) ([]byte, error) {
        resp, err := http.Get(url)
        if err != nil {
                panic("Couldn't load configuration, cannot start. Terminating. Error: " + err.Error())
        }
        body, err := ioutil.ReadAll(resp.Body)
        return body, err
}

type springCloudConfig struct {
        Name            string           `json:"name"`
        Profiles        []string         `json:"profiles"`
        Label           string           `json:"label"`
        Version         string           `json:"version"`
        PropertySources []propertySource `json:"propertySources"`
}

type propertySource struct {
        Name   string                 `json:"name"`
        Source map[string]interface{} `json:"source"`
}

