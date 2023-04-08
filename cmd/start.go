package cmd

import (
	"github.com/spf13/cobra"
)

type startOptions struct{}

func defaultStartOptions() *startOptions {
	return &startOptions{}
}

func newStartCmd() *cobra.Command {
	o := defaultStartOptions()

	cmd := &cobra.Command{
		Use:          "start",
		Short:        "start subcommand which starts the vector-db",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(0),
		RunE:         o.run,
	}

	return cmd
}

func (o *startOptions) run(cmd *cobra.Command, args []string) error {

	return nil
}
