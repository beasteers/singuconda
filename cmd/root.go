/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "singuconda",
	Short: "Let's build a singularity+anaconda container <3 asdsafjaskd !!!",
	Long: `Let's build a singularity+anaconda container <3 asdsafjaskd !!!

First, make sure you are cd'd in the directory where you want your overlay to live.

Then run:
	~/singuconda

What happens:
	1. pick an overlay file
	2. pick a sif file
	3. install miniconda and optionally a specific python version
	4. adds startup environment script (/ext3/env)
	5. menu to install packages in the container
	6. create shortcut script(s) for running the container

Then you're all done!

You can re-run it if you want to change anything (sif file, python version, installs).

This was built for NYU Greene's environment, but it should apply elsewhere too!`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		// configure overlay/sif
		overlay, name, err := GetOverlay()
		if err != nil {
			fmt.Println(err)
			return
		}
		sif, err := GetSif(name)
		if err != nil {
			fmt.Println(err)
			return
		}

		// download and install miniconda (if not done already)
		err = InstallConda(overlay, sif)
		if err != nil {
			fmt.Println(err)
			return
		}

		// we're all good! write out shortcuts
		fmt.Printf("\nGreat you're all set!\n\n")
		err = WriteSingCmds(overlay, sif)
		if err != nil {
			fmt.Println(err)
			return
		}

		// provide quick actions to get started
		err = StartSing(overlay, sif)
		fmt.Printf("\nHappy training! :)\n")
		fmt.Printf("\nQuick commands: \033[32m./sing\033[0m (read-only)    \033[32m./singrw\033[0m (read-write) \n")
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.singuconda.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("quick", "y", false, "Skip steps on subsequent runs.")
}
