package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.melnyk.org/cvt"
	"go.melnyk.org/metricli/internal/metricer"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Health information",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var healthcheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check health status",
	Long:  `Check health status`,
	Run: func(cmd *cobra.Command, args []string) {
		cli := metricer.NewClient(flagsHost)

		health, latency, err := cli.HealthCheck()

		fmt.Print("Health status for ", flagsHost, ": ")

		if err != nil {
			fmt.Println(cvt.BrRedFg+"check failed, "+cvt.ResetColor, err)
			return
		}

		if health.Status == "ok" {
			fmt.Print(cvt.BrGreenFg, health.Status, cvt.ResetColor)
		} else {
			fmt.Print(cvt.BrRedFg, health.Status, cvt.ResetColor)
		}
		fmt.Println(" latency:"+cvt.BrWhiteFg, latency, cvt.ResetColor)
	},
}

var healthmonitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor health status",
	Long:  `Monitor health status`,
	Run: func(cmd *cobra.Command, args []string) {
		cli := metricer.NewClient(flagsHost)

		tick := time.Tick(time.Second * 2)

		for {
			select {
			case <-tick:
				health, latency, err := cli.HealthCheck()

				fmt.Print("Health status for ", flagsHost, ": ")

				if err != nil {
					health.Status = "check failed"
					health.Msg = err.Error()
				}

				if health.Status == "ok" {
					fmt.Print(cvt.BrGreenFg, health.Status, cvt.ResetColor)
				} else {
					fmt.Print(cvt.BrRedFg, health.Status, cvt.ResetColor, "  ", health.Msg)
				}
				fmt.Println(" latency:"+cvt.BrWhiteFg, latency, cvt.ResetColor)
			}
		}
	},
}

func init() {
	healthCmd.AddCommand(healthcheckCmd)
	healthCmd.AddCommand(healthmonitorCmd)
	rootCmd.AddCommand(healthCmd)
	healthcheckCmd.Flags().StringVarP(&flagsHost, "host", "H", "localhost:9110", "Merticer host")
}
