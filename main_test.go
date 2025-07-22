package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"getha/aparelhos"
	"getha/zoonose"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	createAparelho = "create_aparelho"
	deleteAparelho = "delete_aparelho"
	serveAparelhos = "serve_aparelhos"
	serveImage     = "serve_image"
	createZoonose  = "create_zoonose"
	deleteZoonose  = "delete_zoonose"
	serveZoonoses  = "serve_zoonoses"
	getZoonoseFull = "get_zoonose_full"
)

var client = &http.Client{}

func URL(optional ...string) string {
	endpoint := strings.Join(optional, "/")
	return "http://localhost:8080/" + endpoint
}

func TestApplication(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Aparelhos Suite")
}

func waitForServer(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(URL(""))
		if err == nil {
			if err := resp.Body.Close(); err != nil {
				return err
			}
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.New("server failed to start within timeout")
}

var _ = ginkgo.BeforeSuite(func() {
	aparelhos.AparelhoPath = "aparelhos"
	err := os.Setenv("POSTGRES_USER", "admin")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = os.Setenv("POSTGRES_PASSWORD", "enzofernandes123")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = os.Setenv("POSTGRES_DB", "gethadb")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	err = os.Setenv("POSTGRES_HOST", "127.0.0.1")
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	go main()

	gomega.Expect(waitForServer(10 * time.Second)).NotTo(gomega.HaveOccurred())
})

var _ = ginkgo.AfterSuite(func() {
	// Clear test aparelhos from db
	req, err := http.NewRequest(http.MethodGet, URL(serveAparelhos), nil)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	response, err := client.Do(req)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(response.StatusCode).To(gomega.Equal(http.StatusOK))
	body, err := io.ReadAll(response.Body)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}(response.Body)
	var aparelhos []aparelhos.AparelhoJSON
	gomega.Expect(json.Unmarshal(body, &aparelhos)).NotTo(gomega.HaveOccurred())
	for _, aparelho := range aparelhos {
		req, err = http.NewRequest(
			http.MethodDelete,
			URL(deleteAparelho+"?ID="+aparelho.ID.String()),
			nil,
		)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		res, err := client.Do(req)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(res.StatusCode).To(gomega.Equal(http.StatusOK))
	}

	// Clear test zoonoses from db
	req, err = http.NewRequest(http.MethodGet, URL(serveZoonoses), nil)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	response, err = client.Do(req)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(response.StatusCode).To(gomega.Equal(http.StatusOK))
	body, err = io.ReadAll(response.Body)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	var zoonoseCards []zoonose.Card
	gomega.Expect(json.Unmarshal(body, &zoonoseCards)).NotTo(gomega.HaveOccurred())
	for _, zoonoseCard := range zoonoseCards {
		req, err = http.NewRequest(
			http.MethodDelete,
			URL(deleteZoonose+"?ID="+zoonoseCard.ID.String()),
			nil,
		)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		res, err := client.Do(req)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(res.StatusCode).To(gomega.Equal(http.StatusNoContent))
	}
})

var _ = ginkgo.Describe("server tests", func() {
	ginkgo.It("add aparelhos", func() {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		err := writer.WriteField("nome", "Teste")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		addFile(writer, "image_path", "tests/aparelhos/img.jpg")
		addFile(writer, "manual_path", "tests/aparelhos/man.pdf")
		addFile(writer, "video_path", "tests/aparelhos/vid.mp4")
		err = writer.Close()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		req, err := http.NewRequest("POST", URL(createAparelho), body)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := client.Do(req)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		respBody, err := io.ReadAll(resp.Body)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		resJSON := make(map[string]interface{})
		gomega.Expect(json.Unmarshal(respBody, &resJSON)).NotTo(gomega.HaveOccurred())
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusCreated))
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		}(resp.Body)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		response, err := http.Get(URL(serveImage + "?ID=" + resJSON["id"].(string)))
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(response.StatusCode).To(gomega.Equal(http.StatusOK))
	})
	ginkgo.It("add zoonose", func() {
		zoonoseMock := zoonose.JSONZoonose{
			Nome:           "Rabies",
			NomeCientifico: "Lyssavirus",
			Descricao:      "Descrição",
			Organismo:      "Protozoario",
			Agentes:        []string{"Contact with infected animal"},
			Vetores:        []string{"Dogs", "Bats"},
			Transmissoes:   []string{"Animal bite"},
			Profilaxias:    []string{"Vaccination"},
			Diagnosticos:   []string{"Fever", "Hydrophobia"},
			Regioes:        []string{"Bahia", "Paraíba"},
		}
		zoonoseMockJSON, err := json.Marshal(zoonoseMock)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		createReq, err := http.NewRequest("POST", URL(createZoonose), bytes.NewBuffer(zoonoseMockJSON))
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		createReq.Header.Set("Content-Type", "application/json; charset=utf-8")
		createResp, err := client.Do(createReq)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		createRespBody, err := io.ReadAll(createResp.Body)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		createRespJSON := make(map[string]interface{})
		gomega.Expect(json.Unmarshal(createRespBody, &createRespJSON)).NotTo(gomega.HaveOccurred())
		gomega.Expect(createResp.StatusCode).To(gomega.Equal(http.StatusCreated))
		zoonoseFullReq, err := http.NewRequest(http.MethodGet, URL(getZoonoseFull+"?ID="+createRespJSON["id"].(string)), nil)
		zoonoseFullResp, err := client.Do(zoonoseFullReq)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		zoonoseFullBody, err := io.ReadAll(zoonoseFullResp.Body)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(zoonoseFullResp.StatusCode).To(gomega.Equal(http.StatusOK))
		gomega.Expect(zoonoseMockJSON).To(gomega.Equal(zoonoseFullBody))
	})
})

func addFile(writer *multipart.Writer, fieldName, filePath string) {
	file, err := os.Open(filePath)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	defer func(file *os.File) {
		gomega.Expect(file.Close()).Should(gomega.Succeed())
	}(file)
	part, err := writer.CreateFormFile(fieldName, file.Name())
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	_, err = io.Copy(part, file)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}
