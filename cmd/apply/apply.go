package apply

import (
	"context"
	"fmt"
	"github.com/kfsoftware/statuspage/config"
	"github.com/kfsoftware/statuspage/pkg/graphql/models"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/yaml"
	"net/http"
	"os"
)

type applyCmd struct {
	file      string
	folder    string
	gqlClient *graphql.Client
}

func (c applyCmd) validate() error {
	return nil
}

func readFile(f string) ([]byte, error) {
	fileBytes, err := ioutil.ReadFile(f)
	return fileBytes, err
}

type StatusKind string

const (
	HttpHealthCheck StatusKind = "HttpHealthCheck"
	TLSHealthCheck  StatusKind = "TLSHealthCheck"
	StatusPageKind  StatusKind = "StatusPage"
)

func (c applyCmd) applyFile(filePath string) error {
	fileBytes, err := readFile(filePath)
	if err != nil {
		return err
	}
	var initialUnmarshall struct {
		Kind StatusKind `yaml:"kind"`
	}
	err = yaml.Unmarshal(fileBytes, &initialUnmarshall)
	if err != nil {
		return err
	}
	log.Debugf("Kind: %s", initialUnmarshall.Kind)
	switch initialUnmarshall.Kind {
	case HttpHealthCheck:
		return c.applyHttpHealthCheck(fileBytes)
	case TLSHealthCheck:
		return c.applyTLSHealthCheck(fileBytes)
	case StatusPageKind:
		return c.applyStatusPage(fileBytes)
	default:
		return errors.Errorf("Unknown kind: %s", initialUnmarshall.Kind)
	}
	return nil
}
func (c applyCmd) run() error {
	if c.file != "" {
		return c.applyFile(c.file)
	} else if c.folder != "" {
		files, err := ioutil.ReadDir(c.folder)
		if err != nil {
			return err
		}
		var checkFiles []string
		var statusPageFiles []string
		for _, file := range files {
			fileBytes, err := readFile(fmt.Sprintf("%s/%s", c.folder, file.Name()))
			if err != nil {
				return err
			}
			var initialUnmarshall struct {
				Kind StatusKind `yaml:"kind"`
			}
			err = yaml.Unmarshal(fileBytes, &initialUnmarshall)
			if err != nil {
				return err
			}
			log.Debugf("Kind: %s", initialUnmarshall.Kind)
			switch initialUnmarshall.Kind {
			case HttpHealthCheck:
				checkFiles = append(checkFiles, fmt.Sprintf("%s/%s", c.folder, file.Name()))
			case TLSHealthCheck:
				checkFiles = append(checkFiles, fmt.Sprintf("%s/%s", c.folder, file.Name()))
			case StatusPageKind:
				statusPageFiles = append(statusPageFiles, fmt.Sprintf("%s/%s", c.folder, file.Name()))
			default:
				return errors.Errorf("Unknown kind: %s", initialUnmarshall.Kind)
			}
		}
		for _, file := range checkFiles {
			err = c.applyFile(file)
			if err != nil {
				return err
			}
			log.Infof("File %s applied", file)
		}
		for _, file := range statusPageFiles {
			err = c.applyFile(file)
			if err != nil {
				return err
			}
			log.Infof("File %s applied", file)
		}
	}
	return nil
}

type CreateHttpCheckInput struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Frecuency  string `json:"frecuency"`
	URL        string `json:"url"`
	StatusCode int    `json:"statusCode"`
}

func (c applyCmd) applyHttpHealthCheck(bytes []byte) error {
	httpCheck := &config.HttpHealthCheck{}
	err := yaml.Unmarshal(bytes, httpCheck)
	if err != nil {
		return err
	}
	log.Debugf("%+v", httpCheck)
	createHttpCheckInput := CreateHttpCheckInput{
		Name:       httpCheck.Name,
		Namespace:  httpCheck.Namespace,
		Frecuency:  httpCheck.Spec.Frequency,
		URL:        httpCheck.Spec.URL,
		StatusCode: httpCheck.Spec.StatusCode,
	}
	var m struct {
		Check struct {
			Id          string `graphql:"id"`
			Name        string `graphql:"name"`
			Namespace   string `graphql:"namespace"`
			Frecuency   string `graphql:"frecuency"`
			Status      string `graphql:"status"`
			Latestcheck string `graphql:"latestCheck"`
			Message     string `graphql:"message"`
			Errormsg    string `graphql:"errorMsg"`
		} `graphql:"createHttpCheck(input: $input)"`
	}
	vars := map[string]interface{}{
		"input": createHttpCheckInput,
	}
	ctx := context.Background()
	err = c.gqlClient.Mutate(ctx, &m, vars)
	if err != nil {
		return err
	}
	return nil
}

