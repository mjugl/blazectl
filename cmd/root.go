// Copyright © 2019 Alexander Kiel <alexander.kiel@life.uni-leipzig.de>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"errors"
	"fmt"
	"github.com/life-research/blazectl/fhir"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

var cfgFile string
var server string
var ctx string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "blazectl",
	Short: "Control your FHIR® Server from the Command Line",
	Long: `blazectl controls FHIR® servers.

Currently you can upload transaction bundles from a directory and count resources.`,
	Version: "0.2.1",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.blaze/config)")
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "", "the base URL of the server to use")
	rootCmd.PersistentFlags().StringVar(&ctx, "context", "", "the name of the config context to use")
	err := viper.BindPFlag("current-context", rootCmd.PersistentFlags().Lookup("context"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(filepath.Join(home, ".blaze"))
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func initClient() (*fhir.Client, error) {
	if server != "" {
		return &fhir.Client{Base: server}, nil
	}
	if context := viper.GetString("current-context"); context != "" {
		if server := viper.GetString("contexts." + context + ".server"); server != "" {
			if base := viper.GetString("servers." + server + ".base"); base != "" {
				return &fhir.Client{Base: base}, nil
			}
		}
	}
	return nil, errors.New("no server configured")
}
