package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type pod struct {
	Name              string              `json:"name"`
	Conditions        []condition         `json:"conditions"`
	ContainerStatuses []containerStatuses `json:"containerStatuses"`
}

type condition struct {
	LastProbeTime      string    `json:"lastProbeTime"`
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Status             string    `json:"status"`
	Type               string    `json:"type"`
}

type containerStatuses struct {
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	RestartCount int    `json:"restartCount"`
}

type startup struct {
	Pod      string
	Init     time.Time
	Ready    time.Time
	Duration time.Duration
	Restarts int
}

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stdout, "Failed due to: %s\n", err)
		os.Exit(1)
	}
}

func run(in io.Reader, _ io.Writer) error {
	var pods []pod
	err := json.NewDecoder(in).Decode(&pods)
	if err != nil {
		return err
	}

	var ups []startup
	for _, p := range pods {
		var init time.Time
		var ready time.Time
		for _, c := range p.Conditions {
			if c.Type == "Initialized" && c.Status == "True" {
				init = c.LastTransitionTime
			} else if c.Type == "Ready" && c.Status == "True" {
				ready = c.LastTransitionTime
			}
		}
		var restarts int
		for _, s := range p.ContainerStatuses {
			if s.Name == "core" && s.Ready {
				restarts = s.RestartCount
			}
		}
		up := startup{
			Pod:      p.Name,
			Init:     init,
			Ready:    ready,
			Duration: ready.Sub(init),
			Restarts: restarts,
		}
		ups = append(ups, up)
	}

	for _, up := range ups {
		fmt.Printf("%q: %v (ready) - %v (init) = %v (duration) [%d (restarts)]\n", up.Pod, up.Ready, up.Init, up.Duration, up.Restarts)
	}

	return nil
}
