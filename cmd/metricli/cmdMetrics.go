package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.melnyk.org/cvt"
	"go.melnyk.org/metricli/internal/metricer"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Metrics information",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var metricsshowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show metric values",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cli := metricer.NewClient(flagsHost)
		metrics, latency, err := cli.GetMetrics()

		fmt.Println(" Metrics for", flagsHost, ": ")
		fmt.Println(" Latency:"+cvt.BrWhiteFg, latency, cvt.ResetColor)

		if err != nil {
			fmt.Println(cvt.BrRedFg+"Failed"+cvt.ResetColor, err)
			return
		}

		var uptime time.Duration
		if v, ok := metrics["uptime"]; ok {
			if val, ok := v.(json.Number); ok {
				if i, err := val.Int64(); err == nil {
					uptime = time.Duration(i).Round(time.Second)
				}
			}
		}
		fmt.Println(" Uptime:", uptime)

		var submetrics map[string]interface{}
		if v, ok := metrics["metrics"]; ok {
			if val, ok := v.(map[string]interface{}); ok {
				submetrics = val
			}
		}

		fmt.Print("   \033H                       \033H   \033H   \033H   \033H   \033H", cvt.MoveBegin)

		for k, v := range metrics {
			if k != "metrics" {
				fmt.Println(cvt.Tab+cvt.Bold, k, cvt.ResetColor+cvt.Tab, ":", v)
			}
		}

		fmt.Print("\033[3g        \033H              \033H   \033H   \033H   \033H   \033H", cvt.MoveBegin)

		for k, v := range submetrics {
			fmt.Println(cvt.Tab, k, cvt.Tab, ":", v)
		}
	},
}

var metricrateCmd = &cobra.Command{
	Use:   "rate <metric_name>*",
	Short: "Rate metric value",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		cli := metricer.NewClient(flagsHost)
		metrics, latency, err := cli.GetMetrics()

		if len(args) == 0 {
			fmt.Println(cvt.BrRedFg + "Metric name is required" + cvt.ResetColor)
			return
		}

		fmt.Println(" Rate metrics for", flagsHost, ": ")

		if err != nil {
			fmt.Println(cvt.BrRedFg+"Failed"+cvt.ResetColor, err)
			return
		}

		err = errors.New("Metrics section does not exist")
		if v, ok := metrics["metrics"]; ok {
			if subm, ok := v.(map[string]interface{}); ok {
				for _, metricname := range args {
					err = errors.New("Required metric " + metricname + " does not exist")
					if v, ok := subm[metricname]; ok {
						if val, ok := v.(json.Number); ok {
							if _, erri := val.Int64(); erri == nil {
								err = nil
								break
							}
						}
					}
				}
			}
		}
		if err != nil {
			fmt.Println(cvt.BrRedFg, err, cvt.ResetColor)
			return
		}

		fmt.Print(cvt.BrWhiteFg + "\033[3g  \033HUptime       \033HLatency")
		for _ = range args {
			fmt.Print("        \033HValue        \033HRate")
		}
		fmt.Println("     \033H   \033H   \033H")
		fmt.Print(cvt.Tab, cvt.Tab)
		for _, metricname := range args {
			fmt.Print(cvt.Tab, metricname, cvt.Tab)
		}
		fmt.Println(cvt.ResetColor)

		tick := time.Tick(time.Second * 2)

		var (
			previousCheck time.Time
			previousValue = make(map[string]int64)
			previousRate  = make(map[string]int64)
		)

		for {
			select {
			case <-tick:
				value := make(map[string]int64)
				metrics, latency, err = cli.GetMetrics()
				check := time.Now()

				if err != nil {
					fmt.Println(cvt.Tab, cvt.BrRedFg+"Failed"+cvt.ResetColor)
					continue
				}

				var uptime time.Duration
				if v, ok := metrics["uptime"]; ok {
					if val, ok := v.(json.Number); ok {
						if i, err := val.Int64(); err == nil {
							uptime = time.Duration(i).Round(time.Second)
						}
					}
				}

				var submetrics map[string]interface{}
				if v, ok := metrics["metrics"]; ok {
					if val, ok := v.(map[string]interface{}); ok {
						submetrics = val
					}
				}

				for _, metricname := range args {
					if v, ok := submetrics[metricname]; ok {
						if val, ok := v.(json.Number); ok {
							if i, err := val.Int64(); err == nil {
								value[metricname] = i
							}
						}
					}
				}

				if previousCheck.Unix() != 0 {
					fmt.Print(cvt.Tab, uptime, cvt.Tab, latency)

					rate := make(map[string]int64)
					for _, metricname := range args {
						delta := (value[metricname] - previousValue[metricname]) * 1000
						duration := int64(check.Sub(previousCheck).Round(time.Millisecond) / time.Millisecond)
						if duration == 0 {
							duration = 1
						}
						rate[metricname] = delta / duration

						sum := " "
						if 98*rate[metricname] > 100*previousRate[metricname] {
							sum = cvt.GreenFg + cvt.ArrowUp + cvt.ResetColor
						}
						if 100*rate[metricname] < 98*previousRate[metricname] {
							sum = cvt.RedFg + cvt.ArrowDown + cvt.ResetColor
						}

						fmt.Print(cvt.Tab, value[metricname], cvt.Tab, rate[metricname], sum)
					}

					fmt.Println()
					previousRate = rate
				}

				previousCheck = check
				previousValue = value
			}
		}
	},
}

func init() {
	metricsCmd.AddCommand(metricsshowCmd)
	metricsCmd.AddCommand(metricrateCmd)
	rootCmd.AddCommand(metricsCmd)
	metricsshowCmd.Flags().StringVarP(&flagsHost, "host", "H", "localhost:9110", "Metricer host")
	metricrateCmd.Flags().StringVarP(&flagsHost, "host", "H", "localhost:9110", "Metricer host")
}
