package cmd

import (

	"github.com/NEPT-CLOUD/nept-cli-go/internal/app"
	"github.com/spf13/cobra"
)


func Auther(appContainer *app.App) *cobra.Command {
	// var projectName string
	// var bool bool
	var name string
	

	cmd := &cobra.Command{
		Use: "auther",
		Short: "about nept",
		Long: "nept is a scalable command-line tool" ,
		RunE: func(cmd *cobra.Command, args []string) error {

			return  appContainer.PrintResult("hard-sex" , map[string]string{"name": name})
		},
		
	}

	cmd.Flags().StringVarP(&name , "name" , "n","samir" , "the name of owner ")



	return  cmd



}