package adfunc

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
	"strings"
)

func init() {
	register(AdmissionFunc{
		Type: AdmissionTypeMutating,
		Path: "/oversold",
		Func: oversold,
	})
}

func oversold(request *admissionv1.AdmissionRequest) (*admissionv1.AdmissionResponse, error) {
	//获取属性Kind为Node
	switch request.Kind.Kind {
	case "Node":
		node := v1.Node{}
		if err := jsoniter.Unmarshal(request.Object.Raw, &node); err != nil {
			errMsg := fmt.Sprintf("[route.Mutating] /oversold: failed to unmarshal object: %v", err)
			klog.Error(errMsg)
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Code:    http.StatusBadRequest,
					Message: errMsg,
				},
			}, nil
		}
		if node.Labels["kubernetes.io/oversold"] != "oversold" {

			return &admissionv1.AdmissionResponse{
				Allowed:   true,
				PatchType: JSONPatch(),
				Result: &metav1.Status{
					Code:    http.StatusOK,
					Message: "节点无需超售",
				},
			}, nil
		}

		klog.Info(request.UserInfo.Username + "的label为kubernetes.io/oversold:" + node.Labels["kubernetes.io/oversold"])
		klog.Info(request.UserInfo.Username + "===================该节点允许超售========================")
		patches := []Patch{
			{
				Option: PatchOptionReplace,
				Path:   "/status/allocatable/cpu",
				Value:  overcpu(Quantitytostring(node.Status.Capacity.Cpu()), node.Labels["kubernetes.io/overcpu"]),
			},
			{
				Option: PatchOptionReplace,
				Path:   "/status/allocatable/memory",
				Value:  overmem(Quantitytostring(node.Status.Allocatable.Memory()), node.Labels["kubernetes.io/overmem"]),
			},
			{
				Option: PatchOptionReplace,
				Path:   "/status/allocatable/pods",
				Value:  overpods(Quantitytostring(node.Status.Capacity.Pods()), node.Labels["kubernetes.io/overpods"]),
			},
		}
		patch, err := jsoniter.Marshal(patches)
		if err != nil {
			errMsg := fmt.Sprintf("[route.Mutating] /oversold: failed to marshal patch: %v", err)
			logger.Error(errMsg)
			return &admissionv1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Code:    http.StatusInternalServerError,
					Message: errMsg,
				},
			}, nil
		}
		logger.Infof("[route.Mutating] /oversold: patches: %s", string(patch))
		return &admissionv1.AdmissionResponse{
			Allowed:   true,
			Patch:     patch,
			PatchType: JSONPatch(),
			Result: &metav1.Status{
				Code:    http.StatusOK,
				Message: "success",
			},
		}, nil

	default:
		errMsg := fmt.Sprintf("[route.Mutating] /oversold: received wrong kind request: %s, Only support Kind: Deployment", request.Kind.Kind)
		logger.Error(errMsg)
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Code:    http.StatusForbidden,
				Message: errMsg,
			},
		}, nil
	}

}

//*resource.Quantity类型转string
func Quantitytostring(r *resource.Quantity) string {
	return fmt.Sprint(r)
}

//cpu 超售计算
func overcpu(cpu, multiple string) string {
	a, _ := strconv.ParseFloat(cpu, 32)
	if multiple == "" {
		multiple = "1"
	}
	b, _ := strconv.ParseFloat(multiple, 32)
	c := int(a * b)
	return strconv.Itoa(c)
}

// mem 超售计算
func overmem(mem, multiple string) string {
	a, _ := strconv.ParseFloat(strings.Trim(mem, "Ki"), 32)
	if multiple == "" {
		multiple = "1"
	}
	b, _ := strconv.ParseFloat(multiple, 32)
	c := int(a * b)
	return strconv.Itoa(c) + "Ki"
}

//pods 超售计算
func overpods(pods, multiple string) string {
	a, _ := strconv.ParseFloat(pods, 32)
	if multiple == "" {
		multiple = "1"
	}
	b, _ := strconv.ParseFloat(multiple, 32)
	c := int(a * b)
	return strconv.Itoa(c)
}
