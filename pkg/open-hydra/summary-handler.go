package openhydra

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xSumUpV1 "open-hydra/pkg/apis/open-hydra-api/summary/core/v1"
	"open-hydra/pkg/util"

	"github.com/emicklei/go-restful/v3"
)

func (builder *OpenHydraRouteBuilder) AddSummaryGetRoute() {
	path := "/" + SumUpPath
	builder.addPathAuthorization(path, http.MethodGet, 3)
	builder.RootWS.Route(builder.RootWS.GET(path).Operation("getSummary").To(builder.SummaryGetRouteHandler).
		Returns(http.StatusInternalServerError, "internal server error", "").
		Returns(http.StatusForbidden, "forbidden", "").
		Returns(http.StatusUnauthorized, "unauthorized", "").
		Returns(http.StatusOK, "OK", xSumUpV1.SumUp{}))
}

func (builder *OpenHydraRouteBuilder) SummaryGetRouteHandler(request *restful.Request, response *restful.Response) {
	nodeList, err := builder.kubeClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Error getting node list: %v", err))
		return
	}
	// TODO may be all namespace
	pods, err := builder.k8sHelper.ListPod(OpenhydraNamespace, builder.kubeClient)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Error getting pod list: %v", err))
		return
	}

	sumUp, err := builder.SumUpGpuResources(pods, nodeList)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Error sum up gpu resources: %v", err))
		return
	}

	util.FillObjectGVK(sumUp)
	_ = response.WriteEntity(sumUp)
}

func (builder *OpenHydraRouteBuilder) SumUpGpuResources(pods []coreV1.Pod, nodeList *coreV1.NodeList) (*xSumUpV1.SumUp, error) {

	serverConfig, err := builder.GetServerConfigFromConfigMap()
	if err != nil {
		return nil, err
	}

	defRAM := resource.NewQuantity(int64(serverConfig.DefaultRamPerDevice*(1<<20)), resource.BinarySI).String()
	defCPU := resource.NewMilliQuantity(int64(serverConfig.DefaultCpuPerDevice), resource.DecimalSI).String()

	gpuAllocatable := resource.NewQuantity(0, resource.DecimalSI)
	gpuAllocated := resource.NewQuantity(0, resource.DecimalSI)
	podAllocated := 0
	gpuResourceCount := map[string]struct {
		Allocatable *resource.Quantity
		Allocated   *resource.Quantity
	}{}
	var totalLine uint16
	if len(serverConfig.GpuResourceKeys) == 0 {
		// warn
		slog.Warn("gpu resource key is empty, so total gpu number will be 0 and all device use gpu will fall back to default")
	} else {

		for _, node := range nodeList.Items {
			for _, gpuResourceKey := range serverConfig.GpuResourceKeys {
				gpu := node.Status.Allocatable[coreV1.ResourceName(gpuResourceKey)]
				if _, found := gpuResourceCount[gpuResourceKey]; found {
					gpuResourceCount[gpuResourceKey].Allocatable.Add(gpu)
				} else {
					gpuResourceCount[gpuResourceKey] = struct {
						Allocatable *resource.Quantity
						Allocated   *resource.Quantity
					}{
						Allocatable: &gpu,
						Allocated:   resource.NewQuantity(0, resource.DecimalSI),
					}
				}
			}
		}
		for _, pod := range pods {
			for _, ctr := range pod.Spec.Containers {
				for _, gpuResourceKey := range serverConfig.GpuResourceKeys {
					gpuRequests := ctr.Resources.Requests[coreV1.ResourceName(gpuResourceKey)]
					gpuResourceCount[gpuResourceKey].Allocated.Add(gpuRequests)
				}
			}
			if pod.Status.Phase == coreV1.PodPending {
				totalLine++
			}
		}
	}

	// sum up all gpu device allocation status
	gpuDriverSumUp := map[string]xSumUpV1.GpuResourceSumUp{}
	for key, allocationStatus := range gpuResourceCount {
		gpuAllocatable.Add(*allocationStatus.Allocatable)
		gpuAllocated.Add(*allocationStatus.Allocated)
		gpuDriverSumUp[key] = xSumUpV1.GpuResourceSumUp{
			Allocatable: allocationStatus.Allocatable.Value(),
			Allocated:   allocationStatus.Allocated.Value(),
		}
	}

	podAllocatable := 0
	if serverConfig.DefaultGpuPerDevice != 0 {
		podAllocatable = int(gpuAllocatable.Value() / int64(serverConfig.DefaultGpuPerDevice))
	}

	return &xSumUpV1.SumUp{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: metav1.Now(),
		},
		Spec: xSumUpV1.SumUpSpec{
			PodAllocatable:      podAllocatable,
			PodAllocated:        podAllocated,
			GpuAllocatable:      gpuAllocatable.String(),
			GpuAllocated:        gpuAllocated.String(),
			DefaultCpuPerDevice: defCPU,
			DefaultRamPerDevice: defRAM,
			DefaultGpuPerDevice: serverConfig.DefaultGpuPerDevice,
			TotalLine:           totalLine,
			GpuResourceSumUp:    gpuDriverSumUp,
		},
	}, nil
}
