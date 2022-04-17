package cmd

import (
	"github.com/kfsoftware/statuspage/cmd/apply"
	"github.com/kfsoftware/statuspage/cmd/server"
	"github.com/spf13/cobra"
)

const (
	statusPageDesc = `
statuspage exposes a GraphQL API and monitors services on your behalf
so that you are notified before your customers notice
Detailed help for each command is available with 'statuspage help <command>'.
`
)

func NewCmdStatusPage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "statuspage",
		Short: "monitor services",
		Long:  statusPageDesc,
	}
	cmd.AddCommand(
		server.NewServerCmd(),
		apply.NewApplyCMD(),
	)

	return cmd
}
