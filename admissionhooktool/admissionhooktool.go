package admissionhooktool

import (
	"encoding/json"
	"github.com/go-logr/logr"
	"gomodules.xyz/jsonpatch/v2"
	"io"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"net/http"
)

var (
	Deserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
	Log          = logr.Logger{}
)

func PatchResponseFromRaw(request *admissionv1.AdmissionRequest, current []byte) Response {
	var patches []jsonpatch.Operation
	if request != nil {
		original, err := json.Marshal(request.Object.Raw)
		if err != nil {
			return Errored(http.StatusInternalServerError, err)
		}
		patches, err = jsonpatch.CreatePatch(original, current)
		if err != nil {
			return Errored(http.StatusInternalServerError, err)
		}
	}

	return Response{
		Patches: patches,
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: true,
			PatchType: func() *admissionv1.PatchType {
				if len(patches) == 0 {
					return nil
				}
				pt := admissionv1.PatchTypeJSONPatch
				return &pt
			}(),
		},
	}
}

// WriteAdmissionResponse writes ar to w.
func WriteAdmissionResponse(w io.Writer, response Response, req *admissionv1.AdmissionRequest, gvk *schema.GroupVersionKind) {
	ar := admissionv1.AdmissionReview{}
	err := response.CompleteAdmissionResponse(req)
	if err != nil {
		Log.Error(err, "response.CompleteAdmissionResponse err")
		serverErrorResponse := Errored(http.StatusInternalServerError, err)
		ar.Response = &serverErrorResponse.AdmissionResponse
	} else {
		ar.Response = &response.AdmissionResponse
	}

	if gvk == nil || *gvk == (schema.GroupVersionKind{}) {
		ar.SetGroupVersionKind(admissionv1.SchemeGroupVersion.WithKind("AdmissionReview"))
	} else {
		ar.SetGroupVersionKind(*gvk)
	}
	if err := json.NewEncoder(w).Encode(ar); err != nil {
		Log.Error(err, "unable to encode and write the response")
		serverErrorResponse := Errored(http.StatusInternalServerError, err)
		if err = json.NewEncoder(w).Encode(admissionv1.AdmissionReview{Response: &serverErrorResponse.AdmissionResponse}); err != nil {
			Log.Error(err, "still unable to encode and write the InternalServerError response")
		}
	} else {
		res := ar.Response

		if res.Result != nil {
			Log.WithValues("code", res.Result.Code, "reason", res.Result.Reason).Info("wrote response")
		} else {
			Log.V(1).Info("wrote response", "UID", res.UID, "allowed", res.Allowed)
		}

	}
}

type Response struct {
	// Patches are the JSON patches for mutating webhooks.
	// Using this instead of setting Response.Patch to minimize
	// overhead of serialization and deserialization.
	// Patches set here will override any patches in the response,
	// so leave this empty if you want to set the patch response directly.
	Patches []jsonpatch.JsonPatchOperation
	// AdmissionResponse is the raw admission response.
	// The Patch field in it will be overwritten by the listed patches.
	admissionv1.AdmissionResponse
}

func (r *Response) CompleteAdmissionResponse(req *admissionv1.AdmissionRequest) error {
	if req == nil {
		req = &admissionv1.AdmissionRequest{}
	}
	r.UID = req.UID

	// ensure that we have a valid status code
	if r.Result == nil {
		r.Result = &metav1.Status{}
	}
	if r.Result.Code == 0 {
		r.Result.Code = http.StatusOK
	}
	// TODO(directxman12): do we need to populate this further, and/or
	// is code actually necessary (the same webhook doesn't use it)

	if len(r.Patches) == 0 {
		return nil
	}

	var err error
	r.Patch, err = json.Marshal(r.Patches)
	if err != nil {
		return err
	}
	patchType := admissionv1.PatchTypeJSONPatch
	r.PatchType = &patchType

	return nil
}

// Errored creates a new Response for error-handling a request.
func Errored(code int32, err error) Response {
	return Response{
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Code:    code,
				Message: err.Error(),
			},
		},
	}
}
