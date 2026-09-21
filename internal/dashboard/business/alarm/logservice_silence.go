package alarm

import (
	"github.com/oceanbase/ob-operator/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
	"strings"
)

func logServiceSilenceIdentity(value string) (string, string, error) {
	ns, name, ok := strings.Cut(value, "/")
	if !ok || len(validation.IsDNS1123Label(ns)) > 0 || len(validation.IsDNS1123Subdomain(name)) > 0 {
		return "", "", errors.NewBadRequest("LogService identity must be namespace/name")
	}
	return ns, name, nil
}
