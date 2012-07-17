package main

import (
	"encoding/json"
	"github.com/timeredbull/tsuru/cmd"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type Service struct{}
type ServiceCreate struct{}

func (c *Service) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "service",
		Usage:   "service (init|list|create|remove|update) [args]",
		Desc:    "manage services.",
		MinArgs: 1,
	}
}

func (c *Service) Subcommands() map[string]interface{} {
	return map[string]interface{}{
		"create": &ServiceCreate{},
	}
}

func (c *ServiceCreate) Info() *cmd.Info {
	desc := "Creates a service based on a passed manifest. The manifest format should be a yaml and follow the standard described in the documentation (should link to it here)"
	return &cmd.Info{
		Name:    "create",
		Usage:   "create path/to/manifesto",
		Desc:    desc,
		MinArgs: 1,
	}
}

func (c *ServiceCreate) Run(context *cmd.Context, client cmd.Doer) error {
	manifest := context.Args[0]
	url := cmd.GetUrl("/services")
	b, err := ioutil.ReadFile(manifest)
	if err != nil {
		return err
	}
	body := strings.NewReader(string(b))
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	r, err := client.Do(request)
	if err != nil {
		return err
	}
	b, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	io.WriteString(context.Stdout, string(b)+"\n")
	return nil
}

type ServiceRemove struct{}

func (c *ServiceRemove) Run(context *cmd.Context, client cmd.Doer) error {
	serviceName := context.Args[0]
	url := cmd.GetUrl("/services/" + serviceName)
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	_, err = client.Do(request)
	if err != nil {
		return err
	}
	io.WriteString(context.Stdout, "Service successfully removed.\n")
	return nil
}

func (c *ServiceRemove) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "remove",
		Usage:   "remove <servicename>",
		Desc:    "removes a service from catalog",
		MinArgs: 1,
	}
}

type ServiceList struct{}

func (c *ServiceList) Info() *cmd.Info {
	return &cmd.Info{
		Name:  "list",
		Usage: "list",
		Desc:  "list services that belongs to user's team and it's service instances.",
	}
}

type ServiceModel struct {
	Service   string
	Instances []string
}

func (c *ServiceList) Run(ctxt *cmd.Context, client cmd.Doer) error {
	url := cmd.GetUrl("/services")
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(request)
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	rslt, err := c.show(b)
	if err != nil {
		return err
	}
	ctxt.Stdout.Write(rslt)
	return nil
}

func (c *ServiceList) show(b []byte) ([]byte, error) {
	var services []ServiceModel
	err := json.Unmarshal(b, &services)
	if err != nil {
		return []byte{}, err
	}
	table := cmd.NewTable()
	table.Headers = cmd.Row([]string{"Services", "Instances"})
	for _, s := range services {
		insts := strings.Join(s.Instances, ", ")
		r := cmd.Row([]string{s.Service, insts})
		table.AddRow(r)
	}
	return table.Bytes(), nil
}
