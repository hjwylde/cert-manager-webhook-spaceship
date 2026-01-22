package acme

import (
	"context"
	"encoding/json"
	"fmt"

	acme "github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	core "k8s.io/api/core/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type config struct {
	ApiKeySecretRef    core.SecretKeySelector `json:"apiKeySecretRef"`
	ApiSecretSecretRef core.SecretKeySelector `json:"apiSecretSecretRef"`
}

func (s *solver) loadConfig(cfgJSON *apiextensions.JSON) (*config, error) {
	cfg := config{}
	if cfgJSON == nil {
		return &cfg, nil
	}

	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}

	return &cfg, nil
}

type credentials struct {
	apiKey    string
	apiSecret string
}

func (s *solver) loadCredentials(ch *acme.ChallengeRequest) (*credentials, error) {
	cfg, err := s.loadConfig(ch.Config)
	if err != nil {
		return nil, err
	}

	apiKey, err := s.resolveSecretRef(cfg.ApiKeySecretRef, ch.ResourceNamespace)
	if err != nil {
		return nil, err
	}

	apiSecret, err := s.resolveSecretRef(cfg.ApiSecretSecretRef, ch.ResourceNamespace)
	if err != nil {
		return nil, err
	}

	return &credentials{apiKey: apiKey, apiSecret: apiSecret}, nil
}

func (s *solver) resolveSecretRef(selector core.SecretKeySelector, namespace string) (string, error) {
	secret, err := s.kclient.CoreV1().Secrets(namespace).Get(context.TODO(), selector.Name, meta.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("error getting secret %q %q: %v", namespace, selector.Name, err)
	}

	b, ok := secret.Data[selector.Key]
	if !ok {
		return "", fmt.Errorf("secret %q %q does not contain key %q", namespace, selector.Name, selector.Key)
	}

	return string(b), nil
}
