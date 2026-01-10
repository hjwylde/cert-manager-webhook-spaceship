package acme

import (
	"context"
	"strings"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook"
	acme "github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/hjwylde/cert-manager-webhook-spaceship/internal/spaceship"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type solver struct {
	kclient *kubernetes.Clientset
}

func NewSolver() webhook.Solver {
	return &solver{}
}

func (s *solver) Name() string {
	return "spaceship"
}

func (s *solver) Present(ch *acme.ChallengeRequest) error {
	klog.Infof("Presenting ACME challenge: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)

	c, err := s.newClient(ch)
	if err != nil {
		klog.Errorf("error creating client: %v", err)
		return err
	}

	domain := strings.TrimSuffix(ch.ResolvedZone, ".")
	name := strings.TrimSuffix(ch.ResolvedFQDN, "."+ch.ResolvedZone)

	item := spaceship.NewDNSRecordsListTxtItem(name, ch.Key)
	item.TTL = 60

	_, err = c.DNSRecords.Put(context.TODO(), domain, spaceship.DNSRecords{
		Items: []spaceship.DNSRecordsListTxtItem{item},
	})
	if err != nil {
		klog.Errorf("request error: %v", err)
		return err
	}

	return nil
}

func (s *solver) CleanUp(ch *acme.ChallengeRequest) error {
	klog.Infof("Cleaning up ACME challenge: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)

	c, err := s.newClient(ch)
	if err != nil {
		klog.Errorf("error creating client: %v", err)
		return err
	}

	domain := strings.TrimSuffix(ch.ResolvedZone, ".")
	name := strings.TrimSuffix(ch.ResolvedFQDN, "."+ch.ResolvedZone)

	item := spaceship.NewDNSRecordsListTxtItem(name, ch.Key)

	_, err = c.DNSRecords.Delete(context.TODO(), domain, []spaceship.DNSRecordsListTxtItem{item})
	if err != nil {
		klog.Errorf("request error: %v", err)
		return err
	}

	return nil
}

func (s *solver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	c, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	s.kclient = c

	return nil
}

func (s *solver) newClient(ch *acme.ChallengeRequest) (*spaceship.Client, error) {
	c, err := spaceship.NewClient()
	if err != nil {
		return nil, err
	}

	cred, err := s.loadCredentials(ch)
	if err != nil {
		return nil, err
	}

	c.ApiKey = cred.apiKey
	c.ApiSecret = cred.apiSecret

	return c, nil
}
