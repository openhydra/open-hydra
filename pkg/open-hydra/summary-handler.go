package openhydra

import (
	"context"
	"fmt"
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
	pods, err := builder.k8sHelper.ListPod(builder.Config.OpenHydraNamespace, builder.kubeClient)
	if err != nil {
		writeHttpResponseAndLogError(response, http.StatusInternalServerError, fmt.Sprintf("Error getting pod list: %v", err))
		return
	}
	defRAM := resource.NewQuantity(int64(builder.Config.DefaultRamPerDevice*(1<<20)), resource.BinarySI).String()
	defCPU := resource.NewMilliQuantity(int64(builder.Config.DefaultCpuPerDevice), resource.DecimalSI).String()

	gpuAllocatable := resource.NewQuantity(0, resource.DecimalSI)
	gpuAllocated := resource.NewQuantity(0, resource.DecimalSI)
	podAllocated := 0
	var totalLine uint16
	for _, node := range nodeList.Items {
		gpu := node.Status.Allocatable[coreV1.ResourceName(builder.Config.DefaultGpuDriver)]
		gpuAllocatable.Add(gpu)
	}
	for _, pod := range pods {
		allocateGPU := false
		for _, ctr := range pod.Spec.Containers {
			gpuRequests := ctr.Resources.Requests[coreV1.ResourceName(builder.Config.DefaultGpuDriver)]
			gpuAllocated.Add(gpuRequests)
			if gpuRequests.Value() > 0 {
				allocateGPU = true
			}
		}
		if allocateGPU {
			podAllocated++
		}
		if pod.Status.Phase == coreV1.PodPending {
			totalLine++
		}
	}
	podAllocatable := 0
	if builder.Config.DefaultGpuPerDevice != 0 {
		podAllocatable = int(gpuAllocatable.Value() / int64(builder.Config.DefaultGpuPerDevice))
	}
	sumUp := xSumUpV1.SumUp{
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
			DefaultGpuPerDevice: builder.Config.DefaultGpuPerDevice,
			TotalLine:           totalLine,
		},
	}
	util.FillObjectGVK(&sumUp)
	_ = response.WriteEntity(sumUp)
}
