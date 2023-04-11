package transform

import (
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

// ProtoServiceFromService converts model.Service to translatev1.Service.
func ProtoServiceFromService(service *model.Service) *translatev1.Service {
	return &translatev1.Service{Id: service.ID.String(), Name: service.Name}
}
