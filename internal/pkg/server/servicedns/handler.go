package servicedns

import (
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-network-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/network-manager/internal/pkg/entities"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

type Handler struct {
	Manager Manager
}

func NewHandler(manager Manager) *Handler {
	return &Handler{manager}
}

func (h*Handler) AddEntry(ctx context.Context, request *grpc_network_go.AddServiceDNSEntryRequest) (*grpc_common_go.Success, error) {
	vErr := entities.ValidAddServiceDNSEntryRequest(request)
	if vErr != nil{
		return nil, conversions.ToGRPCError(vErr)
	}
	log.Debug().Str("organization_id", request.OrganizationId).Str("fqdn", request.Fqdn).Str("ip", request.Ip).Msg("Add service DNS entry")
	err := h.Manager.AddEntry(request)
	if err != nil{
		return nil, conversions.ToGRPCError(err)
	}
	return &grpc_common_go.Success{}, nil
}

func (h*Handler) DeleteEntry(ctx context.Context, request *grpc_network_go.DeleteServiceDNSEntryRequest) (*grpc_common_go.Success, error) {
	vErr := entities.ValidDeleteServiceDNSEntryRequest(request)
	if vErr != nil{
		return nil, conversions.ToGRPCError(vErr)
	}
	log.Debug().Str("organization_id", request.OrganizationId).Str("fqdn", request.Fqdn).Msg("Remove service DNS entry")
	err := h.Manager.DeleteEntry(request)
	if err != nil{
		return nil, conversions.ToGRPCError(err)
	}
	return &grpc_common_go.Success{}, nil
}

func (h*Handler) ListEntries(ctx context.Context, request *grpc_organization_go.OrganizationId) (*grpc_network_go.ServiceDNSEntryList, error) {
	vErr := entities.ValidOrganizationId(request)
	if vErr != nil{
		return nil, conversions.ToGRPCError(vErr)
	}
	list, err := h.Manager.ListEntries(request)
	if err != nil{
		return nil, conversions.ToGRPCError(err)
	}
	return &grpc_network_go.ServiceDNSEntryList{
		Entries:              list,
	}, nil
}

