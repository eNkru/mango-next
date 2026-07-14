package main

import (
	"fmt"
	"os"

	"github.com/eNkru/mango-next/internal/config"
	"github.com/eNkru/mango-next/internal/storage"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// newAdminCmd builds `mango admin ...` mirroring the Crystal clim subcommands.
func newAdminCmd(configPath *string) *cobra.Command {
	admin := &cobra.Command{
		Use:   "admin",
		Short: "Run admin tools",
	}

	admin.AddCommand(newAdminUserCmd(configPath))
	return admin
}

// newAdminUserCmd builds `mango admin user <action> ...`
func newAdminUserCmd(configPath *string) *cobra.Command {
	var (
		username string
		password string
		isAdmin  bool
	)

	userCmd := &cobra.Command{
		Use:   "user [action] [username]",
		Short: "User management tool",
		Long:  "Action to perform. Can be add/delete/update/list",
		Args:  cobra.MaximumNArgs(2),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// PreRun avoids repeated open/close for subcommands.
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			action := ""
			if len(args) > 0 {
				action = args[0]
			}
			switch action {
			case "add", "delete", "update", "list":
				return fmt.Errorf("use the subcommand: mango admin user %s", action)
			case "":
				return cmd.Help()
			default:
				return fmt.Errorf("unknown action %q", action)
			}
		},
	}

	// `mango admin user list`
	userCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all users",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := openStorage(configPath)
			if err != nil {
				return err
			}
			defer st.Close()

			users, err := st.ListUsers()
			if err != nil {
				return fmt.Errorf("list users: %w", err)
			}

			table := tablewriter.NewTable(os.Stdout,
				tablewriter.WithHeader([]string{"Username", "Admin"}),
			)

			for _, u := range users {
				adminStr := "no"
				if u.IsAdmin {
					adminStr = "yes"
				}
				if err := table.Append(u.Username, adminStr); err != nil {
					return err
				}
			}
			if err := table.Render(); err != nil {
				return err
			}
			return nil
		},
	})

	// `mango admin user add -u username -p password [-a]`
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			if username == "" {
				return fmt.Errorf("username is required (use -u/--username)")
			}
			if password == "" {
				return fmt.Errorf("password is required (use -p/--password)")
			}

			st, err := openStorage(configPath)
			if err != nil {
				return err
			}
			defer st.Close()

			if err := st.NewUser(username, password, isAdmin); err != nil {
				return fmt.Errorf("add user: %w", err)
			}
			fmt.Printf("User %q created.\n", username)
			return nil
		},
	}
	addCmd.Flags().StringVarP(&username, "username", "u", "", "Username")
	addCmd.Flags().StringVarP(&password, "password", "p", "", "Password")
	addCmd.Flags().BoolVarP(&isAdmin, "admin", "a", false, "Admin flag")
	userCmd.AddCommand(addCmd)

	// `mango admin user delete <username>`
	delCmd := &cobra.Command{
		Use:   "delete [username]",
		Short: "Delete a user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := openStorage(configPath)
			if err != nil {
				return err
			}
			defer st.Close()

			if err := st.DeleteUser(args[0]); err != nil {
				return fmt.Errorf("delete user: %w", err)
			}
			fmt.Printf("User %q deleted.\n", args[0])
			return nil
		},
	}
	userCmd.AddCommand(delCmd)

	// `mango admin user update [-u username] [-p password] [-a] <target_username>`
	updateCmd := &cobra.Command{
		Use:   "update [username_to_update]",
		Short: "Update a user",
		Long:  "Update a user's username, password, and/or admin status. Specify the target user as the argument.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			original := args[0]
			newUsername := username
			if newUsername == "" {
				newUsername = original
			}

			st, err := openStorage(configPath)
			if err != nil {
				return err
			}
			defer st.Close()

			if err := st.UpdateUser(original, newUsername, password, isAdmin); err != nil {
				return fmt.Errorf("update user: %w", err)
			}
			fmt.Printf("User %q updated.\n", original)
			return nil
		},
	}
	updateCmd.Flags().StringVarP(&username, "username", "u", "", "New username")
	updateCmd.Flags().StringVarP(&password, "password", "p", "", "New password")
	updateCmd.Flags().BoolVarP(&isAdmin, "admin", "a", false, "Admin flag")
	userCmd.AddCommand(updateCmd)

	return userCmd
}

// openStorage is a helper to load config and open the database.
func openStorage(configPath *string) (*storage.Storage, error) {
	cfg, err := config.Load(*configPath)
	if err != nil {
		return nil, err
	}
	cfg.SetCurrent()

	st, err := storage.Open(cfg.DBPath, cfg.LibraryPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	return st, nil
}
