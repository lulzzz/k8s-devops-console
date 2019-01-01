package controllers

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/revel/revel"
	"k8s-devops-console/app"
	"k8s-devops-console/app/services"
	"github.com/dustin/go-humanize"
)

type ResultCluster struct {
	Name string
	Role string
	Version string
	SpecArch string
	SpecOS string
	SpecMachineCPU string
	SpecMachineMemory string
	SpecRegion string
	SpecZone string
	SpecInstance string
	InternalIp string
	Status string
	Created string
	CreatedAgo string
}

type ApiCluster struct {
	ApiBase
}

func (c ApiCluster) accessCheck() (result revel.Result) {
	return c.ApiBase.accessCheck()
}

func (c ApiCluster) Nodes() revel.Result {
	service := services.Kubernetes{}
	nodes, err := service.Nodes()
	if err != nil {
		c.Log.Error("K8s communication error: %v", err)
		return c.renderJSONError("Unable to contact cluster")
	}

	ret := []ResultCluster{}

	for _, node := range nodes.Items {
		row := ResultCluster{
			Name: node.Name,
			Version: node.Status.NodeInfo.KubeletVersion,
			SpecMachineCPU: node.Status.Capacity.Cpu().String(),
			SpecMachineMemory: humanize.Bytes(uint64(node.Status.Capacity.Memory().Value())),
			Status: fmt.Sprintf("%v", node.Status.Phase),
			Created: node.CreationTimestamp.UTC().String(),
			CreatedAgo: revel.TimeAgo(node.CreationTimestamp.UTC()),
		};

		for _, val := range node.Status.Conditions {
			if val.Reason == "KubeletReady" {
				row.Status = fmt.Sprintf("%v", val.Type)
			}
		}

		for _, item := range node.Status.Addresses {
			if item.Type == "InternalIP" {
				row.InternalIp = item.Address
			}
		}

		if val, ok := node.Labels["kubernetes.io/role"]; ok {
			row.Role = val
		}

		if val, ok := node.Labels["beta.kubernetes.io/arch"]; ok {
			row.SpecArch = val
		}

		if val, ok := node.Labels["beta.kubernetes.io/os"]; ok {
			row.SpecOS = val
		}

		if val, ok := node.Labels["failure-domain.beta.kubernetes.io/region"]; ok {
			row.SpecRegion = val
		}

		if val, ok := node.Labels["failure-domain.beta.kubernetes.io/zone"]; ok {
			row.SpecZone = val
		}

		if val, ok := node.Labels["beta.kubernetes.io/instance-type"]; ok {
			row.SpecInstance = val
		}

		ret = append(ret, row)
	}

	app.PrometheusActions.With(prometheus.Labels{"scope": "k8s", "type": "listNodes"}).Inc()

	return c.RenderJSON(ret)
}
