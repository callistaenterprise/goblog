package config

import (
        "net/http"
        "fmt"
        "io/ioutil"
        "encoding/json"
        "github.com/spf13/viper"
)


// Loads config from for example http://configserver:8888/accountservice/test/P8
func LoadConfigurationFromBranch(configServerUrl string, appName string, profile string, branch string) {
        url := fmt.Sprintf("%s/%s/%s/%s", configServerUrl, appName, profile, branch)
        fmt.Printf("Loading config from %s\n", url)
        body, err := fetchConfiguration(url)
        if err != nil {
                panic("Couldn't load configuration, cannot start. Terminating. Error: " + err.Error())
        }
        parseConfiguration(body)
}

func fetchConfiguration(url string) ([]byte, error) {
        resp, err := http.Get(url)
        if err != nil {
                panic("Couldn't load configuration, cannot start. Terminating. Error: " + err.Error())
        }
        body, err := ioutil.ReadAll(resp.Body)
        return body, err
}

func parseConfiguration(body []byte) {
        var cloudConfig springCloudConfig
        err := json.Unmarshal(body, &cloudConfig)
        if err != nil {
                panic("Cannot parse configuration, message: " + err.Error())
        }

        for key, value := range cloudConfig.PropertySources[0].Source {
                viper.Set(key, value)
                fmt.Printf("Loading config property %v => %v\n", key, value)
        }
        if viper.IsSet("server_name") {
                fmt.Printf("Successfully loaded configuration for service %s\n", viper.GetString("server_name"))
        }
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

