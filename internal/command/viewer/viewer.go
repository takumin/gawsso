package viewer

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/takumin/gawsso/internal/awsso"
	"github.com/takumin/gawsso/internal/config"
)

func NewCommands(cfg *config.Config, flags []cli.Flag) *cli.Command {
	flags = append(flags, []cli.Flag{
		&cli.StringFlag{
			Name:        "identity-store-id",
			Aliases:     []string{"id"},
			Usage:       "identity store id",
			EnvVars:     []string{"IDENTITY_STORE_ID"},
			Value:       cfg.IdentityStoreID,
			Destination: &cfg.IdentityStoreID,
		},
	}...)
	return &cli.Command{
		Name:    "viewer",
		Aliases: []string{"view", "v"},
		Usage:   "identity store viewer",
		Flags:   flags,
		Action:  action(cfg),
	}
}

func action(cfg *config.Config) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		store, err := awsso.NewIdentityStore(ctx.Context, cfg.IdentityStoreID)
		if err != nil {
			return err
		}

		if err := store.GetUsers(ctx.Context); err != nil {
			return err
		}

		if err := store.GetGroups(ctx.Context); err != nil {
			return err
		}

		if err := store.GetMembers(ctx.Context); err != nil {
			return err
		}

		data, err := json.MarshalIndent(store, "", "  ")
		if err != nil {
			return err
		}

		_, err = fmt.Println(string(data))
		return err
	}
}
