package muta

import (
	"encoding/json"
	"gitlab.myshuju.top/heshiying/admission/admissionhooktool"
	"io"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// +kubebuilder:webhook:path=/mutate,mutating=true,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=pod.mutate.myshuju.top,admissionReviewVersions=v1,sideEffects=None

const MODIFY_IMG_PRE = "MODIFY_IMG_PRE"
const MODIFY_IMG_DEFAULT = "MODIFY_IMG_DEFAULT"

var log = klog.NewKlogr().WithValues("mutatingAdmissionWebhook", "")
var reg = regexp.MustCompile(`(.?)@sha256:.*`)
var (
	imageModifyList        = make([]string, 0, 6)
	defaultImageModifyList = []string{"registry.k8s.io", "k8s.gcr.io", "gcr.io"}
)

type mutatingAdmissionWebhook struct {
}

func NewMutatingAdmissionWebhook() *mutatingAdmissionWebhook {
	refreshImageModifyList()
	return &mutatingAdmissionWebhook{}
}

func (h *mutatingAdmissionWebhook) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var body []byte
	if data, err := io.ReadAll(request.Body); err == nil {
		body = data
	}
	requestReview := admissionv1.AdmissionReview{}
	_, gvk, err := admissionhooktool.Deserializer.Decode(body, nil, &requestReview)
	if err != nil {
		admissionhooktool.Log.Error(err, "deserializer.Decode err")
		response := admissionhooktool.Errored(http.StatusBadRequest, err)
		admissionhooktool.WriteAdmissionResponse(writer, response, requestReview.Request, gvk)
		return
	}
	admissionResponse := doHandle(requestReview)
	admissionhooktool.WriteAdmissionResponse(writer, admissionResponse, requestReview.Request, gvk)

}
func doHandle(request admissionv1.AdmissionReview) admissionhooktool.Response {

	if len(imageModifyList) == 0 {
		return admissionhooktool.Response{
			Patches: nil,
			AdmissionResponse: admissionv1.AdmissionResponse{
				Allowed:   true,
				PatchType: nil,
			},
		}
	}

	pod := corev1.Pod{}
	if err := json.Unmarshal(request.Request.Object.Raw, &pod); err != nil {
		admissionhooktool.Log.Error(err, "json.Unmarshal(requestReview.Request.Object.Raw, &pod) err")
		return admissionhooktool.Errored(http.StatusBadRequest, err)
	}

	for i, container := range pod.Spec.InitContainers {
		oldImg := container.Image
		for _, prefix := range imageModifyList {

			if strings.HasPrefix(container.Image, prefix) {
				image := reg.ReplaceAllString(container.Image, `$1`)
				pod.Spec.InitContainers[i].Image = "harbor.myshuju.top/" + image
			}
		}
		log.Info("image url", "old url", oldImg, "new url", pod.Spec.InitContainers[i].Image, "pull", "docker pull "+oldImg, "tag", "docker tag "+oldImg+"  "+pod.Spec.InitContainers[i].Image)
	}

	for i, container := range pod.Spec.Containers {
		oldImg := container.Image
		for _, prefix := range imageModifyList {

			if strings.HasPrefix(container.Image, prefix) {
				image := reg.ReplaceAllString(container.Image, `$1`)
				pod.Spec.Containers[i].Image = "harbor.myshuju.top/" + image
			}
		}
		log.Info("image url", "old url", oldImg, "new url", pod.Spec.Containers[i].Image, "pull", "docker pull "+oldImg, "tag", "docker tag "+oldImg+"  "+pod.Spec.Containers[i].Image)

	}
	nowPodBytes, err := json.Marshal(&pod)
	if err != nil {
		log.Error(err, "Marshal pod Err")
		return admissionhooktool.Errored(http.StatusInternalServerError, err)
	}
	return admissionhooktool.PatchResponseFromRaw(request.Request, nowPodBytes)
}
func refreshImageModifyList() {
	useDefaultModifyImageKey := os.Getenv(MODIFY_IMG_DEFAULT)
	defaultKey := strings.ToLower(useDefaultModifyImageKey)
	if defaultKey == "true" || defaultKey == "1" || defaultKey == "yes" {
		imageModifyList = defaultImageModifyList
	} else {
		imageModifyList = imageModifyList[0:0]
		imagePres := strings.Split(MODIFY_IMG_PRE, ",")
		if len(imagePres) > 0 {
			imageModifyList = append(imageModifyList, imagePres...)
		}
	}
	log.Info("imageModifyList:", "imageModifyList", strings.Join(imageModifyList, ","))

}
