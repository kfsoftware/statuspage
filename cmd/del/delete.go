package del

import (
	"context"
	"fmt"
	"github.com/kfsoftware/statuspage/cmd/cmdutils"
	"github.com/kfsoftware/statuspage/config"
	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type deleteCmd struct {
	gqlClient *graphql.Client
	file      string
	folder    string
}

func (c deleteCmd) validate() error {
	return nil
}

func (c deleteCmd) run() error {
	if c.file != "" {
		return c.deleteFile(c.file)
	} else if c.folder != "" {
		files, err := ioutil.ReadDir(c.folder)
		if err != nil {
			return err
		}
		var checkFiles []string
		var statusPageFiles []string
		for _, file := range files {
			fileBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", c.folder, file.Name()))
			if err != nil {
				return err
			}
			statusKind, err := cmdutils.GetFileType(fileBytes)
			if err != nil {
				return err
			}
			switch statusKind {
			case cmdutils.HttpHealthCheck:
				checkFiles = append(checkFiles, fmt.Sprintf("%s/%s", c.folder, file.Name()))
			case cmdutils.TLSHealthCheck:
				checkFiles = append(checkFiles, fmt.Sprintf("%s/%s", c.folder, file.Name()))
			case cmdutils.StatusPageKind:
				statusPageFiles = append(statusPageFiles, fmt.Sprintf("%s/%s", c.folder, file.Name()))
			default:
				return errors.Errorf("Unknown kind: %s", statusKind)
			}
		}
		for _, file := range checkFiles {
			err = c.deleteFile(file)
			if err != nil {
				return err
			}
			log.Infof("File %s applied", file)
		}
		for _, file := range statusPageFiles {
			err = c.deleteFile(file)
			if err != nil {
				return err
			}
			log.Infof("File %s applied", file)
		}
	}
	return nil
}

func (c deleteCmd) deleteFile(filePath string) error {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	statusKind, err := cmdutils.GetFileType(fileBytes)
	if err != nil {
		return err
	}
	switch statusKind {
	case cmdutils.HttpHealthCheck:
		return c.deleteHttpHealthCheck(fileBytes)
	case cmdutils.TLSHealthCheck:
		return c.deleteTLSHealthCheck(fileBytes)
	case cmdutils.StatusPageKind:
		return c.deleteStatusPage(fileBytes)
	default:
		return errors.Errorf("Unknown kind: %s", statusKind)
	}
}

func (c deleteCmd) deleteHttpHealthCheck(fileBytes []byte) error {
	httpCheck := &config.HttpHealthCheck{}
	err := yaml.Unmarshal(fileBytes, httpCheck)
	if err != nil {
		return err
	}
	log.Infof("%+v", httpCheck)
	var m struct {
		DeleteHttpHealthCheck struct {
			ID string `graphql:"id"`
		} `graphql:"deleteCheck(name: $name, namespace: $namespace)"`
	}
	vars := map[string]interface{}{
		"name":      graphql.String(httpCheck.Name),
		"namespace": graphql.String(httpCheck.Namespace),
	}
	ctx := context.Background()
	err = c.gqlClient.Mutate(ctx, &m, vars)
	if err != nil {
		return err
	}
	return nil
}

func (c deleteCmd) deleteTLSHealthCheck(fileBytes []byte) error {
	tlsCheck := &config.TLSHealthCheck{}
	err := yaml.Unmarshal(fileBytes, tlsCheck)
	if err != nil {
		return err
	}
	log.Debugf("%+v", tlsCheck)
	var m struct {
		DeleteHttpHealthCheck struct {
			ID string `graphql:"id"`
		} `graphql:"deleteCheck(name: $name, namespace: $namespace)"`
	}
	vars := map[string]interface{}{
		"name":      tlsCheck.Name,
		"namespace": tlsCheck.Namespace,
	}
	ctx := context.Background()
	err = c.gqlClient.Mutate(ctx, &m, vars)
	if err != nil {
		return err
	}
	return nil
}

func (c deleteCmd) deleteStatusPage(fileBytes []byte) error {
	statusPage := &config.StatusPage{}
	err := yaml.Unmarshal(fileBytes, statusPage)
	if err != nil {
		return err
	}
	log.Debugf("%+v", statusPage)
	slug := fmt.Sprintf("%s-%s", statusPage.Namespace, statusPage.Namespace)
	var m struct {
		DeleteHttpHealthCheck struct {
			ID string `graphql:"id"`
		} `graphql:"deleteCheckBySlug(name: $name, namespace: $namespace)"`
	}
	vars := map[string]interface{}{
		"slug": slug,
	}
	ctx := context.Background()
	err = c.gqlClient.Mutate(ctx, &m, vars)
	if err != nil {
		return err
	}
	return nil
}

func NewDeleteCMD() *cobra.Command {
	c := deleteCmd{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete",
		Long:  `delete`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := c.validate()
			if err != nil {
				return err
			}
			ctx := context.Background()
			gqlClient := cmdutils.GetGraphqlClient(ctx, "http://localhost:8888/graphql")
			c.gqlClient = gqlClient
			return c.run()
		},
	}
	f := cmd.Flags()
	f.StringVarP(&c.file, "file", "f", "", "file to delete")
	f.StringVarP(&c.folder, "folder", "k", "", "folder where the checks are located")
	return cmd
}