type CreateTlsCheckInput struct {
	Name      string  `json:"name"`
	Namespace string  `json:"namespace"`
	Frecuency string  `json:"frecuency"`
	Address   string  `json:"address"`
	RootCAs   *string `json:"rootCAs"`
}

func (c applyCmd) applyTLSHealthCheck(bytes []byte) error {
	tlsCheck := &config.TLSHealthCheck{}
	err := yaml.Unmarshal(bytes, tlsCheck)
	if err != nil {
		return err
	}
	log.Debugf("%+v", tlsCheck)
	createTlsCheckInput := CreateTlsCheckInput{
		Name:      tlsCheck.Name,
		Namespace: tlsCheck.Namespace,
		Frecuency: tlsCheck.Spec.Frequency,
		Address:   fmt.Sprintf("%s:%d", tlsCheck.Spec.Host, tlsCheck.Spec.Port),
		RootCAs:   tlsCheck.Spec.RootCAs,
	}
	var m struct {
		Check struct {
			Id          string `graphql:"id"`
			Name        string `graphql:"name"`
			Namespace   string `graphql:"namespace"`
			Frecuency   string `graphql:"frecuency"`
			Status      string `graphql:"status"`
			Latestcheck string `graphql:"latestCheck"`
			Message     string `graphql:"message"`
			Errormsg    string `graphql:"errorMsg"`
		} `graphql:"createTlsCheck(input: $input)"`
	}
	vars := map[string]interface{}{
		"input": createTlsCheckInput,
	}
	ctx := context.Background()
	err = c.gqlClient.Mutate(ctx, &m, vars)
	if err != nil {
		return err
	}
	return nil
}

func (c applyCmd) applyStatusPage(fileBytes []byte) error {
	statusPage := &config.StatusPage{}
	err := yaml.Unmarshal(fileBytes, statusPage)
	if err != nil {
		return err
	}
	log.Debugf("%+v", statusPage)
	createStatusPageInput := models.CreateStatusPageInput{
		Name:       statusPage.Name,
		Namespace:  statusPage.Namespace,
		Title:      statusPage.Spec.Title,
		CheckSlugs: statusPage.Spec.Services,
	}
	var m struct {
		StatusPage struct {
			Id        string `graphql:"id"`
			Name      string `graphql:"name"`
			Namespace string `graphql:"namespace"`
			Title     string `graphql:"title"`
		} `graphql:"createStatusPage(input: $input)"`
	}
	vars := map[string]interface{}{
		"input": createStatusPageInput,
	}
	ctx := context.Background()
	err = c.gqlClient.Mutate(ctx, &m, vars)
	if err != nil {
		return err
	}
	return nil
}
func GetGraphqlClient(ctx context.Context, url string) *graphql.Client {
	accessToken := os.Getenv("ACCESS_TOKEN")
	var httpClient *http.Client
	if accessToken != "" {
		src := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: accessToken},
		)
		httpClient = oauth2.NewClient(ctx, src)
	}
	gqlClient := graphql.NewClient(url, httpClient)
	return gqlClient
}

func NewApplyCMD() *cobra.Command {
	c := applyCmd{}
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "apply",
		Long:  `apply`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := c.validate()
			if err != nil {
				return err
			}
			ctx := context.Background()
			gqlClient := GetGraphqlClient(ctx, "http://localhost:8888/graphql")
			c.gqlClient = gqlClient
			return c.run()
		},
	}
	f := cmd.Flags()
	f.StringVarP(&c.file, "file", "f", "", "file to apply")
	f.StringVarP(&c.folder, "folder", "k", "", "folder where the checks are located")
	return cmd
}
