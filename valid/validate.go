package valid

import (
	"gitlab.myshuju.top/heshiying/admission/admissionhooktool"
	"io"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

// +kubebuilder:webhook:path=/validate,mutating=false,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=pod.valid.myshuju.top,admissionReviewVersions=v1,sideEffects=None

type validateAdmissionWebhook struct {
}

func NewValidateAdmissionWebhook() *validateAdmissionWebhook {
	return &validateAdmissionWebhook{}
}
func (v *validateAdmissionWebhook) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	var body []byte
	if data, err := io.ReadAll(request.Body); err == nil {
		body = data
	}
	requestReview := admissionv1.AdmissionReview{}
	obj, gvk, err := admissionhooktool.Deserializer.Decode(body, nil, &requestReview)
	if err != nil {
		admissionhooktool.Log.Error(err, "deserializer.Decode err")
		response := admissionhooktool.Errored(http.StatusBadRequest, err)
		admissionhooktool.WriteAdmissionResponse(writer, response, requestReview.Request, gvk)
		return
	}
	admissionhooktool.Log.Info("AdmissionReview", "object", obj, "schema.GroupVersionKind", gvk)
	admissionhooktool.Log.Info("AdmissionReviewRaw", "raw", string(requestReview.Request.Object.Raw))
	response := v.doHandle(requestReview)
	admissionhooktool.WriteAdmissionResponse(writer, response, requestReview.Request, gvk)
}
func (v *validateAdmissionWebhook) doHandle(requestReview admissionv1.AdmissionReview) admissionhooktool.Response {
	return admissionhooktool.Response{
		Patches: nil,
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: true,
			Result:  &metav1.Status{},
		},
	}
}
