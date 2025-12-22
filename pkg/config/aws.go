package config

import (
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/creasty/defaults"
)

type IEndpoint struct {
	Endpoint string `yaml:"endpoint"`
}

type AWS struct {
	AccessKey string `yaml:"accessKey"`
	SecretKet string `yaml:"secretKey"`
	Region    string `yaml:"region" default:"ap-northeast-1"`
	SNS       SNS    `yaml:"sns"`
}

func (conf *AWS) UnmarshalYAML(unmarshal func(any) error) error {
	if err := defaults.Set(conf); err != nil {
		return err
	}
	type plain AWS
	if err := unmarshal((*plain)(conf)); err != nil {
		return err
	}
	return nil
}

func (conf *AWS) AWSConfig() []func(*awsConfig.LoadOptions) error {
	config := []func(*awsConfig.LoadOptions) error{}
	if conf.AccessKey != "" && conf.SecretKet != "" {
		config = append(config, awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(conf.AccessKey, conf.SecretKet, ""),
		))
	}
	if conf.Region != "" {
		config = append(config, awsConfig.WithRegion(conf.Region))
	}
	return config
}

func (conf *AWS) WhitelistResources() []string {
	return conf.SNS.Resources.GetWhitelist()
}

func (conf *AWS) BlacklistResources() []string {
	return conf.SNS.Resources.GetBlacklist()
}

func (conf *AWS) SkipEvents() []string {
	return conf.SNS.Resources.GetSkipEvents()
}
