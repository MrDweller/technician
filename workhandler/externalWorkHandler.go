package workhandler

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	orchestratormodels "github.com/MrDweller/orchestrator-connection/models"
	"github.com/MrDweller/orchestrator-connection/orchestrator"
)

type ExternalWorkHandler struct {
	TakeWorkServiceDefinition string
	OrchestrationConnection   orchestrator.OrchestratorConnection
	SystemDefinition          orchestratormodels.SystemDefinition
	CertificateInfo           orchestratormodels.CertificateInfo
}

func (w *ExternalWorkHandler) AssignWorker(workId string, workerId string) error {
	orchestrationResponse, err := w.OrchestrationConnection.Orchestration(
		w.TakeWorkServiceDefinition,
		[]string{
			"HTTP-SECURE-JSON",
		},
		orchestratormodels.SystemDefinition{
			Address:    w.SystemDefinition.Address,
			Port:       w.SystemDefinition.Port,
			SystemName: w.SystemDefinition.SystemName,
		},
		orchestratormodels.AdditionalParametersArrowhead_4_6_1{
			OrchestrationFlags: map[string]bool{
				"overrideStore": true,
			},
		},
	)
	if err != nil {
		return err
	}

	providers := orchestrationResponse.Response

	if len(providers) <= 0 {
		return fmt.Errorf("found no providers for service: %s", w.TakeWorkServiceDefinition)
	}

	provider := providers[0]

	payload, err := json.Marshal(AssignWorkerDTO{
		WorkId:   workId,
		WorkerId: workerId,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s:%d%s", provider.Provider.Address, provider.Provider.Port, provider.ServiceUri), bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	client, err := w.getClient()
	if err != nil {
		return err
	}

	response, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("error during assignment of worker: %s", err)
	}

	if response.StatusCode != 200 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error during assignment of worker: %s", string(body))
	}

	return nil
}

func (w *ExternalWorkHandler) getClient() (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(w.CertificateInfo.CertFilePath, w.CertificateInfo.KeyFilePath)
	if err != nil {
		return nil, err
	}

	// Load truststore.p12
	truststoreData, err := os.ReadFile(w.CertificateInfo.Truststore)
	if err != nil {
		return nil, err

	}

	// Extract the root certificate(s) from the truststore
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(truststoreData); !ok {
		return nil, err
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{cert},
				RootCAs:            pool,
				InsecureSkipVerify: false,
			},
		},
	}
	return client, nil
}
