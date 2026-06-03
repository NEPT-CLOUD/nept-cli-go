package cmd

import (
	"errors"

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/spf13/cobra"
)

func Login (appContainer *app.App) *cobra.Command{
	var key string
	
	cmd := &cobra.Command{
		Use: "login",
		Short: "login to nept",
		Long: "nept is a scalable command-line tool" ,
		RunE: func(cmd *cobra.Command, args []string) error {
			if key == ""{
				return app.NewAppError("KEY_MISSING", errors.New("API_KEY is required"))
			}
			else {

			}
			return  appContainer.PrintResult("sex" , map[string]string{"name": "fuck"})
		},
		
	}

	// cmd.Flags().StringVarP(&key , "key" , "k","samir" , "the key of owner ")
	cmd.Flags().StringVarP(&key , "key" , "k","", "collect the key from https://nept-cloud.io/dashboard/apikeys")

	return cmd
}



func checkValidKey(key string) error {
	
}
	
