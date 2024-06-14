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
		overlay, name, singName, err := GetOverlay()
		if err != nil {
			fmt.Println(err)
			return
		}
		sif, err := GetSif(name)
		if err != nil {
			fmt.Println(err)
			return
		}

		// create shortcut scripts
		err = WriteSingCmds(singName, name) // , overlay, sif
		if err != nil {
			fmt.Println(err)
			return
		}

		// download and install miniconda (if not done already)
		err = InstallConda(singName)
		if err != nil {
			fmt.Println(err)
			return
		}

		// we're all good! write out shortcuts
		fmt.Printf("\nGreat you're all set!\n\n")
		HowToRun(singName, overlay, sif)

		// provide quick actions to get started
		err = StartSing(singName) // overlay, sif
		fmt.Printf("\nHappy training! :)\n")
		fmt.Printf("\nQuick commands: \033[32m./%s\033[0m (read-only)    \033[32m./%srw\033[0m (read-write) \n", singName, singName)
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

// var (
// 	sifDir     string
// 	overlayDir string
// 	name       string
// 	overlayName string
// 	sifName    string
// )

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.Flags().StringVar(&sifDir, "sifdir", GetEnvVar("SING_SIF_DIR", "/scratch/work/public/singularity"), "Directory where SIF files are located")
	// rootCmd.Flags().StringVar(&overlayDir, "overlaydir", GetEnvVar("SING_OVERLAY_DIR", "/scratch/work/public/overlay-fs-ext3"), "Directory where overlay files are located")
	// rootCmd.Flags().StringVar(&name, "name", "", "Name for the container")
	// rootCmd.Flags().StringVar(&overlayName, "overlay", "", "Name of the overlay file")
	// rootCmd.Flags().StringVar(&sifName, "sif", "", "Name of the SIF file")
	// rootCmd.Flags().StringVar(&singName, "cmd", GetEnvVar("SING_CMD", "sing"), "Name of the SIF file")

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.singuconda.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("quick", "y", false, "Skip steps on subsequent runs.")
}
