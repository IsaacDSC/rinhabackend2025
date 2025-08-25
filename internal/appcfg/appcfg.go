package appcfg

import (
	"fmt"
	"os"
	"reflect"
)

type Environment struct {
	AppPort              string
	GQueueUrl            string
	RedisUrl             string
	DatabaseUrl          string
	ProcessorDefaultUrl  string
	ProcessorFallbackUrl string
}

var environment Environment

func (e Environment) Validate() error {
	v := reflect.ValueOf(e)
	t := reflect.TypeOf(e)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := t.Field(i).Name

		if field.String() == "" {
			return fmt.Errorf("campo %s está vazio", fieldName)
		}
	}

	return nil
}

func init() {
	environment = Environment{
		GQueueUrl:            os.Getenv("GQUEUE_URL"),
		AppPort:              os.Getenv("APP_PORT"),
		RedisUrl:             os.Getenv("REDIS_URL"),
		DatabaseUrl:          os.Getenv("DATABASE_URL"),
		ProcessorDefaultUrl:  os.Getenv("PROCESSOR_DEFAULT_URL"),
		ProcessorFallbackUrl: os.Getenv("PROCESSOR_FALLBACK_URL"),
	}

	if err := environment.Validate(); err != nil {
		panic(fmt.Sprintf("Erro de configuração: %v", err))
	}
}

func Get() Environment {
	return environment
}
