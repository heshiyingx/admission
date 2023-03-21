package muta

import (
	"regexp"
	"testing"
)

func TestName(t *testing.T) {
	img := "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/resolvers:v0.46.0@sha256:f57448b914c72c03cbf36228134cc9ed24e28fef6d2e0d6d72c34908f38d8742"
	//img = "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/resolvers:v0.46.0"
	var reg = regexp.MustCompile(`(.?)@sha256:.*`)
	allString := reg.ReplaceAllString(img, `$1`)
	t.Log(allString)
}
